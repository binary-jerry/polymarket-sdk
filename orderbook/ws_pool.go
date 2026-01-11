package orderbook

import (
	"fmt"
	"log"
	"sync"
)

// WSPool WebSocket连接池
type WSPool struct {
	mu sync.RWMutex

	config  *Config
	clients []*WSClient

	// token到client的映射
	tokenToClient map[string]*WSClient

	// 下一个客户端ID
	nextClientID int

	// 消息处理回调
	onMessage func([]byte)
	// 状态变更回调
	onStateChange func(string, ConnectionState)

	// 是否已连接
	connected bool
}

// NewWSPool 创建新的连接池
func NewWSPool(config *Config) *WSPool {
	return &WSPool{
		config:        config,
		clients:       make([]*WSClient, 0),
		tokenToClient: make(map[string]*WSClient),
	}
}

// SetMessageHandler 设置消息处理回调
func (p *WSPool) SetMessageHandler(handler func([]byte)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onMessage = handler
}

// SetStateChangeHandler 设置状态变更回调
func (p *WSPool) SetStateChangeHandler(handler func(string, ConnectionState)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.onStateChange = handler
}

// Connect 建立初始连接（不订阅任何 token）
// 创建一个空闲的 WebSocket 连接，等待后续 Subscribe
func (p *WSPool) Connect() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.connected {
		return nil // 已经连接
	}

	// 创建一个空的客户端（不带 token）
	clientID := fmt.Sprintf("client-%d", p.nextClientID)
	p.nextClientID++

	client := NewWSClient(clientID, p.config.WSEndpoint, nil, p.config)

	// 设置消息处理回调
	if p.onMessage != nil {
		handler := p.onMessage
		client.SetMessageHandler(func(data []byte) {
			handler(data)
		})
	}

	// 设置状态变更回调
	if p.onStateChange != nil {
		stateHandler := p.onStateChange
		cid := clientID
		client.SetStateChangeHandler(func(state ConnectionState) {
			stateHandler(cid, state)
		})
	}

	// 建立连接（不发送订阅消息，因为没有 token）
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}

	p.clients = append(p.clients, client)
	p.connected = true

	return nil
}

// IsConnected 检查是否已连接
func (p *WSPool) IsConnected() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.connected
}

// Subscribe 订阅token列表（支持增量订阅）
func (p *WSPool) Subscribe(tokenIDs []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 过滤出真正新的 token
	newTokens := make([]string, 0)
	for _, tokenID := range tokenIDs {
		if _, exists := p.tokenToClient[tokenID]; !exists {
			newTokens = append(newTokens, tokenID)
		}
	}

	if len(newTokens) == 0 {
		return nil
	}

	// 如果还没有连接，先建立连接
	if !p.connected {
		p.mu.Unlock()
		if err := p.Connect(); err != nil {
			p.mu.Lock()
			return err
		}
		p.mu.Lock()
	}

	// 尝试将新 token 添加到现有连接中
	remaining := p.addToExistingClients(newTokens)

	// 如果还有剩余的 token，创建新连接
	if len(remaining) > 0 {
		if err := p.createNewClients(remaining); err != nil {
			return err
		}
	}

	return nil
}

// addToExistingClients 尝试将 token 添加到现有连接
// 返回无法添加的 token 列表
func (p *WSPool) addToExistingClients(tokenIDs []string) []string {
	remaining := make([]string, 0)
	tokensToAdd := tokenIDs

	for _, client := range p.clients {
		if len(tokensToAdd) == 0 {
			break
		}

		// 检查客户端是否还有空间
		currentCount := len(client.TokenIDs())
		available := p.config.MaxTokensPerConn - currentCount

		if available <= 0 {
			continue
		}

		// 计算可以添加多少个 token
		addCount := available
		if addCount > len(tokensToAdd) {
			addCount = len(tokensToAdd)
		}

		// 添加 token 到这个客户端
		tokensForClient := tokensToAdd[:addCount]
		if err := client.AddTokens(tokensForClient); err != nil {
			// 如果添加失败，保留到 remaining
			remaining = append(remaining, tokensForClient...)
		} else {
			// 更新映射
			for _, tokenID := range tokensForClient {
				p.tokenToClient[tokenID] = client
			}
		}

		// 更新剩余列表
		tokensToAdd = tokensToAdd[addCount:]
	}

	// 加上未处理的
	remaining = append(remaining, tokensToAdd...)
	return remaining
}

