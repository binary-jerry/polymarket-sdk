package orderbook

import (
	"fmt"
	"sync"
)

// WSPool WebSocket连接池
type WSPool struct {
	mu sync.RWMutex

	config  *Config
	clients []*WSClient

	// token到client的映射
	tokenToClient map[string]*WSClient

	// 消息处理回调
	onMessage func([]byte)
	// 状态变更回调
	onStateChange func(string, ConnectionState)
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

// Subscribe 订阅token列表
func (p *WSPool) Subscribe(tokenIDs []string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// 按MaxTokensPerConn分组
	groups := p.groupTokens(tokenIDs)

	for i, group := range groups {
		clientID := fmt.Sprintf("client-%d", i)
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

		// 建立连接
		if err := client.Connect(); err != nil {
			// 清理已创建的连接
			for _, c := range p.clients {
				c.Close()
			}
			p.clients = nil
			p.tokenToClient = make(map[string]*WSClient)
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
		if client.GetState() != StateActive {
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
}
