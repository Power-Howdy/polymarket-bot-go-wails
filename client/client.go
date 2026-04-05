package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ─── Data models ─────────────────────────────────────────────────────────────

// StringList decodes Gamma fields that are either a JSON string array or a string
// containing a JSON array (e.g. `"[\"Yes\",\"No\"]"`).
type StringList []string

func (s *StringList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*s = nil
		return nil
	}
	if data[0] == '"' {
		var encoded string
		if err := json.Unmarshal(data, &encoded); err != nil {
			return err
		}
		if encoded == "" {
			*s = nil
			return nil
		}
		return json.Unmarshal([]byte(encoded), (*[]string)(s))
	}
	return json.Unmarshal(data, (*[]string)(s))
}

// FlexFloat64 decodes a number or a numeric string (Gamma often sends volume as a string).
type FlexFloat64 float64

func (f *FlexFloat64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*f = 0
		return nil
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		if str == "" {
			*f = 0
			return nil
		}
		v, err := strconv.ParseFloat(str, 64)
		if err != nil {
			return err
		}
		*f = FlexFloat64(v)
		return nil
	}
	var v float64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*f = FlexFloat64(v)
	return nil
}

// FlexInt64 decodes an int64 from a JSON number or numeric string (CLOB order book timestamps).
type FlexInt64 int64

func (i *FlexInt64) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*i = 0
		return nil
	}
	if data[0] == '"' {
		var str string
		if err := json.Unmarshal(data, &str); err != nil {
			return err
		}
		if str == "" {
			*i = 0
			return nil
		}
		v, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return err
		}
		*i = FlexInt64(v)
		return nil
	}
	var v int64
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*i = FlexInt64(v)
	return nil
}

// GammaMarket represents a market from the Gamma (discovery) API.
type GammaMarket struct {
	ID            string      `json:"id"`
	Question      string      `json:"question"`
	ConditionID   string      `json:"conditionId"`
	Slug          string      `json:"slug"`
	Active        bool        `json:"active"`
	Closed        bool        `json:"closed"`
	EndDate       string      `json:"endDate"`
	OutcomeTokens StringList  `json:"outcomes"`
	// Prices encoded as JSON string, e.g. "[\"0.65\",\"0.35\"]"
	OutcomePrices string      `json:"outcomePrices"`
	Volume        FlexFloat64 `json:"volume"`
	Liquidity     float64     `json:"liquidityNum"`
	Tags          []Tag       `json:"tags"`
}

// Prices parses the OutcomePrices string into floats.
func (m GammaMarket) Prices() ([]float64, error) {
	var raw []string
	if err := json.Unmarshal([]byte(m.OutcomePrices), &raw); err != nil {
		return nil, err
	}
	out := make([]float64, len(raw))
	for i, s := range raw {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, err
		}
		out[i] = f
	}
	return out, nil
}

// Tag represents a market category tag.
type Tag struct {
	ID    int    `json:"id"`
	Label string `json:"label"`
	Slug  string `json:"slug"`
}

// Token represents a single outcome token.
type Token struct {
	TokenID  string `json:"token_id"`
	Outcome  string `json:"outcome"`
	Price    float64 `json:"price"`
	Winner   bool   `json:"winner"`
}

// CLOBMarket is a market returned from the CLOB API.
type CLOBMarket struct {
	ConditionID string  `json:"condition_id"`
	QuestionID  string  `json:"question_id"`
	Question    string  `json:"question"`
	Tokens      []Token `json:"tokens"`
	Active      bool    `json:"active"`
	Closed      bool    `json:"closed"`
	MinTickSize float64 `json:"minimum_tick_size"`
	MinOrderSize float64 `json:"minimum_order_size"`
}

// OrderBook represents bids and asks for a token.
type OrderBook struct {
	Market    string       `json:"market"`
	AssetID   string       `json:"asset_id"`
	Bids      []PriceLevel `json:"bids"`
	Asks      []PriceLevel `json:"asks"`
	Timestamp FlexInt64    `json:"timestamp"`
}

// PriceLevel is a single price/size entry in the order book.
type PriceLevel struct {
	Price string `json:"price"`
	Size  string `json:"size"`
}

