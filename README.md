# Polymarket OrderBook SDK

一个高性能的 Go 语言 SDK，用于实时订阅和查询 Polymarket 预测市场的订单簿数据。

## 功能特性

- **实时订单簿同步**: 通过 WebSocket 实时接收订单簿快照和增量更新
- **连接池管理**: 自动分片管理多个 WebSocket 连接，每个连接最多订阅 50 个 token
- **自动重连**: 内置指数退避重连机制，保证连接稳定性
- **心跳保活**: 自动发送 ping/pong 心跳，检测连接状态
- **线程安全**: 所有 API 都是并发安全的
- **高精度计算**: 使用 decimal 库进行价格和数量计算，避免浮点精度问题

## 安装

```bash
go get github.com/binary-jerry/polymarket-sdk
```

## 快速开始

```go
package main

import (
    "log"
    "time"

    "github.com/binary-jerry/polymarket-sdk/orderbook"
)

func main() {
    // 创建 SDK 实例（使用默认配置）
    sdk := orderbook.NewSDK(nil)
    defer sdk.Close()

    // 订阅 token 列表
    tokenIDs := []string{
        "86048179007629022807705037775458342506338650261339576882051926945843401279995",
        "28309349346511781004115470700800075779705976016724226277609716433615157969772",
    }

    if err := sdk.Subscribe(tokenIDs); err != nil {
        log.Fatalf("订阅失败: %v", err)
    }

    // 等待订单簿初始化
    for !sdk.IsAllInitialized() {
        time.Sleep(100 * time.Millisecond)
    }

    // 查询最优买卖价
    for _, tokenID := range tokenIDs {
        bestBid, _ := sdk.GetBestBid(tokenID)
        bestAsk, _ := sdk.GetBestAsk(tokenID)
        log.Printf("Token: %s...", tokenID[:20])
        log.Printf("  Best Bid: %s @ %s", bestBid.Size, bestBid.Price)
        log.Printf("  Best Ask: %s @ %s", bestAsk.Size, bestAsk.Price)
    }
}
```

## 配置选项

```go
config := &orderbook.Config{
    // WebSocket 端点
    WSEndpoint: "wss://ws-subscriptions-clob.polymarket.com/ws/market",

    // 每个连接最大 token 数量（超过会自动创建新连接）
    MaxTokensPerConn: 50,

    // 重连配置
    ReconnectMinInterval: 1000,  // 最小重连间隔（毫秒）
    ReconnectMaxInterval: 30000, // 最大重连间隔（毫秒）
    ReconnectMaxAttempts: 0,     // 最大重连次数，0 表示无限重连

    // 心跳配置
    PingInterval: 30, // ping 间隔（秒）
    PongTimeout:  10, // pong 超时（秒）

    // 缓冲区配置
    MessageBufferSize: 1000, // 消息缓冲区大小
    UpdateChannelSize: 1000, // 更新通知 channel 大小
}

sdk := orderbook.NewSDK(config)
```

## API 文档

### 初始化与生命周期

| 方法 | 说明 |
|------|------|
| `NewSDK(config *Config) *SDK` | 创建 SDK 实例，传 nil 使用默认配置 |
| `Subscribe(tokenIDs []string) error` | 订阅 token 列表，只能调用一次 |
| `Close()` | 关闭 SDK，释放所有资源 |

### 状态查询

| 方法 | 说明 |
|------|------|
| `IsInitialized(tokenID string) bool` | 检查指定 token 的订单簿是否已初始化 |
| `IsAllInitialized() bool` | 检查所有订单簿是否都已初始化 |
| `GetConnectionStatus() map[string]ConnectionState` | 获取所有连接的状态 |

### 价格查询

| 方法 | 说明 |
|------|------|
| `GetBestBid(tokenID string) (*BestPrice, error)` | 获取最优买价（价格 + 数量） |
| `GetBestAsk(tokenID string) (*BestPrice, error)` | 获取最优卖价（价格 + 数量） |
| `GetBBO(tokenID string) (*BBO, error)` | 获取最优买卖价 |
| `GetMidPrice(tokenID string) (decimal.Decimal, error)` | 获取中间价 |
| `GetSpread(tokenID string) (decimal.Decimal, error)` | 获取买卖价差 |

