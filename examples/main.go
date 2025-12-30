package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/binary-jerry/polymarket-sdk/orderbook"
	"github.com/shopspring/decimal"
)

// 全局数据库连接
var db *sql.DB

func main() {
	// 初始化数据库连接
	var err error
	db, err = sql.Open("mysql", "root:Daheng467.@tcp(127.0.0.1:3306)/polymarket?parseTime=true")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 测试数据库连接
	if err = db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Println("Database connected successfully")

	// 创建SDK配置（可选，使用默认配置）
	config := orderbook.DefaultConfig()
	// 可以自定义配置
	// config.ReconnectMaxAttempts = 10
	// config.PingInterval = 20

	// 创建SDK实例
	sdk := orderbook.NewSDK(config)
	defer sdk.Close()

	// 要订阅的token列表
	// 这里使用示例token ID，实际使用时替换为真实的token ID
	tokenIDs := []string{
		"86048179007629022807705037775458342506338650261339576882051926945843401279995",
		"28309349346511781004115470700800075779705976016724226277609716433615157969772",
		// 可以添加更多token，超过50个会自动创建新的WebSocket连接
	}

	// 订阅
	log.Println("Subscribing to tokens...")
	if err := sdk.Subscribe(tokenIDs); err != nil {
		log.Fatalf("Failed to subscribe: %v", err)
	}
	log.Println("Subscribed successfully")

	// 启动更新监听goroutine
	go func() {
		updates := sdk.Updates()
		if updates == nil {
			return
		}

		for range updates {
			printOrderBookInfo(sdk, tokenIDs)
		}
	}()

	// 等待初始化完成
	log.Println("Waiting for orderbooks to initialize...")
	for !sdk.IsAllInitialized() {
		time.Sleep(100 * time.Millisecond)
	}
	log.Println("All orderbooks initialized")

	// 演示各种API调用
	for _, tokenID := range tokenIDs {
		demonstrateAPI(sdk, tokenID)
	}

	// 优雅关闭
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Println("Press Ctrl+C to exit...")
	<-sigChan

	log.Println("Shutting down...")
}