// BestBid returns the best bid as float64.
func (o *OrderBook) BestBid() float64 {
	if len(o.Bids) == 0 {
		return 0
	}
	v, _ := strconv.ParseFloat(o.Bids[0].Price, 64)
	return v
}

// BestAsk returns the best ask as float64.
func (o *OrderBook) BestAsk() float64 {
	if len(o.Asks) == 0 {
		return 1
	}
	v, _ := strconv.ParseFloat(o.Asks[0].Price, 64)
	return v
}

// MidPrice returns the mid-point between best bid and ask.
func (o *OrderBook) MidPrice() float64 {
	return (o.BestBid() + o.BestAsk()) / 2.0
}

// Spread returns the bid-ask spread as a fraction.
func (o *OrderBook) Spread() float64 {
	if o.BestAsk() == 0 {
		return 0
	}
	return o.BestAsk() - o.BestBid()
}

// OrderRequest is a new order to submit.
type OrderRequest struct {
	TokenID   string  `json:"token_id"`
	Price     float64 `json:"price"`
	Size      float64 `json:"size"`
	Side      string  `json:"side"` // "BUY" or "SELL"
	OrderType string  `json:"type"` // "GTC", "FOK", "GTD"
	Expiry    int64   `json:"expiration,omitempty"`
}

// OrderResponse is the API response for a submitted order.
type OrderResponse struct {
	Success      bool   `json:"success"`
	ErrorMsg     string `json:"errorMsg"`
	OrderID      string `json:"orderID"`
	TransactTime string `json:"transactTime"`
}

// ─── Client ───────────────────────────────────────────────────────────────────

// Client handles all Polymarket API communication.
type Client struct {
	clobHost   string
	gammaHost  string
	apiKey     string
	apiSecret  string
	passphrase string
	httpClient *http.Client
}

