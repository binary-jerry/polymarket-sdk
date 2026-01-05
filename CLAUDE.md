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

# 运行测试
go test ./...

# 运行测试（带覆盖率）
go test ./... -cover

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...

# 运行市场查询示例
go run examples/markets/main.go

# 运行交易示例（需要私钥）
go run examples/trading/main.go
```

## Architecture

### Module Structure

```
polymarket-sdk/
├── sdk.go              # 统一 SDK 入口
├── config.go           # 全局配置和常量
├── common/             # 公共模块
│   ├── errors.go       # 统一错误定义
│   ├── http.go         # HTTP 客户端封装
│   └── utils.go        # 工具函数
├── gamma/              # Gamma API（市场数据）
│   ├── client.go       # Gamma 客户端
│   ├── markets.go      # 市场查询方法
│   └── types.go        # 市场数据类型
├── auth/               # 认证模块
│   ├── types.go        # 认证类型定义
│   ├── signer.go       # 签名器接口
│   ├── l1_signer.go    # L1 EIP-712 签名
│   ├── l2_signer.go    # L2 HMAC 签名
│   └── credentials.go  # 凭证管理
├── clob/               # CLOB 交易模块
│   ├── client.go       # CLOB 客户端
│   ├── types.go        # 订单/交易类型
│   ├── signing.go      # 订单签名
│   ├── orders.go       # 订单操作
│   ├── account.go      # 账户查询
│   └── trades.go       # 交易历史
├── orderbook/          # 订单簿模块（WebSocket）
│   ├── sdk.go          # 订单簿 SDK
│   ├── manager.go      # 订单簿管理器
│   ├── ws_pool.go      # WebSocket 连接池
│   ├── ws_client.go    # WebSocket 客户端
│   └── orderbook.go    # 订单簿数据结构
└── examples/           # 示例代码
```

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

### Authentication Flow

```
1. L1 认证（创建 API Key）
   私钥 -> EIP-712 签名 -> POST /auth/api-key -> Credentials

2. L2 认证（API 请求）
   Credentials.Secret + 请求数据 -> HMAC-SHA256 -> 请求头
```

### Key Design Patterns

1. **连接池分片**: WebSocket 连接池自动分片（每连接50个token）
2. **惰性排序**: 订单簿使用 dirty flag，仅查询时排序
3. **自动重连**: 指数退避 + 抖动的重连策略
4. **统一错误处理**: 所有模块使用 common/errors.go 定义的错误类型

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

## Dependencies

- `github.com/gorilla/websocket`: WebSocket 客户端
- `github.com/shopspring/decimal`: 精确十进制运算
- `github.com/ethereum/go-ethereum`: EIP-712 签名
- `github.com/google/go-querystring`: URL 参数编码

## Testing

测试覆盖率目标：
- SDK 入口: >95%
- common: >85%
- gamma: >80%
- auth: >70%
- clob: >70%

运行特定模块测试：
```bash
go test ./auth/... -v
go test ./clob/... -v
go test ./gamma/... -v
```

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
