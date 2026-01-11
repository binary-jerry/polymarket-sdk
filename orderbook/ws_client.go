package orderbook

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// WSClient WebSocket客户端（单连接）
type WSClient struct {
	mu sync.RWMutex

	id       string   // 客户端唯一标识
	endpoint string   // WebSocket端点
	tokenIDs []string // 订阅的token列表
	config   *Config

	conn  *websocket.Conn
	state ConnectionState

	// 消息处理回调
	onMessage func([]byte)
	// 状态变更回调
	onStateChange func(ConnectionState)

	// 控制通道
	ctx       context.Context
	cancel    context.CancelFunc
	writeChan chan []byte
	closeChan chan struct{}
	closeOnce sync.Once

	// goroutine 生命周期控制
	loopCtx    context.Context
	loopCancel context.CancelFunc
	loopWg     sync.WaitGroup

	// 重连控制
	reconnectAttempts int32
	reconnecting      int32 // 原子标记，防止多次触发重连

	// 心跳控制
	lastPong time.Time
}

// NewWSClient 创建新的WebSocket客户端
func NewWSClient(id string, endpoint string, tokenIDs []string, config *Config) *WSClient {
	ctx, cancel := context.WithCancel(context.Background())

	return &WSClient{
		id:        id,
		endpoint:  endpoint,
		tokenIDs:  tokenIDs,
		config:    config,
		state:     StateDisconnected,
		ctx:       ctx,
		cancel:    cancel,
		writeChan: make(chan []byte, config.MessageBufferSize),
		closeChan: make(chan struct{}),
		lastPong:  time.Now(),
	}
}

// SetMessageHandler 设置消息处理回调
func (c *WSClient) SetMessageHandler(handler func([]byte)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onMessage = handler
}

// SetStateChangeHandler 设置状态变更回调
func (c *WSClient) SetStateChangeHandler(handler func(ConnectionState)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onStateChange = handler
}

// GetState 获取当前连接状态
func (c *WSClient) GetState() ConnectionState {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.state
}

// setState 设置连接状态（内部调用）
func (c *WSClient) setState(state ConnectionState) {
	c.mu.Lock()
	oldState := c.state
	c.state = state
	handler := c.onStateChange
	c.mu.Unlock()

	if oldState != state && handler != nil {
		handler(state)
	}
}

// Connect 建立连接（不发送订阅消息）
func (c *WSClient) Connect() error {
	// 先停止旧的 goroutine
	c.stopLoops()

	c.setState(StateConnecting)

	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(c.ctx, c.endpoint, nil)
	if err != nil {
		c.setState(StateDisconnected)
		return err
	}

	// 创建新的 loop context
	c.mu.Lock()
	c.conn = conn
	c.lastPong = time.Now()
	c.loopCtx, c.loopCancel = context.WithCancel(c.ctx)
	c.mu.Unlock()

	c.setState(StateConnected)

	// 设置pong处理
	conn.SetPongHandler(func(appData string) error {
		c.mu.Lock()
		c.lastPong = time.Now()
		c.mu.Unlock()
		return nil
	})

	// 启动goroutines
	c.loopWg.Add(3)
	go c.readLoop()
	go c.writeLoop()
	go c.heartbeatLoop()

	// 如果有初始 token，立即订阅（使用初始订阅格式）
	if len(c.tokenIDs) > 0 {
		if err := c.sendInitialSubscribe(c.tokenIDs); err != nil {
			c.stopLoops()
			c.closeConnection()
			return err
		}
	}

	c.setState(StateActive)
	atomic.StoreInt32(&c.reconnectAttempts, 0)
	atomic.StoreInt32(&c.reconnecting, 0)

	return nil
}

