package orderbook

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
)

// Manager 订单簿管理器
type Manager struct {
	mu sync.RWMutex

	config *Config
	pool   *WSPool

	// tokenID -> OrderBook
	orderBooks map[string]*OrderBook

	// 已订阅的 token 集合
	subscribedTokens map[string]bool

	// 更新通知channel
	updateChan chan OrderBookUpdate

	// 待处理的price_change消息（订单簿初始化前）
	pendingChanges map[string][]*pendingPriceChange

	// 关闭控制
	closeChan chan struct{}
	closeOnce sync.Once
}

// pendingPriceChange 待处理的价格变动
type pendingPriceChange struct {
	change    *PriceChange
	timestamp int64
}

// NewManager 创建新的订单簿管理器
func NewManager(config *Config) *Manager {
	if config == nil {
		config = DefaultConfig()
	}

	m := &Manager{
		config:           config,
		orderBooks:       make(map[string]*OrderBook),
		subscribedTokens: make(map[string]bool),
		updateChan:       make(chan OrderBookUpdate, config.UpdateChannelSize),
		pendingChanges:   make(map[string][]*pendingPriceChange),
		closeChan:        make(chan struct{}),
	}

	return m
}

// Connect 建立 WebSocket 连接（不订阅任何 token）
// 这是 "Connect first, Subscribe later" 模式的第一步
func (m *Manager) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 如果连接池不存在，创建并设置回调
	if m.pool == nil {
		m.pool = NewWSPool(m.config)

		// 设置消息处理回调
		m.pool.SetMessageHandler(m.handleMessage)

		// 设置状态变更回调
		m.pool.SetStateChangeHandler(func(clientID string, state ConnectionState) {
			log.Printf("[Manager] client %s state changed to %s", clientID, state)

			// 如果断开连接，清除相关订单簿的初始化状态
			if state == StateReconnecting || state == StateDisconnected {
				m.handleClientDisconnect(clientID)
			}
		})
	}

	// 建立连接
	return m.pool.Connect()
}

// IsConnected 检查是否已连接
func (m *Manager) IsConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.pool == nil {
		return false
	}
	return m.pool.IsConnected()
}

// Subscribe 订阅token列表（支持增量订阅）
func (m *Manager) Subscribe(tokenIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 过滤出新的 token（未订阅过的）
	newTokens := make([]string, 0)
	for _, tokenID := range tokenIDs {
		if !m.subscribedTokens[tokenID] {
			newTokens = append(newTokens, tokenID)
			m.subscribedTokens[tokenID] = true
		}
	}

	// 如果没有新 token，直接返回
	if len(newTokens) == 0 {
		return nil
	}

	// 初始化新 token 的订单簿
	for _, tokenID := range newTokens {
		if _, exists := m.orderBooks[tokenID]; !exists {
			m.orderBooks[tokenID] = NewOrderBook(tokenID)
			m.pendingChanges[tokenID] = make([]*pendingPriceChange, 0)
		}
	}

	// 如果连接池不存在，创建并设置回调
	if m.pool == nil {
		m.pool = NewWSPool(m.config)

		// 设置消息处理回调
		m.pool.SetMessageHandler(m.handleMessage)

		// 设置状态变更回调
		m.pool.SetStateChangeHandler(func(clientID string, state ConnectionState) {
			log.Printf("[Manager] client %s state changed to %s", clientID, state)

			// 如果断开连接，清除相关订单簿的初始化状态
			if state == StateReconnecting || state == StateDisconnected {
				m.handleClientDisconnect(clientID)
			}
		})
	}

	// 向连接池添加订阅
	return m.pool.Subscribe(newTokens)
}

// Unsubscribe 取消订阅指定的 token
func (m *Manager) Unsubscribe(tokenIDs []string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, tokenID := range tokenIDs {
		delete(m.subscribedTokens, tokenID)
		delete(m.orderBooks, tokenID)
		delete(m.pendingChanges, tokenID)
	}

	if m.pool != nil {
		return m.pool.Unsubscribe(tokenIDs)
	}

	return nil
}

// GetSubscribedTokens 获取已订阅的 token 列表
func (m *Manager) GetSubscribedTokens() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokens := make([]string, 0, len(m.subscribedTokens))
	for tokenID := range m.subscribedTokens {
		tokens = append(tokens, tokenID)
	}
	return tokens
}

// handleClientDisconnect 处理客户端断开连接
func (m *Manager) handleClientDisconnect(clientID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.pool == nil {
		return
	}

	// 遍历所有订单簿，重置与该客户端相关的
	for tokenID, ob := range m.orderBooks {
		c := m.pool.GetClientForToken(tokenID)
		if c != nil && c.ID() == clientID {
			// 重置订单簿状态（保留对象引用，避免外部持有旧引用的问题）
			ob.Reset()
			m.pendingChanges[tokenID] = make([]*pendingPriceChange, 0)
			log.Printf("[Manager] reset orderbook for token %s due to client %s disconnect", tokenID, clientID)
		}
	}
}

// handleMessage 处理WebSocket消息
// 支持两种格式：
// 1. 数组格式（初始化订阅时批量发送）：[{event_type: "book", ...}, ...]
// 2. 单个对象格式（后续增量更新）：{event_type: "book", ...}
func (m *Manager) handleMessage(data []byte) {
	// 检查是否是数组格式（以 '[' 开头）
	if len(data) > 0 && data[0] == '[' {
		m.handleMessageArray(data)
		return
	}

	// 单个对象格式
	m.handleSingleMessage(data)
}

