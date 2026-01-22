# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Polymarket SDK 是一个 Go 语言编写的完整 Polymarket 预测市场 SDK，提供：
- 实时订单簿订阅（WebSocket）
- 市场数据查询（Gamma API）
- 交易操作（CLOB API）
- EIP-712 签名和 HMAC 认证

## Build and Run Commands

```bash
# 构建
go build ./...

# 运行所有测试
go test ./...

# 运行单个模块测试
go test ./auth/... -v
go test ./clob/... -v
go test ./gamma/... -v
go test ./orderbook/... -v

# 运行单个测试函数
go test ./auth/... -v -run TestL1Signer

# 运行测试（带覆盖率）
go test ./... -cover

# 运行测试（带竞态检测）
go test ./... -race

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...

# 运行完整示例（订单簿订阅 + 套利检测）
go run examples/main.go

# 运行市场查询示例
go run examples/markets/main.go

# 运行交易示例（需要私钥）
go run examples/trading/main.go
```

## Architecture

SDK 采用模块化分层架构，包含 5 个主要模块：
- `sdk.go` / `config.go` - 统一入口和全局配置
- `common/` - 公共模块（错误定义、HTTP 客户端、工具函数）
- `gamma/` - Gamma API（市场数据查询）
- `auth/` - 认证模块（L1 EIP-712 签名、L2 HMAC 签名）
- `clob/` - CLOB 交易模块（订单操作、账户查询）
- `orderbook/` - 订单簿模块（WebSocket 实时订阅）

### SDK Initialization

```go
// 公开 SDK（无需私钥，仅查询）
sdk := polymarket.NewPublicSDK(nil)

// 完整 SDK（需要私钥，支持交易）
sdk, err := polymarket.NewSDK(nil, privateKey)

// 带已有凭证的交易 SDK
sdk, err := polymarket.NewTradingSDK(nil, privateKey, creds)
```

### Core Components

#### 1. 统一入口 (sdk.go)
- `SDK` 结构体整合所有子模块
- `OrderBook` - 实时订单簿
- `Markets` - 市场数据查询
- `Trading` - 交易操作

#### 2. Gamma API (gamma/)
- 市场列表查询
- 单个市场详情
- 按分类/标签筛选
- 搜索市场

#### 3. Auth 模块 (auth/)
- **L1Signer**: EIP-712 类型数据签名（钱包签名）
- **L2Signer**: HMAC-SHA256 签名（API 请求签名）
- **CredentialsManager**: API 凭证创建和管理

#### 4. CLOB 模块 (clob/)
- 订单创建/取消
- 批量订单操作
- 余额/持仓查询
- 交易历史

#### 5. OrderBook 模块 (orderbook/)
- WebSocket 实时订阅
- 订单簿快照和增量更新
- 连接池管理（每连接最多50个token）

### OrderBook 架构层次

```
┌─────────────────────┐
│   SDK (Public API)  │  - Subscribe/Unsubscribe, GetDepth, GetBBO
└──────────┬──────────┘
           │
┌──────────▼──────────┐
│   Manager           │  - 消息路由、订单簿状态管理、pending 消息缓存
└──────────┬──────────┘
           │
┌──────────▼──────────┐
│   WSPool            │  - 连接池，自动分片（50 token/连接）
└──────────┬──────────┘
           │
┌──────────▼──────────┐
│  WSClient (x N)     │  - 单连接、心跳、指数退避重连
└─────────────────────┘
```

### Authentication Flow

```
1. L1 认证（创建 API Key）
   私钥 -> EIP-712 签名 -> POST /auth/api-key -> Credentials

2. L2 认证（API 请求）
   Credentials.Secret + 请求数据 -> HMAC-SHA256 -> 请求头
```

### Order Signing Flow

```
CreateOrderRequest
    ↓
OrderSigner.CreateSignedOrder() (EIP-712 签名)
    ↓
PostOrderRequest (已签名)
    ↓
L2Signer.GetAuthHeaders() (HMAC 签名)
    ↓
HTTP 请求到 CLOB API
```

### Key Design Patterns

1. **连接池分片**: WebSocket 连接池自动分片（每连接50个token）
2. **惰性排序**: 订单簿使用 dirty flag，仅查询时排序
3. **自动重连**: 指数退避 + 抖动的重连策略
4. **统一错误处理**: 所有模块使用 common/errors.go 定义的错误类型

### Default Configuration

| 配置项 | 默认值 |
|--------|--------|
| HTTPTimeout | 30s |
| MaxRetries | 3 |
| RetryDelayMs | 1000 |
| MaxTokensPerConn | 50 |
| ReconnectMinInterval | 1000ms |
| ReconnectMaxInterval | 30000ms |
| PingInterval | 30s |
| PongTimeout | 10s |

### Contract Addresses (Polygon Mainnet)

| 合约 | 地址 |
|------|------|
| CTF Exchange | 0x4bFb41d5B3570DeFd03C39a9A4D8De6Bd8b8982e |
| NegRisk Exchange | 0xC5d563A36AE78145C45a50134d48A1215220f80a |
| NegRisk Adapter | 0xd91E80cF2E7be2e162c6513ceD06f1dD0dA35296 |
| Collateral (USDC) | 0x2791Bca1f2de4661ED88A30C99A7a9449Aa84174 |

### API Endpoints

| API | 端点 |
|-----|------|
| Gamma (Markets) | https://gamma-api.polymarket.com |
| CLOB (Trading) | https://clob.polymarket.com |
| WebSocket | wss://ws-subscriptions-clob.polymarket.com/ws/market |

## Key Code Paths

