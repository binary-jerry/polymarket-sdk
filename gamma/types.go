package gamma

import (
	"encoding/json"
	"time"

	"github.com/shopspring/decimal"
)

// Market 市场信息
type Market struct {
	// 基础信息
	ID          string `json:"id"`
	Question    string `json:"question"`
	Description string `json:"description"`
	Slug        string `json:"slug"`
	ConditionID string `json:"conditionId"`

	// Token 信息
	Tokens        []Token `json:"tokens"`
	Outcomes      string  `json:"outcomes"`      // "Yes,No" 格式
	OutcomePrices string  `json:"outcomePrices"` // "0.5,0.5" 格式
	ClobTokenIds  string  `json:"clobTokenIds"`  // token IDs 逗号分隔

	// 状态
	Active          bool `json:"active"`
	Closed          bool `json:"closed"`
	Archived        bool `json:"archived"`
	AcceptingOrders bool `json:"acceptingOrders"`
	EnableOrderBook bool `json:"enableOrderBook"`
	New             bool `json:"new"`
	Featured        bool `json:"featured"`

	// 价格和流动性
	Volume       string  `json:"volume"`
	Liquidity    string  `json:"liquidity"`
	Volume24hr   float64 `json:"volume24hr"`
	VolumeNum    float64 `json:"volumeNum"`
	LiquidityNum float64 `json:"liquidityNum"`

	// 价格变动
	OneDayPriceChange  float64 `json:"oneDayPriceChange"`
	OneHourPriceChange float64 `json:"oneHourPriceChange"`
	OneWeekPriceChange float64 `json:"oneWeekPriceChange"`

	// 订单配置
	OrderPriceMinTickSize float64 `json:"orderPriceMinTickSize"`
	OrderMinSize          float64 `json:"orderMinSize"`

	// 时间
	EndDate      string `json:"endDate"`
	EndDateIso   string `json:"endDateIso"`
	StartDate    string `json:"startDate,omitempty"`
	StartDateIso string `json:"startDateIso,omitempty"`
	CreatedAt    string `json:"createdAt,omitempty"`
	UpdatedAt    string `json:"updatedAt,omitempty"`

	// 分类
	Category string `json:"category"`
	Tags     []Tag  `json:"tags,omitempty"`

	// 市场类型
	MarketType       string `json:"marketType"` // "binary" 等
	NegRisk          bool   `json:"negRisk"`
	NegRiskMarketID  string `json:"negRiskMarketId,omitempty"`
	NegRiskRequestID string `json:"negRiskRequestId,omitempty"`

	// 解析相关
	ResolutionSource string `json:"resolutionSource"`
	ResolvedBy       string `json:"resolvedBy,omitempty"`

	// 图片
	Image            string `json:"image,omitempty"`
	Icon             string `json:"icon,omitempty"`
	TwitterCardImage string `json:"twitterCardImage,omitempty"`

	// 其他
	Fee            string  `json:"fee,omitempty"`
	Spread         float64 `json:"spread,omitempty"`
	BestBid        float64 `json:"bestBid,omitempty"`
	BestAsk        float64 `json:"bestAsk,omitempty"`
	LastTradePrice float64 `json:"lastTradePrice,omitempty"`
}

// Token 代币信息
type Token struct {
	TokenID string `json:"token_id"`
	Outcome string `json:"outcome"` // "Yes" 或 "No"
	Price   string `json:"price,omitempty"`
	Winner  bool   `json:"winner"`
}

// Tag 标签信息
type Tag struct {
	ID        string `json:"id"`
	Label     string `json:"label"`
	Slug      string `json:"slug"`
	ForceShow bool   `json:"forceShow"`
	ForceHide bool   `json:"forceHide"`
}

