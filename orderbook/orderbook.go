package orderbook

import (
	"sort"
	"sync"

	"github.com/shopspring/decimal"
)

// OrderBook 订单簿
type OrderBook struct {
	mu sync.RWMutex

	tokenID     string
	market      string
	hash        string
	timestamp   int64 // 上次更新时间戳（毫秒）
	initialized bool  // 是否已初始化（收到过book消息）

	// 买单：按价格降序排列，使用map存储便于O(1)更新
	bids map[string]decimal.Decimal // price -> size
	// 卖单：按价格升序排列
	asks map[string]decimal.Decimal // price -> size

	// 缓存的排序后的价格档位
	sortedBids []OrderSummary
	sortedAsks []OrderSummary
	bidsDirty  bool
	asksDirty  bool
}

// NewOrderBook 创建新的订单簿
func NewOrderBook(tokenID string) *OrderBook {
	return &OrderBook{
		tokenID:   tokenID,
		bids:      make(map[string]decimal.Decimal),
		asks:      make(map[string]decimal.Decimal),
		bidsDirty: true,
		asksDirty: true,
	}
}

// Reset 重置订单簿状态（保留tokenID，清空其他数据）
func (ob *OrderBook) Reset() {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	ob.market = ""
	ob.hash = ""
	ob.timestamp = 0
	ob.initialized = false
	ob.bids = make(map[string]decimal.Decimal)
	ob.asks = make(map[string]decimal.Decimal)
	ob.sortedBids = nil
	ob.sortedAsks = nil
	ob.bidsDirty = true
	ob.asksDirty = true
}

// TokenID 获取token ID
func (ob *OrderBook) TokenID() string {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.tokenID
}

// Market 获取market ID
func (ob *OrderBook) Market() string {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.market
}

// Hash 获取订单簿hash
func (ob *OrderBook) Hash() string {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.hash
}

// Timestamp 获取上次更新时间戳
func (ob *OrderBook) Timestamp() int64 {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.timestamp
}

// IsInitialized 检查订单簿是否已初始化
func (ob *OrderBook) IsInitialized() bool {
	ob.mu.RLock()
	defer ob.mu.RUnlock()
	return ob.initialized
}

// ApplyBookSnapshot 应用完整订单簿快照
func (ob *OrderBook) ApplyBookSnapshot(msg *BookMessage, ts int64) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// 时间戳检查：如果是旧消息则丢弃
	if ts < ob.timestamp {
		return false
	}

	// 清空现有数据
	ob.bids = make(map[string]decimal.Decimal)
	ob.asks = make(map[string]decimal.Decimal)

	// 应用买单
	for _, bid := range msg.Bids {
		price, err := decimal.NewFromString(bid.Price)
		if err != nil {
			continue
		}
		size, err := decimal.NewFromString(bid.Size)
		if err != nil {
			continue
		}
		if size.IsPositive() {
			ob.bids[bid.Price] = size
		}
		_ = price // 用于验证价格格式
	}

	// 应用卖单
	for _, ask := range msg.Asks {
		price, err := decimal.NewFromString(ask.Price)
		if err != nil {
			continue
		}
		size, err := decimal.NewFromString(ask.Size)
		if err != nil {
			continue
		}
		if size.IsPositive() {
			ob.asks[ask.Price] = size
		}
		_ = price
	}

	ob.market = msg.Market
	ob.hash = msg.Hash
	ob.timestamp = ts
	ob.initialized = true
	ob.bidsDirty = true
	ob.asksDirty = true

	return true
}

// ApplyPriceChange 应用价格变动
func (ob *OrderBook) ApplyPriceChange(change *PriceChange, ts int64) bool {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	// 必须先初始化
	if !ob.initialized {
		return false
	}

	// 时间戳检查：如果是旧消息则丢弃
	if ts < ob.timestamp {
		return false
	}

	size, err := decimal.NewFromString(change.Size)
	if err != nil {
		return false
	}

	if change.Side == SideBuy {
		if size.IsZero() {
			delete(ob.bids, change.Price)
		} else {
			ob.bids[change.Price] = size
		}
		ob.bidsDirty = true
	} else if change.Side == SideSell {
		if size.IsZero() {
			delete(ob.asks, change.Price)
		} else {
			ob.asks[change.Price] = size
		}
		ob.asksDirty = true
	}

	ob.hash = change.Hash
	ob.timestamp = ts

	return true
}