### 深度查询

| 方法 | 说明 |
|------|------|
| `GetDepth(tokenID string, depth int) (bids, asks []OrderSummary, error)` | 获取指定深度的订单簿 |
| `GetAllBids(tokenID string) ([]OrderSummary, error)` | 获取所有买单（按价格降序） |
| `GetAllAsks(tokenID string) ([]OrderSummary, error)` | 获取所有卖单（按价格升序） |
| `GetTotalBidSize(tokenID string) (decimal.Decimal, error)` | 获取买单总量 |
| `GetTotalAskSize(tokenID string) (decimal.Decimal, error)` | 获取卖单总量 |

### 扫描查询

| 方法 | 说明 |
|------|------|
| `ScanAsksBelow(tokenID string, maxPrice decimal.Decimal) (*ScanResult, error)` | 扫描价格 ≤ maxPrice 的所有卖单 |
| `ScanBidsAbove(tokenID string, minPrice decimal.Decimal) (*ScanResult, error)` | 扫描价格 ≥ minPrice 的所有买单 |

扫描结果包含：
- `Orders`: 符合条件的订单列表
- `TotalSize`: 总数量
- `AvgPrice`: 加权平均价格

### 元数据查询

| 方法 | 说明 |
|------|------|
| `GetOrderBookTimestamp(tokenID string) (int64, error)` | 获取订单簿最后更新时间戳（毫秒） |
| `GetOrderBookHash(tokenID string) (string, error)` | 获取订单簿 hash |

### 更新通知

```go
// 获取更新通知 channel
updates := sdk.Updates()

go func() {
    for update := range updates {
        log.Printf("Token %s 更新, 类型: %s, 时间戳: %d",
            update.TokenID, update.EventType, update.Timestamp)
    }
}()
```

## 数据类型

### BestPrice

```go
type BestPrice struct {
    Price     decimal.Decimal // 价格
    Size      decimal.Decimal // 数量
    Timestamp int64           // 时间戳（毫秒）
}
```

### OrderSummary

```go
type OrderSummary struct {
    Price decimal.Decimal // 价格档位
    Size  decimal.Decimal // 该档位总数量
}
```

### ScanResult

```go
type ScanResult struct {
    Orders    []OrderSummary  // 符合条件的订单列表
    TotalSize decimal.Decimal // 总数量
    AvgPrice  decimal.Decimal // 加权平均价格
}
```

### ConnectionState

```go
const (
    StateDisconnected  // 已断开
    StateConnecting    // 连接中
    StateConnected     // 已连接
    StateSubscribing   // 订阅中
    StateActive        // 活跃（正常工作）
    StateReconnecting  // 重连中
    StateClosed        // 已关闭
)
```

### EventType

```go
const (
    EventTypeBook           = "book"             // 订单簿快照
    EventTypePriceChange    = "price_change"     // 价格变动
    EventTypeTickSizeChange = "tick_size_change" // tick size 变更
    EventTypeLastTradePrice = "last_trade_price" // 最后成交价
)
```

## 架构设计

```
┌─────────────────────────────────────────────────────────────┐
│                          SDK                                 │
│  (对外接口层 - 提供所有查询方法)                              │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Manager                               │
│  (管理层 - 消息分发、订单簿状态管理、pending消息缓存)          │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        WSPool                                │
│  (连接池 - 管理多个WebSocket客户端，token到客户端映射)         │
└─────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┼───────────────┐
              ▼               ▼               ▼
        ┌──────────┐    ┌──────────┐    ┌──────────┐
        │ WSClient │    │ WSClient │    │ WSClient │
        │ (50 tokens)│   │ (50 tokens)│   │ (N tokens)│
        └──────────┘    └──────────┘    └──────────┘
              │               │               │
              └───────────────┴───────────────┘
                              │
                              ▼
                    Polymarket WebSocket Server
```

### 核心组件

