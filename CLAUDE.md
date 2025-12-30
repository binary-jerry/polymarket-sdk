# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Polymarket SDK是一个Go语言编写的Polymarket订单簿实时数据SDK。通过WebSocket连接订阅市场数据，实时维护本地订单簿状态。

## Build and Run Commands

```bash
# 构建
go build ./...

# 运行示例（需要MySQL数据库）
go run examples/main.go

# 格式化代码
go fmt ./...

# 检查代码问题
go vet ./...
```

## Architecture

### Core Components (orderbook package)

- **SDK** (`sdk.go`): 对外暴露的主接口，封装所有订单簿查询方法
- **Manager** (`manager.go`): 订单簿管理器，处理WebSocket消息分发和订单簿状态管理
- **WSPool** (`ws_pool.go`): WebSocket连接池，管理多个WebSocket客户端（每个连接最多50个token）
- **WSClient** (`ws_client.go`): 单个WebSocket客户端，处理连接、心跳、重连
- **OrderBook** (`orderbook.go`): 单个token的订单簿数据结构，维护买卖盘数据

### Data Flow

```
Polymarket WS -> WSClient -> WSPool -> Manager -> OrderBook
                                             ↓
                                    SDK (query interface)
```

### Key Design Patterns

1. **连接池分片**: 超过50个token自动创建新的WebSocket连接
2. **消息缓存**: 在收到订单簿快照前，缓存价格变动消息（pendingChanges）
3. **惰性排序**: 订单簿使用dirty flag，只在查询时重建排序缓存
4. **自动重连**: 指数退避 + 抖动的重连策略

### Message Types

- `book`: 订单簿完整快照（初始化时）
- `price_change`: 增量价格变动
- `tick_size_change`: tick size变更（暂未处理）
- `last_trade_price`: 最后成交价（暂未处理）

## Dependencies

- `github.com/gorilla/websocket`: WebSocket客户端
- `github.com/shopspring/decimal`: 精确十进制运算
- `github.com/go-sql-driver/mysql`: MySQL驱动（仅示例使用）