// MarketListParams 市场列表查询参数
type MarketListParams struct {
	Limit  int `url:"limit,omitempty"`
	Offset int `url:"offset,omitempty"`

	// 筛选条件
	Active   *bool `url:"active,omitempty"`
	Closed   *bool `url:"closed,omitempty"`
	Archived *bool `url:"archived,omitempty"`
	New      *bool `url:"new,omitempty"`
	Featured *bool `url:"featured,omitempty"`
	NegRisk  *bool `url:"neg_risk,omitempty"`

	// 分类筛选
	Slug     string `url:"slug,omitempty"`     // 市场 slug
	TagSlug  string `url:"tag_slug,omitempty"` // 标签 slug
	Category string `url:"category,omitempty"`

	// 排序
	Order     string `url:"order,omitempty"` // volume, liquidity, end_date_min, created_at
	Ascending bool   `url:"ascending,omitempty"`
	TagId     int    `url:"tag_id,omitempty"`

	// 时间筛选
	StartDateMin string `url:"start_date_min,omitempty"`
	StartDateMax string `url:"start_date_max,omitempty"`
	EndDateMin   string `url:"end_date_min,omitempty"`
	EndDateMax   string `url:"end_date_max,omitempty"`

	// 搜索
	TextQuery string `url:"text_query,omitempty"` // 文本搜索
	Id        int    `url:"id,omitempty"`
}

// MarketListResponse 市场列表响应
type MarketListResponse struct {
	Data       []Market `json:"data,omitempty"`
	NextCursor string   `json:"next_cursor,omitempty"`
	Limit      int      `json:"limit,omitempty"`
	Count      int      `json:"count,omitempty"`
}

// GetOutcomePrices 解析 outcomePrices 字符串
func (m *Market) GetOutcomePrices() ([]decimal.Decimal, error) {
	if m.OutcomePrices == "" {
		return nil, nil
	}

	var prices []string
	if err := json.Unmarshal([]byte(m.OutcomePrices), &prices); err != nil {
		// 尝试解析为逗号分隔格式
		prices = splitString(m.OutcomePrices, ",")
	}

	result := make([]decimal.Decimal, 0, len(prices))
	for _, p := range prices {
		d, err := decimal.NewFromString(p)
		if err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, nil
}

// GetClobTokenIDs 解析 clobTokenIds 字符串
func (m *Market) GetClobTokenIDs() []string {
	if m.ClobTokenIds == "" {
		return nil
	}

	var ids []string
	if err := json.Unmarshal([]byte(m.ClobTokenIds), &ids); err != nil {
		// 尝试解析为逗号分隔格式
		ids = splitString(m.ClobTokenIds, ",")
	}
	return ids
}

// GetEndDate 解析结束日期
func (m *Market) GetEndDate() (time.Time, error) {
	if m.EndDateIso != "" {
		return time.Parse(time.RFC3339, m.EndDateIso)
	}
	if m.EndDate != "" {
		return time.Parse(time.RFC3339, m.EndDate)
	}
	return time.Time{}, nil
}

// IsActive 判断市场是否活跃
func (m *Market) IsActive() bool {
	return m.Active && !m.Closed && !m.Archived
}

// IsNegRisk 判断是否为 NegRisk 市场
func (m *Market) IsNegRisk() bool {
	return m.NegRisk
}

// GetYesToken 获取 YES token
func (m *Market) GetYesToken() *Token {
	for i := range m.Tokens {
		if m.Tokens[i].Outcome == "Yes" {
			return &m.Tokens[i]
		}
	}
	return nil
}

// GetNoToken 获取 NO token
func (m *Market) GetNoToken() *Token {
	for i := range m.Tokens {
		if m.Tokens[i].Outcome == "No" {
			return &m.Tokens[i]
		}
	}
	return nil
}

// splitString 分割字符串
func splitString(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := make([]string, 0)
	for _, part := range split(s, sep) {
		if trimmed := trim(part); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func split(s, sep string) []string {
	if s == "" {
		return nil
	}
	result := make([]string, 0, 8)
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, s[start:])
	return result
}

func trim(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

// BoolPtr 返回 bool 指针
func BoolPtr(b bool) *bool {
	return &b
}