// rebuildSortedBids 重建排序后的买单列表（内部调用，需持有锁）
func (ob *OrderBook) rebuildSortedBids() {
	if !ob.bidsDirty {
		return
	}

	ob.sortedBids = make([]OrderSummary, 0, len(ob.bids))
	for priceStr, size := range ob.bids {
		price, _ := decimal.NewFromString(priceStr)
		ob.sortedBids = append(ob.sortedBids, OrderSummary{
			Price: price,
			Size:  size,
		})
	}

	// 按价格降序排列
	sort.Slice(ob.sortedBids, func(i, j int) bool {
		return ob.sortedBids[i].Price.GreaterThan(ob.sortedBids[j].Price)
	})

	ob.bidsDirty = false
}

// rebuildSortedAsks 重建排序后的卖单列表（内部调用，需持有锁）
func (ob *OrderBook) rebuildSortedAsks() {
	if !ob.asksDirty {
		return
	}

	ob.sortedAsks = make([]OrderSummary, 0, len(ob.asks))
	for priceStr, size := range ob.asks {
		price, _ := decimal.NewFromString(priceStr)
		ob.sortedAsks = append(ob.sortedAsks, OrderSummary{
			Price: price,
			Size:  size,
		})
	}

	// 按价格升序排列
	sort.Slice(ob.sortedAsks, func(i, j int) bool {
		return ob.sortedAsks[i].Price.LessThan(ob.sortedAsks[j].Price)
	})

	ob.asksDirty = false
}

// GetBestBid 获取最优买价（包括量）
func (ob *OrderBook) GetBestBid() *BestPrice {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized || len(ob.bids) == 0 {
		return nil
	}

	ob.rebuildSortedBids()
	if len(ob.sortedBids) == 0 {
		return nil
	}

	return &BestPrice{
		Price: ob.sortedBids[0].Price,
		Size:  ob.sortedBids[0].Size,
	}
}

// GetBestAsk 获取最优卖价（包括量）
func (ob *OrderBook) GetBestAsk() *BestPrice {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized || len(ob.asks) == 0 {
		return nil
	}

	ob.rebuildSortedAsks()
	if len(ob.sortedAsks) == 0 {
		return nil
	}

	return &BestPrice{
		Price:     ob.sortedAsks[0].Price,
		Size:      ob.sortedAsks[0].Size,
		Timestamp: ob.timestamp,
	}
}

// GetBBO 获取最优买卖价
func (ob *OrderBook) GetBBO() *BBO {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedBids()
	ob.rebuildSortedAsks()

	bbo := &BBO{}

	if len(ob.sortedBids) > 0 {
		bbo.BestBid = &BestPrice{
			Price: ob.sortedBids[0].Price,
			Size:  ob.sortedBids[0].Size,
		}
	}

	if len(ob.sortedAsks) > 0 {
		bbo.BestAsk = &BestPrice{
			Price: ob.sortedAsks[0].Price,
			Size:  ob.sortedAsks[0].Size,
		}
	}

	return bbo
}

// GetMidPrice 获取中间价
func (ob *OrderBook) GetMidPrice() *decimal.Decimal {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedBids()
	ob.rebuildSortedAsks()

	if len(ob.sortedBids) == 0 || len(ob.sortedAsks) == 0 {
		return nil
	}

	mid := ob.sortedBids[0].Price.Add(ob.sortedAsks[0].Price).Div(decimal.NewFromInt(2))
	return &mid
}

// GetSpread 获取价差
func (ob *OrderBook) GetSpread() *decimal.Decimal {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedBids()
	ob.rebuildSortedAsks()

	if len(ob.sortedBids) == 0 || len(ob.sortedAsks) == 0 {
		return nil
	}

	spread := ob.sortedAsks[0].Price.Sub(ob.sortedBids[0].Price)
	return &spread
}

