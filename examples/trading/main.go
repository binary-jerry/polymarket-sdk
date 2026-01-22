package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/shopspring/decimal"

	polymarket "github.com/binary-jerry/polymarket-sdk"
	"github.com/binary-jerry/polymarket-sdk/clob"
)

func main() {
	// 从环境变量获取私钥
	privateKey := os.Getenv("POLYMARKET_PRIVATE_KEY")
	if privateKey == "" {
		log.Fatal("请设置 POLYMARKET_PRIVATE_KEY 环境变量")
	}

	// 创建 SDK
	sdk, err := polymarket.NewSDK(nil, privateKey)
	if err != nil {
		log.Fatalf("创建 SDK 失败: %v", err)
	}
	defer sdk.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("钱包地址: %s\n\n", sdk.GetAddress())

	// 1. 创建或衍生 API 凭证
	fmt.Println("=== 创建 API 凭证 ===")
	creds, err := sdk.CreateOrDeriveAPICredentials(ctx)
	if err != nil {
		log.Fatalf("创建 API 凭证失败: %v", err)
	}
	fmt.Printf("API Key: %s\n", creds.APIKey)
	fmt.Printf("Passphrase: %s\n\n", creds.Passphrase)

	// 2. 查询余额
	fmt.Println("=== 查询 USDC 余额 ===")
	balance, err := sdk.Trading.GetCollateralBalance(ctx)
	if err != nil {
		log.Fatalf("查询余额失败: %v", err)
	}
	fmt.Printf("USDC 余额: %s\n", balance.Balance)
	fmt.Printf("USDC 授权: %s\n\n", balance.Allowance)

	// 3. 获取活跃市场
	fmt.Println("=== 获取市场用于交易 ===")
	markets, err := sdk.Markets.GetActiveMarkets(ctx, 1)
	if err != nil {
		log.Fatalf("获取市场失败: %v", err)
	}

	if len(markets) == 0 {
		log.Fatal("没有找到活跃市场")
	}

	market := markets[0]
	fmt.Printf("选择市场: %s\n", market.Question)

	tokenIDs := market.GetClobTokenIDs()
	if len(tokenIDs) < 2 {
		log.Fatal("市场没有足够的 token")
	}

	yesTokenID := tokenIDs[0]
	fmt.Printf("YES Token ID: %s\n\n", yesTokenID)

	// 4. 获取当前价格
	fmt.Println("=== 获取当前价格 ===")
	priceInfo, err := sdk.Trading.GetPrice(ctx, yesTokenID)
	if err != nil {
		log.Printf("获取价格失败: %v\n", err)
	} else {
		fmt.Printf("当前价格: %s\n\n", priceInfo.Price)
	}

	// 5. 获取 tick size
	fmt.Println("=== 获取最小价格单位 ===")
	tickSize, err := sdk.Trading.GetTickSize(ctx, yesTokenID)
	if err != nil {
		log.Printf("获取 tick size 失败: %v\n", err)
	} else {
		fmt.Printf("Tick Size: %s\n\n", tickSize.TickSize)
	}

	// 6. 查询活跃订单
	fmt.Println("=== 查询活跃订单 ===")
	orders, err := sdk.Trading.GetOpenOrders(ctx)
	if err != nil {
		log.Printf("查询订单失败: %v\n", err)
	} else {
		fmt.Printf("活跃订单数量: %d\n", len(orders))
		for _, order := range orders {
			fmt.Printf("  - %s: %s @ %s (状态: %s)\n",
				order.Side, order.OriginalSize, order.Price, order.Status)
		}
	}
	fmt.Println()

	// 7. 创建订单（演示，实际不会执行）
	fmt.Println("=== 创建订单演示 ===")
	fmt.Println("注意: 以下是订单创建示例，取消注释后可执行")
	fmt.Println()

	// 示例：创建一个限价买单
	// 买入 10 shares，价格 0.01 USDC
	orderReq := &clob.CreateOrderRequest{
		TokenID:   yesTokenID,
		Side:      clob.OrderSideBuy,
		Price:     decimal.NewFromFloat(0.01), // 非常低的价格，不会成交
		Size:      decimal.NewFromInt(10),
		Type:      clob.OrderTypeGTC,
		IsNegRisk: market.IsNegRisk(),
	}

	fmt.Printf("准备创建订单:\n")
	fmt.Printf("  Token ID: %s\n", orderReq.TokenID)
	fmt.Printf("  方向: %s\n", orderReq.Side)
	fmt.Printf("  价格: %s\n", orderReq.Price)
	fmt.Printf("  数量: %s\n", orderReq.Size)
	fmt.Printf("  类型: %s\n", orderReq.Type)
	fmt.Printf("  NegRisk: %t\n", orderReq.IsNegRisk)
	fmt.Println()

	// 取消下面的注释以实际创建订单
	/*
		resp, err := sdk.Trading.CreateOrder(ctx, orderReq)
		if err != nil {
			log.Fatalf("创建订单失败: %v", err)
		}
		fmt.Printf("订单创建成功!\n")
		fmt.Printf("  Order ID: %s\n", resp.OrderID)
		fmt.Printf("  Status: %s\n", resp.Status)

		// 取消订单
		if resp.OrderID != "" {
			err = sdk.Trading.CancelOrder(ctx, resp.OrderID)
			if err != nil {
				log.Printf("取消订单失败: %v", err)
			} else {
				fmt.Println("订单已取消")
			}
		}
	*/

	// 8. 获取交易历史
	fmt.Println("=== 获取交易历史 ===")
	trades, err := sdk.Trading.GetRecentTrades(ctx, 5)
	if err != nil {
		log.Printf("获取交易历史失败: %v\n", err)
	} else {
		fmt.Printf("最近交易数量: %d\n", len(trades))
		for _, trade := range trades {
			fmt.Printf("  - %s: %s @ %s (%s)\n",
				trade.Side, trade.Size, trade.Price, trade.MatchTime)
		}
	}

	fmt.Println("\n示例完成!")
}