// New creates a new API client.
func New(clobHost, gammaHost, apiKey, apiSecret, passphrase string) *Client {
	return &Client{
		clobHost:   strings.TrimRight(clobHost, "/"),
		gammaHost:  strings.TrimRight(gammaHost, "/"),
		apiKey:     apiKey,
		apiSecret:  apiSecret,
		passphrase: passphrase,
		httpClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// ─── Gamma API ────────────────────────────────────────────────────────────────

// GetMarkets fetches active markets from the Gamma API with optional keyword filter.
func (c *Client) GetMarkets(keywords []string, limit int) ([]GammaMarket, error) {
	params := url.Values{}
	params.Set("active", "true")
	params.Set("closed", "false")
	params.Set("limit", strconv.Itoa(limit))
	if len(keywords) > 0 {
		params.Set("_c", strings.Join(keywords, ","))
	}

	endpoint := fmt.Sprintf("%s/markets?%s", c.gammaHost, params.Encode())
	var markets []GammaMarket
	if err := c.get(endpoint, false, &markets); err != nil {
		return nil, fmt.Errorf("get markets: %w", err)
	}
	return markets, nil
}

// GetMarketBySlug fetches a single market by slug.
func (c *Client) GetMarketBySlug(slug string) (*GammaMarket, error) {
	endpoint := fmt.Sprintf("%s/markets?slug=%s", c.gammaHost, url.QueryEscape(slug))
	var markets []GammaMarket
	if err := c.get(endpoint, false, &markets); err != nil {
		return nil, err
	}
	if len(markets) == 0 {
		return nil, fmt.Errorf("market not found: %s", slug)
	}
	return &markets[0], nil
}

// ─── CLOB API ─────────────────────────────────────────────────────────────────

// GetCLOBMarket fetches market info from the CLOB API by condition ID.
func (c *Client) GetCLOBMarket(conditionID string) (*CLOBMarket, error) {
	endpoint := fmt.Sprintf("%s/markets/%s", c.clobHost, conditionID)
	var m CLOBMarket
	if err := c.get(endpoint, false, &m); err != nil {
		return nil, fmt.Errorf("get clob market: %w", err)
	}
	return &m, nil
}

// GetOrderBook fetches the order book for a token.
func (c *Client) GetOrderBook(tokenID string) (*OrderBook, error) {
	endpoint := fmt.Sprintf("%s/book?token_id=%s", c.clobHost, url.QueryEscape(tokenID))
	var ob OrderBook
	if err := c.get(endpoint, false, &ob); err != nil {
		return nil, fmt.Errorf("get order book: %w", err)
	}
	return &ob, nil
}

// GetMidpointPrice fetches the mid-point price for a token.
func (c *Client) GetMidpointPrice(tokenID string) (float64, error) {
	endpoint := fmt.Sprintf("%s/midpoint?token_id=%s", c.clobHost, url.QueryEscape(tokenID))
	var resp struct {
		Mid string `json:"mid"`
	}
	if err := c.get(endpoint, false, &resp); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(resp.Mid, 64)
}

// GetSpread fetches the current spread for a token.
func (c *Client) GetSpread(tokenID string) (float64, error) {
	endpoint := fmt.Sprintf("%s/spread?token_id=%s", c.clobHost, url.QueryEscape(tokenID))
	var resp struct {
		Spread string `json:"spread"`
	}
	if err := c.get(endpoint, false, &resp); err != nil {
		return 0, err
	}
	return strconv.ParseFloat(resp.Spread, 64)
}

// PlaceOrder submits a new order. Requires valid API credentials.
func (c *Client) PlaceOrder(req *OrderRequest) (*OrderResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf("%s/order", c.clobHost)
	var resp OrderResponse
	if err := c.post(endpoint, body, &resp); err != nil {
		return nil, fmt.Errorf("place order: %w", err)
	}
	if !resp.Success {
		return &resp, fmt.Errorf("order rejected: %s", resp.ErrorMsg)
	}
	return &resp, nil
}

// CancelOrder cancels an open order by ID.
func (c *Client) CancelOrder(orderID string) error {
	body, _ := json.Marshal(map[string]string{"orderID": orderID})
	endpoint := fmt.Sprintf("%s/order/%s", c.clobHost, orderID)
	var resp struct {
		Deleted bool `json:"deleted"`
	}
	return c.delete(endpoint, body, &resp)
}

// GetOpenOrders fetches all open orders.
func (c *Client) GetOpenOrders() ([]map[string]interface{}, error) {
	endpoint := fmt.Sprintf("%s/orders?status=LIVE", c.clobHost)
	var orders []map[string]interface{}
	return orders, c.get(endpoint, true, &orders)
}

// ─── HTTP helpers ─────────────────────────────────────────────────────────────

func (c *Client) get(endpoint string, auth bool, out interface{}) error {
	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return err
	}
	if auth {
		c.signRequest(req, "")
	}
	return c.do(req, out)
}

func (c *Client) post(endpoint string, body []byte, out interface{}) error {
	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.signRequest(req, string(body))
	return c.do(req, out)
}

func (c *Client) delete(endpoint string, body []byte, out interface{}) error {
	req, err := http.NewRequest(http.MethodDelete, endpoint, strings.NewReader(string(body)))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	c.signRequest(req, string(body))
	return c.do(req, out)
}

func (c *Client) do(req *http.Request, out interface{}) error {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return fmt.Errorf("API error %d: %s", resp.StatusCode, string(data))
	}

	if out != nil {
		if err := json.Unmarshal(data, out); err != nil {
			return fmt.Errorf("decoding response: %w (body: %s)", err, string(data))
		}
	}
	return nil
}

// signRequest adds HMAC-SHA256 auth headers for the CLOB API.
func (c *Client) signRequest(req *http.Request, body string) {
	if c.apiKey == "" {
		return
	}
	ts := strconv.FormatInt(time.Now().UnixMilli(), 10)
	message := ts + req.Method + req.URL.Path + body
	mac := hmac.New(sha256.New, []byte(c.apiSecret))
	mac.Write([]byte(message))
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	req.Header.Set("POLY_ADDRESS", c.apiKey)
	req.Header.Set("POLY_SIGNATURE", sig)
	req.Header.Set("POLY_TIMESTAMP", ts)
	req.Header.Set("POLY_PASSPHRASE", c.passphrase)
}