// sendInitialSubscribe 发送初始订阅请求（连接时使用 type: "MARKET"）
func (c *WSClient) sendInitialSubscribe(tokenIDs []string) error {
	if len(tokenIDs) == 0 {
		return nil
	}

	req := SubscribeRequest{
		AssetsIDs: tokenIDs,
		Type:      "MARKET",
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	select {
	case c.writeChan <- data:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-loopCtx.Done():
		return loopCtx.Err()
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}

// sendDynamicSubscribe 发送动态订阅请求（连接后添加订阅使用 operation: "subscribe"）
func (c *WSClient) sendDynamicSubscribe(tokenIDs []string) error {
	return c.sendDynamicOperation(tokenIDs, "subscribe")
}

// sendDynamicUnsubscribe 发送动态取消订阅请求（使用 operation: "unsubscribe"）
func (c *WSClient) sendDynamicUnsubscribe(tokenIDs []string) error {
	return c.sendDynamicOperation(tokenIDs, "unsubscribe")
}

// sendDynamicOperation 发送动态操作请求
func (c *WSClient) sendDynamicOperation(tokenIDs []string, operation string) error {
	if len(tokenIDs) == 0 {
		return nil
	}

	req := DynamicSubscribeRequest{
		AssetsIDs: tokenIDs,
		Operation: operation,
	}

	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	select {
	case c.writeChan <- data:
		return nil
	case <-c.ctx.Done():
		return c.ctx.Err()
	case <-loopCtx.Done():
		return loopCtx.Err()
	case <-time.After(5 * time.Second):
		return context.DeadlineExceeded
	}
}

// stopLoops 停止所有循环 goroutine
func (c *WSClient) stopLoops() {
	c.mu.Lock()
	if c.loopCancel != nil {
		c.loopCancel()
	}
	c.mu.Unlock()

	// 等待所有 goroutine 退出
	c.loopWg.Wait()
}

// readLoop 读取消息循环
func (c *WSClient) readLoop() {
	defer c.loopWg.Done()
	defer c.triggerReconnect()

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.closeChan:
			return
		case <-loopCtx.Done():
			return
		default:
		}

		c.mu.RLock()
		conn := c.conn
		c.mu.RUnlock()

		if conn == nil {
			return
		}

		// 设置读取超时
		conn.SetReadDeadline(time.Now().Add(time.Duration(c.config.PingInterval+c.config.PongTimeout) * time.Second))

		_, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("[WSClient %s] read error: %v", c.id, err)
			}
			return
		}

		c.mu.RLock()
		handler := c.onMessage
		c.mu.RUnlock()

		if handler != nil {
			handler(message)
		}
	}
}

// writeLoop 写入消息循环
func (c *WSClient) writeLoop() {
	defer c.loopWg.Done()

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.closeChan:
			return
		case <-loopCtx.Done():
			return
		case data := <-c.writeChan:
			c.mu.RLock()
			conn := c.conn
			c.mu.RUnlock()

			if conn == nil {
				continue
			}

			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("[WSClient %s] write error: %v", c.id, err)
				return
			}
		}
	}
}

// heartbeatLoop 心跳循环
func (c *WSClient) heartbeatLoop() {
	defer c.loopWg.Done()

	ticker := time.NewTicker(time.Duration(c.config.PingInterval) * time.Second)
	defer ticker.Stop()

	c.mu.RLock()
	loopCtx := c.loopCtx
	c.mu.RUnlock()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.closeChan:
			return
		case <-loopCtx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			conn := c.conn
			lastPong := c.lastPong
			c.mu.RUnlock()

			if conn == nil {
				return
			}

			// 检查pong超时
			if time.Since(lastPong) > time.Duration(c.config.PingInterval+c.config.PongTimeout)*time.Second {
				log.Printf("[WSClient %s] pong timeout, reconnecting...", c.id)
				return
			}

			// 发送ping
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("[WSClient %s] ping error: %v", c.id, err)
				return
			}
		}
	}
}

// triggerReconnect 触发重连（确保只触发一次）
func (c *WSClient) triggerReconnect() {
	// 检查是否已关闭
	select {
	case <-c.closeChan:
		return
	case <-c.ctx.Done():
		return
	default:
	}

	c.mu.RLock()
	currentState := c.state
	c.mu.RUnlock()

	if currentState == StateClosed {
		return
	}

	// 使用 CAS 确保只有一个 goroutine 触发重连
	if !atomic.CompareAndSwapInt32(&c.reconnecting, 0, 1) {
		return
	}

	// 取消 loopCtx，通知所有 loop goroutine 退出
	c.mu.Lock()
	if c.loopCancel != nil {
		c.loopCancel()
	}
	c.mu.Unlock()

	c.closeConnection()
	c.setState(StateReconnecting)

	// 启动重连（在新 goroutine 中，因为当前 goroutine 要退出）
	go c.reconnect()
}

// closeConnection 关闭当前连接（不触发重连）
func (c *WSClient) closeConnection() {
	c.mu.Lock()
	conn := c.conn
	c.conn = nil
	c.mu.Unlock()

	if conn != nil {
		conn.Close()
	}
}

