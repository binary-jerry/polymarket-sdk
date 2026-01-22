package clob

import (
	"context"
	"fmt"
)

const (
	// EndCursor 分页结束标识
	EndCursor = "LTE="
	// DefaultCursor 默认游标
	DefaultCursor = "MA=="
)

// TradesResponse 交易历史分页响应
type TradesResponse struct {
	NextCursor string   `json:"next_cursor"`
	Data       []*Trade `json:"data"`
}

// tradesQueryParamsWithCursor 带游标的交易查询参数
type tradesQueryParamsWithCursor struct {
	Market     string `url:"market,omitempty"`
	AssetID    string `url:"asset_id,omitempty"`
	Maker      string `url:"maker,omitempty"`
	Before     string `url:"before,omitempty"`
	After      string `url:"after,omitempty"`
	Limit      int    `url:"limit,omitempty"`
	NextCursor string `url:"next_cursor,omitempty"`
}

// GetTrades 获取交易历史 (自动分页获取所有记录)
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

	var allTrades []*Trade
	nextCursor := DefaultCursor

	for nextCursor != EndCursor {
		// 构建带游标的查询参数
		queryParams := &tradesQueryParamsWithCursor{
			Market:     params.Market,
			AssetID:    params.AssetID,
			Maker:      params.Maker,
			Before:     params.Before,
			After:      params.After,
			Limit:      params.Limit,
			NextCursor: nextCursor,
		}

		var resp TradesResponse
		err = c.httpClient.DoWithAuthAndParams(ctx, "GET", "/trades", queryParams, nil, authHeaders, &resp)
		if err != nil {
			return nil, fmt.Errorf("failed to get trades: %w", err)
		}

		allTrades = append(allTrades, resp.Data...)
		nextCursor = resp.NextCursor

		// 如果指定了 Limit 且已获取足够数据，停止分页
		if params.Limit > 0 && len(allTrades) >= params.Limit {
			break
		}
	}

	// 如果指定了 Limit，截断结果
	if params.Limit > 0 && len(allTrades) > params.Limit {
		allTrades = allTrades[:params.Limit]
	}

	return allTrades, nil
}

// GetTradesPage 获取单页交易历史 (用于手动分页)
func (c *Client) GetTradesPage(ctx context.Context, params *TradesQueryParams, cursor string) (*TradesResponse, error) {
	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	if params == nil {
		params = &TradesQueryParams{}
	}

	if cursor == "" {
		cursor = DefaultCursor
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("GET", "/trades", "")
	if err != nil {
		return nil, err
	}

	// 构建带游标的查询参数
	queryParams := &tradesQueryParamsWithCursor{
		Market:     params.Market,
		AssetID:    params.AssetID,
		Maker:      params.Maker,
		Before:     params.Before,
		After:      params.After,
		Limit:      params.Limit,
		NextCursor: cursor,
	}

	var resp TradesResponse
	err = c.httpClient.DoWithAuthAndParams(ctx, "GET", "/trades", queryParams, nil, authHeaders, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to get trades: %w", err)
	}

	return &resp, nil
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
