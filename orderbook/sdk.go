package orderbook

import (
	"errors"
	"fmt"
	"sync"

	"github.com/shopspring/decimal"
)

var (
	ErrNotInitialized = errors.New("orderbook not initialized")
	ErrTokenNotFound  = errors.New("token not found")
	ErrNoData         = errors.New("no data available")
	ErrAlreadyStarted = errors.New("sdk already started")
)

// SDK 订单簿SDK对外接口
type SDK struct {
	mu      sync.RWMutex
	manager *Manager
	config  *Config
	started bool
}

// NewSDK 创建新的SDK实例
func NewSDK(config *Config) *SDK {
	if config == nil {
		config = DefaultConfig()
	}

	return &SDK{
		config: config,
	}
}

// Subscribe 订阅token列表
func (s *SDK) Subscribe(tokenIDs []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return ErrAlreadyStarted
	}

	if len(tokenIDs) == 0 {
		return errors.New("tokenIDs cannot be empty")
	}

	s.manager = NewManager(s.config)
	if err := s.manager.Subscribe(tokenIDs); err != nil {
		s.manager = nil
		return err
	}

	s.started = true
	return nil
}

// Updates 获取更新通知channel
func (s *SDK) Updates() <-chan OrderBookUpdate {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.manager == nil {
		return nil
	}
	return s.manager.Updates()
}

// Close 关闭SDK
func (s *SDK) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.manager != nil {
		s.manager.Close()
		s.manager = nil
	}
	s.started = false
}

// IsInitialized 检查指定token的订单簿是否已初始化
func (s *SDK) IsInitialized(tokenID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.manager == nil {
		return false
	}
	return s.manager.IsInitialized(tokenID)
}

// IsAllInitialized 检查所有订单簿是否都已初始化
func (s *SDK) IsAllInitialized() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.manager == nil {
		return false
	}
	return s.manager.IsAllInitialized()
}

// GetConnectionStatus 获取连接状态
func (s *SDK) GetConnectionStatus() map[string]ConnectionState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.manager == nil {
		return nil
	}
	return s.manager.GetConnectionStatus()
}

// getOrderBook 获取订单簿（内部方法，调用者需持有读锁）
func (s *SDK) getOrderBookLocked(tokenID string) (*OrderBook, error) {
	if s.manager == nil {
		return nil, errors.New("sdk not initialized, call Subscribe first")
	}

	ob := s.manager.GetOrderBook(tokenID)
	if ob == nil {
		return nil, fmt.Errorf("%w: %s", ErrTokenNotFound, tokenID)
	}

	return ob, nil
}

// GetBestBid 获取最优买价（包括量）
func (s *SDK) GetBestBid(tokenID string) (*BestPrice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	// GetBestBid 内部会检查 initialized，并在同一个锁内返回数据
	result := ob.GetBestBid()
	if result == nil {
		return nil, ErrNoData
	}

	return result, nil
}

// GetBestAsk 获取最优卖价（包括量）
func (s *SDK) GetBestAsk(tokenID string) (*BestPrice, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.GetBestAsk()
	if result == nil {
		return nil, ErrNoData
	}

	return result, nil
}

// GetBBO 获取最优买卖价
func (s *SDK) GetBBO(tokenID string) (*BBO, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.GetBBO()
	if result == nil {
		return nil, ErrNotInitialized
	}

	return result, nil
}

// GetMidPrice 获取中间价
func (s *SDK) GetMidPrice(tokenID string) (decimal.Decimal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return decimal.Zero, err
	}

	result := ob.GetMidPrice()
	if result == nil {
		return decimal.Zero, ErrNoData
	}

	return *result, nil
}

// GetSpread 获取价差
func (s *SDK) GetSpread(tokenID string) (decimal.Decimal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return decimal.Zero, err
	}

	result := ob.GetSpread()
	if result == nil {
		return decimal.Zero, ErrNoData
	}

	return *result, nil
}

// GetDepth 获取指定深度的订单簿
func (s *SDK) GetDepth(tokenID string, depth int) (bids []OrderSummary, asks []OrderSummary, err error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, nil, err
	}

	bids, asks = ob.GetDepth(depth)
	if bids == nil && asks == nil {
		return nil, nil, ErrNotInitialized
	}

	return bids, asks, nil
}

// GetTotalBidSize 获取买单总量
func (s *SDK) GetTotalBidSize(tokenID string) (decimal.Decimal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return decimal.Zero, err
	}

	if !ob.IsInitialized() {
		return decimal.Zero, ErrNotInitialized
	}

	return ob.GetTotalBidSize(), nil
}

// GetTotalAskSize 获取卖单总量
func (s *SDK) GetTotalAskSize(tokenID string) (decimal.Decimal, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return decimal.Zero, err
	}

	if !ob.IsInitialized() {
		return decimal.Zero, ErrNotInitialized
	}

	return ob.GetTotalAskSize(), nil
}

// GetAllAsks 获取所有卖单（按价格升序）
func (s *SDK) GetAllAsks(tokenID string) ([]OrderSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.GetAllAsks()
	if result == nil {
		return nil, ErrNotInitialized
	}

	return result, nil
}

// GetAllBids 获取所有买单（按价格降序）
func (s *SDK) GetAllBids(tokenID string) ([]OrderSummary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.GetAllBids()
	if result == nil {
		return nil, ErrNotInitialized
	}

	return result, nil
}

// ScanAsksBelow 扫描价格低于等于 maxPrice 的所有卖单
// 返回可成交的订单列表 + 总数量 + 加权平均价格
func (s *SDK) ScanAsksBelow(tokenID string, maxPrice decimal.Decimal) (*ScanResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.ScanAsksBelow(maxPrice)
	if result == nil {
		return nil, ErrNotInitialized
	}

	return result, nil
}

// ScanBidsAbove 扫描价格高于等于 minPrice 的所有买单
// 返回可成交的订单列表 + 总数量 + 加权平均价格
func (s *SDK) ScanBidsAbove(tokenID string, minPrice decimal.Decimal) (*ScanResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.ScanBidsAbove(minPrice)
	if result == nil {
		return nil, ErrNotInitialized
	}

	return result, nil
}

// GetOrderBookTimestamp 获取订单簿最后更新时间戳
func (s *SDK) GetOrderBookTimestamp(tokenID string) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return 0, err
	}

	if !ob.IsInitialized() {
		return 0, ErrNotInitialized
	}

	return ob.Timestamp(), nil
}

// GetOrderBookHash 获取订单簿hash
func (s *SDK) GetOrderBookHash(tokenID string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return "", err
	}

	if !ob.IsInitialized() {
		return "", ErrNotInitialized
	}

	return ob.Hash(), nil
}

// SimulateBuyAsks 模拟买入卖单（吃单）
// 根据所需数量，从最优卖价开始累加，计算加权平均成交价格
// requiredSize: 需要买入的数量
// 返回: 成交结果，包含加权平均价格和是否能完全成交
func (s *SDK) SimulateBuyAsks(tokenID string, requiredSize decimal.Decimal) (*FillResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ob, err := s.getOrderBookLocked(tokenID)
	if err != nil {
		return nil, err
	}

	result := ob.SimulateBuyAsks(requiredSize)
	if result == nil {
		return nil, ErrNotInitialized
	}

	return result, nil
}
