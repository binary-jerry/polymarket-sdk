package gamma

import (
	"context"
	"fmt"
)

// GetMarkets 获取市场列表
func (c *Client) GetMarkets(ctx context.Context, params *MarketListParams) (*MarketListResponse, error) {
	if params == nil {
		params = &MarketListParams{}
	}

	var result []Market
	err := c.httpClient.Get(ctx, "/markets", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get markets: %w", err)
	}

	return &MarketListResponse{
		Data:  result,
		Count: len(result),
	}, nil
}

// GetAllMarkets 获取所有市场（自动分页）
func (c *Client) GetAllMarkets(ctx context.Context, params *MarketListParams) ([]Market, error) {
	if params == nil {
		params = &MarketListParams{}
	}

	if params.Limit == 0 {
		params.Limit = 100
	}

	var allMarkets []Market
	offset := 0

	for {
		params.Offset = offset
		resp, err := c.GetMarkets(ctx, params)
		if err != nil {
			return allMarkets, err
		}

		if len(resp.Data) == 0 {
			break
		}

		allMarkets = append(allMarkets, resp.Data...)

		if len(resp.Data) < params.Limit {
			break
		}

		offset += params.Limit

		// 安全限制：最多获取 10000 个市场
		if len(allMarkets) >= 10000 {
			break
		}
	}

	return allMarkets, nil
}

// GetMarket 获取单个市场
func (c *Client) GetMarket(ctx context.Context, marketID string) (*Market, error) {
	if marketID == "" {
		return nil, fmt.Errorf("market ID is required")
	}

	var result Market
	err := c.httpClient.Get(ctx, "/markets/"+marketID, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get market %s: %w", marketID, err)
	}

	return &result, nil
}

// GetMarketBySlug 通过 slug 获取市场
func (c *Client) GetMarketBySlug(ctx context.Context, slug string) (*Market, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug is required")
	}

	var result Market
	err := c.httpClient.Get(ctx, "/markets/slug/"+slug, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get market by slug %s: %w", slug, err)
	}

	return &result, nil
}

// GetMarketByConditionID 通过 conditionID 获取市场
func (c *Client) GetMarketByConditionID(ctx context.Context, conditionID string) (*Market, error) {
	if conditionID == "" {
		return nil, fmt.Errorf("condition ID is required")
	}

	params := &MarketListParams{
		Limit: 1,
	}

	var result []Market
	err := c.httpClient.Get(ctx, "/markets", params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get market by condition ID %s: %w", conditionID, err)
	}

	// 遍历查找匹配的 conditionID
	for _, m := range result {
		if m.ConditionID == conditionID {
			return &m, nil
		}
	}

	return nil, fmt.Errorf("market with condition ID %s not found", conditionID)
}

// GetActiveMarkets 获取活跃市场
func (c *Client) GetActiveMarkets(ctx context.Context, limit int) ([]Market, error) {
	if limit <= 0 {
		limit = 100
	}

	params := &MarketListParams{
		Limit:  limit,
		Active: BoolPtr(true),
		Closed: BoolPtr(false),
		Order:  "volume",
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetFeaturedMarkets 获取精选市场
func (c *Client) GetFeaturedMarkets(ctx context.Context, limit int) ([]Market, error) {
	if limit <= 0 {
		limit = 20
	}

	params := &MarketListParams{
		Limit:    limit,
		Featured: BoolPtr(true),
		Active:   BoolPtr(true),
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetNegRiskMarkets 获取 NegRisk 市场
func (c *Client) GetNegRiskMarkets(ctx context.Context, limit int) ([]Market, error) {
	if limit <= 0 {
		limit = 100
	}

	params := &MarketListParams{
		Limit:   limit,
		NegRisk: BoolPtr(true),
		Active:  BoolPtr(true),
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// SearchMarkets 搜索市场
func (c *Client) SearchMarkets(ctx context.Context, query string, limit int) ([]Market, error) {
	if query == "" {
		return nil, fmt.Errorf("search query is required")
	}

	if limit <= 0 {
		limit = 50
	}

	params := &MarketListParams{
		Limit:     limit,
		TextQuery: query,
		Active:    BoolPtr(true),
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetMarketsByCategory 按分类获取市场
func (c *Client) GetMarketsByCategory(ctx context.Context, category string, limit int) ([]Market, error) {
	if category == "" {
		return nil, fmt.Errorf("category is required")
	}

	if limit <= 0 {
		limit = 100
	}

	params := &MarketListParams{
		Limit:    limit,
		Category: category,
		Active:   BoolPtr(true),
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetMarketsByTag 按标签获取市场
func (c *Client) GetMarketsByTag(ctx context.Context, tagSlug string, limit int) ([]Market, error) {
	if tagSlug == "" {
		return nil, fmt.Errorf("tag slug is required")
	}

	if limit <= 0 {
		limit = 100
	}

	params := &MarketListParams{
		Limit:   limit,
		TagSlug: tagSlug,
		Active:  BoolPtr(true),
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetTopVolumeMarkets 获取交易量最高的市场
func (c *Client) GetTopVolumeMarkets(ctx context.Context, limit int) ([]Market, error) {
	if limit <= 0 {
		limit = 20
	}

	params := &MarketListParams{
		Limit:  limit,
		Active: BoolPtr(true),
		Order:  "volume",
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}

// GetEndingSoonMarkets 获取即将结束的市场
func (c *Client) GetEndingSoonMarkets(ctx context.Context, limit int) ([]Market, error) {
	if limit <= 0 {
		limit = 20
	}

	params := &MarketListParams{
		Limit:     limit,
		Active:    BoolPtr(true),
		Order:     "end_date_min",
		Ascending: true,
	}

	resp, err := c.GetMarkets(ctx, params)
	if err != nil {
		return nil, err
	}

	return resp.Data, nil
}
