package clob

import (
	"context"
	"fmt"
)

// GetTrades 获取交易历史
func (c *Client) GetTrades(ctx context.Context, params *TradesQueryParams) ([]*Trade, error) {
	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	if params == nil {
		params = &TradesQueryParams{}
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("GET", "/trades", "")
	if err != nil {
		return nil, err
	}

	var result []*Trade
	err = c.httpClient.DoWithAuthAndParams(ctx, "GET", "/trades", params, nil, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	return result, nil
}

// GetTradesByMarket 获取指定市场的交易
func (c *Client) GetTradesByMarket(ctx context.Context, marketID string, limit int) ([]*Trade, error) {
	if marketID == "" {
		return nil, fmt.Errorf("market ID is required")
	}

	if limit <= 0 {
		limit = 100
	}

	params := &TradesQueryParams{
		Market: marketID,
		Limit:  limit,
	}

	return c.GetTrades(ctx, params)
}

// GetTradesByAsset 获取指定资产的交易
func (c *Client) GetTradesByAsset(ctx context.Context, assetID string, limit int) ([]*Trade, error) {
	if assetID == "" {
		return nil, fmt.Errorf("asset ID is required")
	}

	if limit <= 0 {
		limit = 100
	}

	params := &TradesQueryParams{
		AssetID: assetID,
		Limit:   limit,
	}

	return c.GetTrades(ctx, params)
}

// GetRecentTrades 获取最近的交易
func (c *Client) GetRecentTrades(ctx context.Context, limit int) ([]*Trade, error) {
	if limit <= 0 {
		limit = 50
	}

	params := &TradesQueryParams{
		Limit: limit,
	}

	return c.GetTrades(ctx, params)
}

// GetTradesByTimeRange 获取指定时间范围的交易
func (c *Client) GetTradesByTimeRange(ctx context.Context, after, before string, limit int) ([]*Trade, error) {
	if limit <= 0 {
		limit = 100
	}

	params := &TradesQueryParams{
		After:  after,
		Before: before,
		Limit:  limit,
	}

	return c.GetTrades(ctx, params)
}