// reconnect 重连逻辑
func (c *WSClient) reconnect() {
	// 等待旧的 goroutine 退出
	c.loopWg.Wait()

	// 清空 writeChan 中的旧消息
	c.drainWriteChan()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-c.closeChan:
			return
		default:
		}

		attempts := atomic.AddInt32(&c.reconnectAttempts, 1)

		// 检查最大重连次数
		if c.config.ReconnectMaxAttempts > 0 && int(attempts) > c.config.ReconnectMaxAttempts {
			log.Printf("[WSClient %s] max reconnect attempts reached", c.id)
			c.setState(StateDisconnected)
			atomic.StoreInt32(&c.reconnecting, 0)
			return
		}

		// 计算退避时间（指数退避 + 抖动）
		backoff := c.calculateBackoff(int(attempts))
		log.Printf("[WSClient %s] reconnecting in %v (attempt %d)", c.id, backoff, attempts)

		select {
		case <-time.After(backoff):
		case <-c.ctx.Done():
			return
		case <-c.closeChan:
			return
		}

		// 尝试重连
		if err := c.Connect(); err != nil {
			log.Printf("[WSClient %s] reconnect failed: %v", c.id, err)
			continue
		}

		log.Printf("[WSClient %s] reconnected successfully", c.id)
		return
	}
}

// drainWriteChan 清空写入通道中的旧消息
func (c *WSClient) drainWriteChan() {
	for {
		select {
		case <-c.writeChan:
		default:
			return
		}
	}
}

// calculateBackoff 计算退避时间
func (c *WSClient) calculateBackoff(attempts int) time.Duration {
	minInterval := time.Duration(c.config.ReconnectMinInterval) * time.Millisecond
	maxInterval := time.Duration(c.config.ReconnectMaxInterval) * time.Millisecond

	// 指数退避
	backoff := minInterval * time.Duration(1<<uint(attempts-1))
	if backoff > maxInterval {
		backoff = maxInterval
	}

	// 添加抖动（±20%）
	jitter := time.Duration(rand.Float64()*0.4-0.2) * backoff
	backoff += jitter

	if backoff < minInterval {
		backoff = minInterval
	}

	return backoff
}

// Close 关闭客户端
func (c *WSClient) Close() {
	c.closeOnce.Do(func() {
		c.setState(StateClosed)
		c.cancel()
		close(c.closeChan)
		c.stopLoops()
		c.closeConnection()
	})
}

// ID 获取客户端ID
func (c *WSClient) ID() string {
	return c.id
}

// TokenIDs 获取订阅的token列表
func (c *WSClient) TokenIDs() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	result := make([]string, len(c.tokenIDs))
	copy(result, c.tokenIDs)
	return result
}

// AddTokens 动态添加新的 token 订阅
func (c *WSClient) AddTokens(tokenIDs []string) error {
	c.mu.Lock()
	// 检查连接状态
	if c.state != StateActive && c.state != StateConnected {
		c.mu.Unlock()
		return fmt.Errorf("client not active, current state: %s", c.state)
	}

	// 添加到 token 列表
	c.tokenIDs = append(c.tokenIDs, tokenIDs...)
	c.mu.Unlock()

	// 发送动态订阅请求（使用 operation: "subscribe"）
	return c.sendDynamicSubscribe(tokenIDs)
}

// RemoveTokens 从订阅列表中移除 token，并发送取消订阅请求
func (c *WSClient) RemoveTokens(tokenIDs []string) error {
	c.mu.Lock()
	// 检查连接状态
	if c.state != StateActive && c.state != StateConnected {
		c.mu.Unlock()
		return fmt.Errorf("client not active, current state: %s", c.state)
	}

	// 创建要移除的 token 集合
	toRemove := make(map[string]bool)
	for _, tokenID := range tokenIDs {
		toRemove[tokenID] = true
	}

	// 过滤 token 列表
	newTokenIDs := make([]string, 0, len(c.tokenIDs))
	for _, tokenID := range c.tokenIDs {
		if !toRemove[tokenID] {
			newTokenIDs = append(newTokenIDs, tokenID)
		}
	}
	c.tokenIDs = newTokenIDs
	c.mu.Unlock()

	// 发送动态取消订阅请求
	return c.sendDynamicUnsubscribe(tokenIDs)
}