// handleMessageArray 处理消息数组
func (m *Manager) handleMessageArray(data []byte) {
	var rawMessages []json.RawMessage
	if err := json.Unmarshal(data, &rawMessages); err != nil {
		log.Printf("[Manager] failed to unmarshal message array: %v", err)
		return
	}

	log.Printf("[Manager] received batch of %d messages", len(rawMessages))

	for _, rawMsg := range rawMessages {
		m.handleSingleMessage(rawMsg)
	}
}

// handleSingleMessage 处理单条消息
func (m *Manager) handleSingleMessage(data []byte) {
	// 首先解析消息类型
	var raw RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		log.Printf("[Manager] failed to unmarshal raw message: %v", err)
		return
	}

	switch raw.EventType {
	case EventTypeBook:
		m.handleBookMessage(data)
	case EventTypePriceChange:
		m.handlePriceChangeMessage(data)
	case EventTypeTickSizeChange:
		// 暂不处理tick size变更
		//log.Printf("[Manager] received tick_size_change message")
	case EventTypeLastTradePrice:
		// 暂不处理最后成交价
		//log.Printf("[Manager] received last_trade_price message")
	default:
		//log.Printf("[Manager] unknown event type: %s", raw.EventType)
	}
}

// handleBookMessage 处理订单簿快照消息
func (m *Manager) handleBookMessage(data []byte) {
	var msg BookMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[Manager] failed to unmarshal book message: %v", err)
		return
	}

	// 解析时间戳
	ts, err := strconv.ParseInt(msg.Timestamp, 10, 64)
	if err != nil {
		log.Printf("[Manager] failed to parse timestamp: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	ob, exists := m.orderBooks[msg.AssetID]
	if !exists {
		log.Printf("[Manager] received book for unknown token: %s", msg.AssetID)
		return
	}

	// 应用快照
	if ob.ApplyBookSnapshot(&msg, ts) {
		//log.Printf("[Manager] applied book snapshot for token %s, bids: %d, asks: %d",
		//	msg.AssetID, len(msg.Bids), len(msg.Asks))

		// 应用待处理的price_change消息
		pending := m.pendingChanges[msg.AssetID]
		appliedCount := 0
		for _, p := range pending {
			if p.timestamp >= ts {
				if ob.ApplyPriceChange(p.change, p.timestamp) {
					appliedCount++
				}
			}
		}
		if appliedCount > 0 {
			log.Printf("[Manager] applied %d pending price changes for token %s", appliedCount, msg.AssetID)
		}

		// 清空待处理消息
		m.pendingChanges[msg.AssetID] = make([]*pendingPriceChange, 0)

		// 发送更新通知
		m.sendUpdate(OrderBookUpdate{
			TokenID:   msg.AssetID,
			EventType: EventTypeBook,
			Timestamp: ts,
		})
	}
}

// handlePriceChangeMessage 处理价格变动消息
func (m *Manager) handlePriceChangeMessage(data []byte) {
	var msg PriceChangeMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Printf("[Manager] failed to unmarshal price_change message: %v", err)
		return
	}

	// 解析时间戳
	ts, err := strconv.ParseInt(msg.Timestamp, 10, 64)
	if err != nil {
		log.Printf("[Manager] failed to parse timestamp: %v", err)
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// 处理每个价格变动
	for _, change := range msg.PriceChanges {
		changeCopy := change // 创建副本避免闭包问题

		ob, exists := m.orderBooks[change.AssetID]
		if !exists {
			log.Printf("[Manager] received price_change for unknown token: %s", change.AssetID)
			continue
		}

		// 如果订单簿未初始化，缓存消息
		if !ob.IsInitialized() {
			m.pendingChanges[change.AssetID] = append(m.pendingChanges[change.AssetID], &pendingPriceChange{
				change:    &changeCopy,
				timestamp: ts,
			})
			continue
		}

		// 应用价格变动
		if ob.ApplyPriceChange(&changeCopy, ts) {
			// 发送更新通知
			m.sendUpdate(OrderBookUpdate{
				TokenID:   change.AssetID,
				EventType: EventTypePriceChange,
				Timestamp: ts,
			})
		}
	}
}

// sendUpdate 发送更新通知
func (m *Manager) sendUpdate(update OrderBookUpdate) {
	select {
	case m.updateChan <- update:
	default:
		// channel满了，丢弃旧消息
		select {
		case <-m.updateChan:
		default:
		}
		select {
		case m.updateChan <- update:
		default:
		}
	}
}

// Updates 获取更新通知channel
func (m *Manager) Updates() <-chan OrderBookUpdate {
	return m.updateChan
}

// GetOrderBook 获取指定token的订单簿
func (m *Manager) GetOrderBook(tokenID string) *OrderBook {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.orderBooks[tokenID]
}

// GetAllOrderBooks 获取所有订单簿
func (m *Manager) GetAllOrderBooks() map[string]*OrderBook {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*OrderBook)
	for k, v := range m.orderBooks {
		result[k] = v
	}
	return result
}

// IsInitialized 检查指定token的订单簿是否已初始化
func (m *Manager) IsInitialized(tokenID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ob, exists := m.orderBooks[tokenID]
	if !exists {
		return false
	}
	return ob.IsInitialized()
}

// IsAllInitialized 检查所有订单簿是否都已初始化
func (m *Manager) IsAllInitialized() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, ob := range m.orderBooks {
		if !ob.IsInitialized() {
			return false
		}
	}

	return len(m.orderBooks) > 0
}

// GetConnectionStatus 获取连接状态
func (m *Manager) GetConnectionStatus() map[string]ConnectionState {
	if m.pool == nil {
		return nil
	}
	return m.pool.GetStatus()
}

// Close 关闭管理器
func (m *Manager) Close() {
	m.closeOnce.Do(func() {
		close(m.closeChan)

		if m.pool != nil {
			m.pool.Close()
		}

		close(m.updateChan)
	})
}