func printOrderBookInfo(sdk *orderbook.SDK, tokenIDs []string) {
	yes, err := sdk.GetBestAsk(tokenIDs[0])
	if err != nil {
		log.Printf("GetBestAsk error: %v", err)
		return
	}
	no, err := sdk.GetBestAsk(tokenIDs[1])
	if err != nil {
		log.Printf("GetBestAsk error: %v", err)
		return
	}
	priceSum := yes.Price.Add(no.Price)

	_, err = db.Exec(`
		INSERT INTO orderbook
		(yes_token_id, no_token_id, yes_price, yes_size, no_price, no_size, price_sum, yes_time, no_time)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		tokenIDs[0],
		tokenIDs[1],
		yes.Price.String(),
		yes.Size.String(),
		no.Price.String(),
		no.Size.String(),
		priceSum.String(),
		yes.Timestamp,
		no.Timestamp,
	)
	if err != nil {
		log.Printf("Failed to insert into database: %v", err)
		return
	}

	if priceSum.LessThan(decimal.NewFromInt(1)) {
		log.Printf("Yes Price %s, No Price %s, sum %s \n", yes.Price, no.Price, priceSum)
	}
}

func demonstrateAPI(sdk *orderbook.SDK, tokenID string) {
	log.Printf("\n=== Demonstrating API for token: %s ===\n", tokenID[:20]+"...")

	// 1. 获取最优买价
	bestBid, err := sdk.GetBestBid(tokenID)
	if err != nil {
		log.Printf("GetBestBid error: %v", err)
	} else if bestBid != nil {
		log.Printf("Best Bid: Price=%s, Size=%s", bestBid.Price, bestBid.Size)
	}

	// 2. 获取最优卖价
	bestAsk, err := sdk.GetBestAsk(tokenID)
	if err != nil {
		log.Printf("GetBestAsk error: %v", err)
	} else if bestAsk != nil {
		log.Printf("Best Ask: Price=%s, Size=%s", bestAsk.Price, bestAsk.Size)
	}

	// 3. 获取中间价
	midPrice, err := sdk.GetMidPrice(tokenID)
	if err != nil {
		log.Printf("GetMidPrice error: %v", err)
	} else {
		log.Printf("Mid Price: %s", midPrice)
	}

	// 4. 获取价差
	spread, err := sdk.GetSpread(tokenID)
	if err != nil {
		log.Printf("GetSpread error: %v", err)
	} else {
		log.Printf("Spread: %s", spread)
	}

	// 5. 获取指定深度的订单簿
	bids, asks, err := sdk.GetDepth(tokenID, 5)
	if err != nil {
		log.Printf("GetDepth error: %v", err)
	} else {
		log.Printf("Top 5 Bids:")
		for i, bid := range bids {
			log.Printf("  %d. Price=%s, Size=%s", i+1, bid.Price, bid.Size)
		}
		log.Printf("Top 5 Asks:")
		for i, ask := range asks {
			log.Printf("  %d. Price=%s, Size=%s", i+1, ask.Price, ask.Size)
		}
	}

	// 6. 获取买单总量
	totalBidSize, err := sdk.GetTotalBidSize(tokenID)
	if err != nil {
		log.Printf("GetTotalBidSize error: %v", err)
	} else {
		log.Printf("Total Bid Size: %s", totalBidSize)
	}

	// 7. 获取卖单总量
	totalAskSize, err := sdk.GetTotalAskSize(tokenID)
	if err != nil {
		log.Printf("GetTotalAskSize error: %v", err)
	} else {
		log.Printf("Total Ask Size: %s", totalAskSize)
	}

	// 8. 获取所有买单
	allBids, err := sdk.GetAllBids(tokenID)
	if err != nil {
		log.Printf("GetAllBids error: %v", err)
	} else {
		log.Printf("All Bids count: %d", len(allBids))
	}

	// 9. 获取所有卖单
	allAsks, err := sdk.GetAllAsks(tokenID)
	if err != nil {
		log.Printf("GetAllAsks error: %v", err)
	} else {
		log.Printf("All Asks count: %d", len(allAsks))
	}

	// 10. 扫描卖单
	maxPrice := decimal.NewFromFloat(0.6)
	scanResult, err := sdk.ScanAsksBelow(tokenID, maxPrice)
	if err != nil {
		log.Printf("ScanAsksBelow error: %v", err)
	} else {
		log.Printf("Scan Asks Below %s:", maxPrice)
		log.Printf("  Orders count: %d", len(scanResult.Orders))
		log.Printf("  Total Size: %s", scanResult.TotalSize)
		log.Printf("  Avg Price: %s", scanResult.AvgPrice)
	}

	// 11. 扫描买单
	minPrice := decimal.NewFromFloat(0.4)
	scanResult, err = sdk.ScanBidsAbove(tokenID, minPrice)
	if err != nil {
		log.Printf("ScanBidsAbove error: %v", err)
	} else {
		log.Printf("Scan Bids Above %s:", minPrice)
		log.Printf("  Orders count: %d", len(scanResult.Orders))
		log.Printf("  Total Size: %s", scanResult.TotalSize)
		log.Printf("  Avg Price: %s", scanResult.AvgPrice)
	}

	// 12. 获取订单簿时间戳
	timestamp, err := sdk.GetOrderBookTimestamp(tokenID)
	if err != nil {
		log.Printf("GetOrderBookTimestamp error: %v", err)
	} else {
		log.Printf("OrderBook Timestamp: %d", timestamp)
	}

	// 13. 获取订单簿hash
	hash, err := sdk.GetOrderBookHash(tokenID)
	if err != nil {
		log.Printf("GetOrderBookHash error: %v", err)
	} else {
		log.Printf("OrderBook Hash: %s", hash)
	}

	// 14. 获取连接状态
	status := sdk.GetConnectionStatus()
	log.Printf("Connection Status:")
	for clientID, state := range status {
		log.Printf("  %s: %s", clientID, state)
	}

	fmt.Println()
}