| 组件 | 职责 |
|------|------|
| **SDK** | 对外暴露的主接口，封装所有查询方法，线程安全 |
| **Manager** | 管理所有订单簿实例，处理消息分发，缓存未初始化时的增量消息 |
| **WSPool** | 连接池，按 MaxTokensPerConn 自动分片创建多个 WebSocket 连接 |
| **WSClient** | 单个 WebSocket 连接，处理连接、心跳、重连逻辑 |
| **OrderBook** | 单个 token 的订单簿，维护买卖盘数据，惰性排序 |

### 消息处理流程

1. **初始化阶段**
   - 连接 WebSocket 后发送订阅请求
   - 服务器返回 `book` 消息（订单簿完整快照）
   - 在收到快照前，`price_change` 消息会被缓存

2. **增量更新阶段**
   - 收到 `price_change` 消息后更新对应价格档位
   - 如果 size 为 0，删除该档位
   - 更新后标记排序缓存为 dirty

3. **重连处理**
   - 检测到连接断开后，重置相关订单簿状态
   - 使用指数退避策略重连
   - 重连成功后重新订阅，等待新的快照

## 使用示例

### 监听价格变动

```go
sdk := orderbook.NewSDK(nil)
defer sdk.Close()

tokenIDs := []string{"your-token-id"}
sdk.Subscribe(tokenIDs)

// 等待初始化
for !sdk.IsAllInitialized() {
    time.Sleep(100 * time.Millisecond)
}

// 监听更新
for update := range sdk.Updates() {
    if update.EventType == orderbook.EventTypePriceChange {
        bbo, _ := sdk.GetBBO(update.TokenID)
        log.Printf("BBO 更新: Bid=%s, Ask=%s",
            bbo.BestBid.Price, bbo.BestAsk.Price)
    }
}
```

### 套利机会检测

```go
// 假设 tokenIDs[0] 是 YES token，tokenIDs[1] 是 NO token
for update := range sdk.Updates() {
    yesAsk, _ := sdk.GetBestAsk(tokenIDs[0])
    noAsk, _ := sdk.GetBestAsk(tokenIDs[1])

    if yesAsk != nil && noAsk != nil {
        priceSum := yesAsk.Price.Add(noAsk.Price)
        if priceSum.LessThan(decimal.NewFromInt(1)) {
            log.Printf("套利机会! YES=%s + NO=%s = %s < 1",
                yesAsk.Price, noAsk.Price, priceSum)
        }
    }
}
```

### 查询可成交深度

```go
// 查询价格 ≤ 0.55 的所有卖单
maxPrice := decimal.NewFromFloat(0.55)
result, err := sdk.ScanAsksBelow(tokenID, maxPrice)
if err == nil {
    log.Printf("可买入数量: %s, 平均价格: %s",
        result.TotalSize, result.AvgPrice)
}

// 查询价格 ≥ 0.45 的所有买单
minPrice := decimal.NewFromFloat(0.45)
result, err = sdk.ScanBidsAbove(tokenID, minPrice)
if err == nil {
    log.Printf("可卖出数量: %s, 平均价格: %s",
        result.TotalSize, result.AvgPrice)
}
```

## 错误处理

SDK 定义了以下错误类型：

```go
var (
    ErrNotInitialized = errors.New("orderbook not initialized")
    ErrTokenNotFound  = errors.New("token not found")
    ErrNoData         = errors.New("no data available")
    ErrAlreadyStarted = errors.New("sdk already started")
)
```

使用示例：

```go
bestBid, err := sdk.GetBestBid(tokenID)
if err != nil {
    if errors.Is(err, orderbook.ErrNotInitialized) {
        // 订单簿尚未初始化，等待或重试
    } else if errors.Is(err, orderbook.ErrTokenNotFound) {
        // token 未订阅
    } else if errors.Is(err, orderbook.ErrNoData) {
        // 暂无数据（订单簿为空）
    }
}
```

## 注意事项

1. **Subscribe 只能调用一次**: 订阅后不能追加或修改订阅列表
2. **等待初始化**: 查询前应等待 `IsAllInitialized()` 返回 true
3. **处理 Updates channel**: 如果不消费 Updates channel，当缓冲区满时旧消息会被丢弃
4. **资源释放**: 使用完毕后调用 `Close()` 释放资源

## 依赖

- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket 客户端
- [shopspring/decimal](https://github.com/shopspring/decimal) - 高精度十进制运算

## License

MIT