// GetDepth 获取指定深度的订单簿
func (ob *OrderBook) GetDepth(depth int) (bids []OrderSummary, asks []OrderSummary) {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil, nil
	}

	ob.rebuildSortedBids()
	ob.rebuildSortedAsks()

	// 复制买单
	bidCount := depth
	if bidCount > len(ob.sortedBids) {
		bidCount = len(ob.sortedBids)
	}
	bids = make([]OrderSummary, bidCount)
	copy(bids, ob.sortedBids[:bidCount])

	// 复制卖单
	askCount := depth
	if askCount > len(ob.sortedAsks) {
		askCount = len(ob.sortedAsks)
	}
	asks = make([]OrderSummary, askCount)
	copy(asks, ob.sortedAsks[:askCount])

	return bids, asks
}

// GetTotalBidSize 获取买单总量
func (ob *OrderBook) GetTotalBidSize() decimal.Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if !ob.initialized {
		return decimal.Zero
	}

	total := decimal.Zero
	for _, size := range ob.bids {
		total = total.Add(size)
	}
	return total
}

// GetTotalAskSize 获取卖单总量
func (ob *OrderBook) GetTotalAskSize() decimal.Decimal {
	ob.mu.RLock()
	defer ob.mu.RUnlock()

	if !ob.initialized {
		return decimal.Zero
	}

	total := decimal.Zero
	for _, size := range ob.asks {
		total = total.Add(size)
	}
	return total
}

// GetAllAsks 获取所有卖单（按价格升序）
func (ob *OrderBook) GetAllAsks() []OrderSummary {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedAsks()

	result := make([]OrderSummary, len(ob.sortedAsks))
	copy(result, ob.sortedAsks)
	return result
}

// GetAllBids 获取所有买单（按价格降序）
func (ob *OrderBook) GetAllBids() []OrderSummary {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedBids()

	result := make([]OrderSummary, len(ob.sortedBids))
	copy(result, ob.sortedBids)
	return result
}

// ScanAsksBelow 扫描价格低于等于 maxPrice 的所有卖单
// 返回可成交的订单列表 + 总数量 + 加权平均价格
func (ob *OrderBook) ScanAsksBelow(maxPrice decimal.Decimal) *ScanResult {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedAsks()

	result := &ScanResult{
		Orders:    make([]OrderSummary, 0),
		TotalSize: decimal.Zero,
		AvgPrice:  decimal.Zero,
	}

	totalValue := decimal.Zero

	for _, order := range ob.sortedAsks {
		if order.Price.LessThanOrEqual(maxPrice) {
			result.Orders = append(result.Orders, order)
			result.TotalSize = result.TotalSize.Add(order.Size)
			totalValue = totalValue.Add(order.Price.Mul(order.Size))
		} else {
			// 因为是升序排列，超过maxPrice后面的都不符合条件
			break
		}
	}

	if result.TotalSize.IsPositive() {
		result.AvgPrice = totalValue.Div(result.TotalSize)
	}

	return result
}

// ScanBidsAbove 扫描价格高于等于 minPrice 的所有买单
// 返回可成交的订单列表 + 总数量 + 加权平均价格
func (ob *OrderBook) ScanBidsAbove(minPrice decimal.Decimal) *ScanResult {
	ob.mu.Lock()
	defer ob.mu.Unlock()

	if !ob.initialized {
		return nil
	}

	ob.rebuildSortedBids()

	result := &ScanResult{
		Orders:    make([]OrderSummary, 0),
		TotalSize: decimal.Zero,
		AvgPrice:  decimal.Zero,
	}

	totalValue := decimal.Zero

	for _, order := range ob.sortedBids {
		if order.Price.GreaterThanOrEqual(minPrice) {
			result.Orders = append(result.Orders, order)
			result.TotalSize = result.TotalSize.Add(order.Size)
			totalValue = totalValue.Add(order.Price.Mul(order.Size))
		} else {
			// 因为是降序排列，低于minPrice后面的都不符合条件
			break
		}
	}

	if result.TotalSize.IsPositive() {
		result.AvgPrice = totalValue.Div(result.TotalSize)
	}

	return result
}
