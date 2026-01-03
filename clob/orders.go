package clob

import (
	"context"
	"encoding/json"
	"fmt"
)

// CreateOrder 创建订单
func (c *Client) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*OrderResponse, error) {
	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	// 创建已签名订单
	signedOrder, err := c.orderSigner.CreateSignedOrder(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create signed order: %w", err)
	}

	// 确定订单类型
	orderType := req.Type
	if orderType == "" {
		orderType = OrderTypeGTC
	}

	// 构建提交请求
	postReq := &PostOrderRequest{
		Order:     signedOrder,
		Owner:     c.GetAddress(),
		OrderType: orderType,
	}

	// 序列化请求体
	bodyBytes, err := json.Marshal(postReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("POST", "/order", string(bodyBytes))
	if err != nil {
		return nil, err
	}

	// 发送请求
	var result OrderResponse
	err = c.httpClient.DoWithAuth(ctx, "POST", "/order", postReq, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	return &result, nil
}

// CreateOrders 批量创建订单
func (c *Client) CreateOrders(ctx context.Context, reqs []*CreateOrderRequest) ([]*OrderResponse, error) {
	if len(reqs) == 0 {
		return nil, nil
	}

	if len(reqs) > 15 {
		return nil, fmt.Errorf("maximum 15 orders per batch, got %d", len(reqs))
	}

	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	// 创建已签名订单
	postReqs := make([]*PostOrderRequest, 0, len(reqs))
	for _, req := range reqs {
		signedOrder, err := c.orderSigner.CreateSignedOrder(req)
		if err != nil {
			return nil, fmt.Errorf("failed to create signed order: %w", err)
		}

		orderType := req.Type
		if orderType == "" {
			orderType = OrderTypeGTC
		}

		postReqs = append(postReqs, &PostOrderRequest{
			Order:     signedOrder,
			Owner:     c.GetAddress(),
			OrderType: orderType,
		})
	}

	// 序列化请求体
	bodyBytes, err := json.Marshal(postReqs)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("POST", "/orders", string(bodyBytes))
	if err != nil {
		return nil, err
	}

	// 发送请求
	var results []*OrderResponse
	err = c.httpClient.DoWithAuth(ctx, "POST", "/orders", postReqs, authHeaders, &results)
	if err != nil {
		return nil, fmt.Errorf("failed to create orders: %w", err)
	}

	return results, nil
}

// GetOrder 查询订单
func (c *Client) GetOrder(ctx context.Context, orderID string) (*Order, error) {
	if orderID == "" {
		return nil, fmt.Errorf("order ID is required")
	}

	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	path := "/order/" + orderID

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("GET", path, "")
	if err != nil {
		return nil, err
	}

	var result Order
	err = c.httpClient.DoWithAuthAndParams(ctx, "GET", path, nil, nil, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &result, nil
}

// GetOrders 查询活跃订单
func (c *Client) GetOrders(ctx context.Context, params *OrdersQueryParams) ([]*Order, error) {
	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	if params == nil {
		params = &OrdersQueryParams{}
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("GET", "/orders", "")
	if err != nil {
		return nil, err
	}

	var result []*Order
	err = c.httpClient.DoWithAuthAndParams(ctx, "GET", "/orders", params, nil, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get orders: %w", err)
	}

	return result, nil
}

// GetOpenOrders 获取所有活跃订单
func (c *Client) GetOpenOrders(ctx context.Context) ([]*Order, error) {
	return c.GetOrders(ctx, nil)
}

// CancelOrder 取消单个订单
func (c *Client) CancelOrder(ctx context.Context, orderID string) error {
	if orderID == "" {
		return fmt.Errorf("order ID is required")
	}

	if err := c.ensureCredentials(ctx); err != nil {
		return fmt.Errorf("failed to ensure credentials: %w", err)
	}

	path := "/order/" + orderID

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("DELETE", path, "")
	if err != nil {
		return err
	}

	err = c.httpClient.DoWithAuth(ctx, "DELETE", path, nil, authHeaders, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel order: %w", err)
	}

	return nil
}

// CancelOrders 批量取消订单
func (c *Client) CancelOrders(ctx context.Context, orderIDs []string) (*CancelResponse, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}

	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	body := &BatchCancelRequest{
		OrderIDs: orderIDs,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("DELETE", "/orders", string(bodyBytes))
	if err != nil {
		return nil, err
	}

	var result CancelResponse
	err = c.httpClient.DoWithAuth(ctx, "DELETE", "/orders", body, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel orders: %w", err)
	}

	return &result, nil
}

// CancelOrdersByMarket 取消指定市场的所有订单
func (c *Client) CancelOrdersByMarket(ctx context.Context, marketID string) (*CancelResponse, error) {
	if marketID == "" {
		return nil, fmt.Errorf("market ID is required")
	}

	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	body := &BatchCancelRequest{
		Market: marketID,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("DELETE", "/orders", string(bodyBytes))
	if err != nil {
		return nil, err
	}

	var result CancelResponse
	err = c.httpClient.DoWithAuth(ctx, "DELETE", "/orders", body, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel orders by market: %w", err)
	}

	return &result, nil
}

// CancelOrdersByAsset 取消指定资产的所有订单
func (c *Client) CancelOrdersByAsset(ctx context.Context, assetID string) (*CancelResponse, error) {
	if assetID == "" {
		return nil, fmt.Errorf("asset ID is required")
	}

	if err := c.ensureCredentials(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure credentials: %w", err)
	}

	body := &BatchCancelRequest{
		AssetID: assetID,
	}

	bodyBytes, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("DELETE", "/orders", string(bodyBytes))
	if err != nil {
		return nil, err
	}

	var result CancelResponse
	err = c.httpClient.DoWithAuth(ctx, "DELETE", "/orders", body, authHeaders, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel orders by asset: %w", err)
	}

	return &result, nil
}

// CancelAllOrders 取消所有订单
func (c *Client) CancelAllOrders(ctx context.Context) error {
	if err := c.ensureCredentials(ctx); err != nil {
		return fmt.Errorf("failed to ensure credentials: %w", err)
	}

	// 获取认证头
	authHeaders, err := c.getL2AuthHeaders("DELETE", "/cancel-all", "")
	if err != nil {
		return err
	}

	err = c.httpClient.DoWithAuth(ctx, "DELETE", "/cancel-all", nil, authHeaders, nil)
	if err != nil {
		return fmt.Errorf("failed to cancel all orders: %w", err)
	}

	return nil
}
