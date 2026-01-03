package clob

import (
	"context"
	"fmt"
)

// GetBalanceAllowance 获取余额和授权
func (c *Client) GetBalanceAllowance(ctx context.Context, params *BalanceAllowanceParams) (*BalanceAllowance, error) {
	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("GET", "/balance-allowance", "")
	if err != nil {
		return nil, err
	}

	var result BalanceAllowance
	err = c.httpClient.DoWithAuthAndParams(ctx, "GET", "/balance-allowance", params, nil, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get balance allowance: %w", err)
	}

	return &result, nil
}

// GetCollateralBalance 获取抵押品余额 (USDC)
func (c *Client) GetCollateralBalance(ctx context.Context) (*BalanceAllowance, error) {
	params := &BalanceAllowanceParams{
		AssetType: AssetTypeCollateral,
	}
	return c.GetBalanceAllowance(ctx, params)
}

// GetConditionalBalance 获取条件代币余额
func (c *Client) GetConditionalBalance(ctx context.Context, tokenID string) (*BalanceAllowance, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}

	params := &BalanceAllowanceParams{
		AssetType: AssetTypeConditional,
		TokenID:   tokenID,
	}
	return c.GetBalanceAllowance(ctx, params)
}

// GetTickSize 获取价格最小变动单位
func (c *Client) GetTickSize(ctx context.Context, tokenID string) (*TickSize, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}

	path := "/tick-size"
	params := struct {
		TokenID string `url:"token_id"`
	}{
		TokenID: tokenID,
	}

	var result TickSize
	err := c.httpClient.Get(ctx, path, params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get tick size: %w", err)
	}

	return &result, nil
}

// GetPrice 获取当前价格
func (c *Client) GetPrice(ctx context.Context, tokenID string) (*PriceInfo, error) {
	if tokenID == "" {
		return nil, fmt.Errorf("token ID is required")
	}

	path := "/price"
	params := struct {
		TokenID string `url:"token_id"`
	}{
		TokenID: tokenID,
	}

	var result PriceInfo
	err := c.httpClient.Get(ctx, path, params, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get price: %w", err)
	}

	return &result, nil
}

// GetPrices 批量获取价格
func (c *Client) GetPrices(ctx context.Context, tokenIDs []string) ([]*PriceInfo, error) {
	if len(tokenIDs) == 0 {
		return nil, nil
	}

	results := make([]*PriceInfo, 0, len(tokenIDs))
	for _, tokenID := range tokenIDs {
		price, err := c.GetPrice(ctx, tokenID)
		if err != nil {
			return nil, fmt.Errorf("failed to get price for %s: %w", tokenID, err)
		}
		results = append(results, price)
	}

	return results, nil
}