| 功能 | 文件位置 |
|------|---------|
| SDK 创建入口 | `sdk.go:NewSDK()` |
| OrderBook WebSocket 消息处理 | `orderbook/manager.go:handleMessage()` |
| WebSocket 连接管理 | `orderbook/ws_client.go` |
| 订单簿数据结构 | `orderbook/orderbook.go` |
| L1 EIP-712 签名 | `auth/l1_signer.go:Sign()` |
| L2 HMAC 签名 | `auth/l2_signer.go:Sign()` |
| 订单创建和签名 | `clob/orders.go` / `clob/signing.go` |
| 市场查询 | `gamma/markets.go` |
| 错误类型定义 | `common/errors.go` |

## Dependencies

- `github.com/gorilla/websocket`: WebSocket 客户端
- `github.com/shopspring/decimal`: 精确十进制运算
- `github.com/ethereum/go-ethereum`: EIP-712 签名
- `github.com/google/go-querystring`: URL 参数编码

## 官方 SDK 参考规范

**开发 Golang 版本的 SDK 相关功能时，必须先阅读官方提供的 Python 版本 SDK 作为参考。**

### Polymarket 官方 Python SDK 目录
```
/Users/houjie/web3/polymarket/predict-arb/vendor-sdk/py_clob_client
```

### 开发流程
1. **先阅读官方 SDK**: 在实现任何 SDK 功能前，必须先查看 `vendor-sdk` 目录中对应的官方实现
2. **严格遵循官方逻辑**: 实现方式必须与官方 SDK 保持一致，包括：
   - API 调用方式
   - 签名算法
   - 数据结构
   - 错误处理
3. **记录差异**: 如果因语言特性需要调整，需在代码注释中说明与官方实现的差异

## NegRisk Markets

NegRisk 市场使用不同的合约地址：
- 使用 `NegRiskExchangeAddress` 而非 `CTFExchangeAddress`
- 订单的 `taker` 字段设为 `NegRiskAdapterAddress`
- 通过 `CreateOrderRequest.IsNegRisk = true` 标识

## Common Pitfalls

1. **地址格式**: EIP-712 签名需要校验和格式的地址字符串（使用 `common.Address.Hex()`）
2. **USDC 精度**: USDC 使用 6 位小数（`Decimal6 = 1000000`）
3. **API 凭证**: 首次使用需要调用 `CreateOrDeriveAPICredentials()` 获取凭证
4. **订单类型**: 支持 GTC、GTD、FOK、FAK 四种订单类型
5. **批量订单限制**: `CreateOrders` 最多支持 15 个订单
6. **订单簿初始化**: 调用 `GetDepth` 前必须等待 `IsInitialized(tokenID)` 返回 true
7. **WebSocket 订阅**: `Subscribe()` 只能调用一次，不能追加订阅
8. **Updates Channel**: 如果不消费 `Updates()` channel，缓冲区满时旧消息会被丢弃

## OrderBook 数据一致性保障

### 初始化机制
- **initialized 状态**: 订单簿需要收到快照后才标记为已初始化
- **pendingChanges 缓冲**: 在快照到达前的增量更新会被缓冲，快照后按序列号重放
- **序列号检查**: 只接受比当前序列号更新的增量更新

### 重连处理
- 断开连接时自动清空所有订单簿数据
- 重连后重新订阅并等待新快照
- 防止混合新旧数据导致不一致

### 获取订单簿数据
```go
// 获取完整深度（从内存）
bids, asks, err := sdk.OrderBook.GetDepth(tokenID, 50)

// 获取所有买单/卖单
allBids, err := sdk.OrderBook.GetAllBids(tokenID)
allAsks, err := sdk.OrderBook.GetAllAsks(tokenID)

// 扫描特定价格范围
scanResult, err := sdk.OrderBook.ScanAsksBelow(tokenID, maxPrice)
scanResult, err := sdk.OrderBook.ScanBidsAbove(tokenID, minPrice)

// 模拟成交计算加权平均价格
fillResult, err := sdk.OrderBook.SimulateBuyAsks(tokenID, requiredSize)
```

### 重要提醒
- **始终从 SDK 内存获取订单簿**: 使用 `GetDepth`/`GetAllBids`/`GetAllAsks`，不要调用 REST API
- **检查初始化状态**: 使用前调用 `IsInitialized(tokenID)` 确认订单簿已就绪
- **处理深度而非仅 BBO**: 套利计算应考虑完整深度，使用 `ScanAsksBelow`/`ScanBidsAbove`

## 单元测试规范

**所有完成的功能必须完成单元测试才能交付。**

### 测试文件规范
- 测试文件必须以 `_test.go` 结尾（例如 `orderbook_test.go`）
- 测试文件与被测试文件放在同一目录下
- **禁止**单独编写 `main.go` 文件进行测试

### 测试命名规范
```go
// 测试函数命名
func TestFunctionName(t *testing.T)           // 基本测试
func TestFunctionName_Scenario(t *testing.T)  // 场景测试
func TestType_MethodName(t *testing.T)        // 方法测试

// 示例
func TestOrderBook_Apply(t *testing.T)
func TestL1Signer_Sign(t *testing.T)
func TestWSClient_Reconnect(t *testing.T)
```

### 测试运行
```bash
go test ./...                                    # 全部测试
go test ./orderbook/... -v                       # 单个包
go test ./auth/... -run TestL1Signer             # 单个测试
go test ./... -cover                             # 带覆盖率
go test ./... -race                              # 竞态检测
```

### 交付检查清单
- [ ] 新功能已编写对应的 `_test.go` 文件
- [ ] 测试覆盖主要逻辑路径和边界条件
- [ ] `go test ./...` 全部通过
- [ ] 未使用 `main.go` 进行测试