// createNewClients 为剩余的 token 创建新连接
func (p *WSPool) createNewClients(tokenIDs []string) error {
	// 按 MaxTokensPerConn 分组
	groups := p.groupTokens(tokenIDs)

	for _, group := range groups {
		clientID := fmt.Sprintf("client-%d", p.nextClientID)
		p.nextClientID++

		client := NewWSClient(clientID, p.config.WSEndpoint, group, p.config)

		// 设置消息处理回调
		if p.onMessage != nil {
			handler := p.onMessage
			client.SetMessageHandler(func(data []byte) {
				handler(data)
			})
		}

		// 设置状态变更回调
		if p.onStateChange != nil {
			stateHandler := p.onStateChange
			cid := clientID
			client.SetStateChangeHandler(func(state ConnectionState) {
				stateHandler(cid, state)
			})
		}

		// 建立连接（会自动发送订阅消息，因为有 token）
		if err := client.Connect(); err != nil {
			return fmt.Errorf("failed to connect client %s: %w", clientID, err)
		}

		p.clients = append(p.clients, client)

		// 建立token到client的映射
		for _, tokenID := range group {
			p.tokenToClient[tokenID] = client
		}
	}

	return nil
}

// Unsubscribe 取消订阅指定的 token
func (p *WSPool) Unsubscribe(tokenIDs []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 按客户端分组需要取消的 token
	clientTokens := make(map[*WSClient][]string)
	for _, tokenID := range tokenIDs {
		if client, exists := p.tokenToClient[tokenID]; exists {
			clientTokens[client] = append(clientTokens[client], tokenID)
			delete(p.tokenToClient, tokenID)
		}
	}

	// 通知每个客户端移除 token
	for client, tokens := range clientTokens {
		if err := client.RemoveTokens(tokens); err != nil {
			log.Printf("[WSPool] failed to unsubscribe tokens from client %s: %v", client.ID(), err)
		}

		// 如果客户端没有任何 token 了，但保留连接（作为空闲连接）
		// 不关闭它，以便后续可以复用
	}

	return nil
}

// removeClient 从客户端列表中移除指定客户端
func (p *WSPool) removeClient(client *WSClient) {
	newClients := make([]*WSClient, 0, len(p.clients)-1)
	for _, c := range p.clients {
		if c != client {
			newClients = append(newClients, c)
		}
	}
	p.clients = newClients
}

// groupTokens 将token列表按MaxTokensPerConn分组
func (p *WSPool) groupTokens(tokenIDs []string) [][]string {
	maxPerConn := p.config.MaxTokensPerConn
	groups := make([][]string, 0)

	for i := 0; i < len(tokenIDs); i += maxPerConn {
		end := i + maxPerConn
		if end > len(tokenIDs) {
			end = len(tokenIDs)
		}
		groups = append(groups, tokenIDs[i:end])
	}

	return groups
}

// GetClientForToken 获取负责指定token的客户端
func (p *WSPool) GetClientForToken(tokenID string) *WSClient {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.tokenToClient[tokenID]
}

// GetAllClients 获取所有客户端
func (p *WSPool) GetAllClients() []*WSClient {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]*WSClient, len(p.clients))
	copy(result, p.clients)
	return result
}

// GetClientCount 获取客户端数量
func (p *WSPool) GetClientCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.clients)
}

// GetTokenCount 获取订阅的token总数
func (p *WSPool) GetTokenCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.tokenToClient)
}

// IsAllActive 检查所有连接是否都处于活跃状态
func (p *WSPool) IsAllActive() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, client := range p.clients {
		state := client.GetState()
		if state != StateActive && state != StateConnected {
			return false
		}
	}

	return len(p.clients) > 0
}

// GetStatus 获取连接池状态
func (p *WSPool) GetStatus() map[string]ConnectionState {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := make(map[string]ConnectionState)
	for _, client := range p.clients {
		status[client.ID()] = client.GetState()
	}

	return status
}

// Close 关闭连接池
func (p *WSPool) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, client := range p.clients {
		client.Close()
	}

	p.clients = nil
	p.tokenToClient = make(map[string]*WSClient)
	p.connected = false
}
