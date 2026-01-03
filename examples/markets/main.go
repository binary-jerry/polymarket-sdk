package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/binary-jerry/polymarket-sdk/gamma"
)

func main() {
	// 创建 Gamma 客户端（无需认证）
	client := gamma.NewClient(nil)
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 1. 获取活跃市场
	fmt.Println("=== 获取活跃市场 ===")
	activeMarkets, err := client.GetActiveMarkets(ctx, 10)
	if err != nil {
		log.Fatalf("获取活跃市场失败: %v", err)
	}

	for i, market := range activeMarkets {
		fmt.Printf("%d. %s\n", i+1, market.Question)
		fmt.Printf("   Slug: %s\n", market.Slug)
		fmt.Printf("   Volume: %s\n", market.Volume)
		fmt.Printf("   Liquidity: %s\n", market.Liquidity)

		// 获取 token IDs
		tokenIDs := market.GetClobTokenIDs()
		if len(tokenIDs) > 0 {
			fmt.Printf("   Token IDs: %v\n", tokenIDs)
		}

		// 获取 YES/NO token
		yesToken := market.GetYesToken()
		noToken := market.GetNoToken()
		if yesToken != nil && noToken != nil {
			fmt.Printf("   YES Token: %s\n", yesToken.TokenID)
			fmt.Printf("   NO Token: %s\n", noToken.TokenID)
		}
		fmt.Println()
	}

	// 2. 获取精选市场
	fmt.Println("=== 获取精选市场 ===")
	featuredMarkets, err := client.GetFeaturedMarkets(ctx, 5)
	if err != nil {
		log.Fatalf("获取精选市场失败: %v", err)
	}

	for i, market := range featuredMarkets {
		fmt.Printf("%d. %s (Volume: %s)\n", i+1, market.Question, market.Volume)
	}
	fmt.Println()

	// 3. 搜索市场
	fmt.Println("=== 搜索市场: 'Bitcoin' ===")
	searchResults, err := client.SearchMarkets(ctx, "Bitcoin", 5)
	if err != nil {
		log.Fatalf("搜索市场失败: %v", err)
	}

	for i, market := range searchResults {
		fmt.Printf("%d. %s\n", i+1, market.Question)
	}
	fmt.Println()

	// 4. 获取单个市场详情
	if len(activeMarkets) > 0 {
		fmt.Println("=== 获取市场详情 ===")
		market, err := client.GetMarketBySlug(ctx, activeMarkets[0].Slug)
		if err != nil {
			log.Fatalf("获取市场详情失败: %v", err)
		}

		fmt.Printf("Question: %s\n", market.Question)
		fmt.Printf("Description: %s\n", market.Description)
		fmt.Printf("End Date: %s\n", market.EndDateIso)
		fmt.Printf("NegRisk: %t\n", market.IsNegRisk())

		// 解析价格
		prices, err := market.GetOutcomePrices()
		if err == nil && len(prices) >= 2 {
			fmt.Printf("YES Price: %s\n", prices[0])
			fmt.Printf("NO Price: %s\n", prices[1])
		}
	}

	// 5. 获取 NegRisk 市场
	fmt.Println("\n=== 获取 NegRisk 市场 ===")
	negRiskMarkets, err := client.GetNegRiskMarkets(ctx, 5)
	if err != nil {
		log.Fatalf("获取 NegRisk 市场失败: %v", err)
	}

	for i, market := range negRiskMarkets {
		fmt.Printf("%d. [NegRisk] %s\n", i+1, market.Question)
	}

	fmt.Println("\n示例完成!")
}
