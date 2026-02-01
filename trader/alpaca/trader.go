package alpaca

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"nofx/logger"
	"nofx/trader/types"
	"strconv"
	"time"
)

const (
	LiveBaseURL  = "https://api.alpaca.markets"
	PaperBaseURL = "https://paper-api.alpaca.markets"
)

// AlpacaTrader implements Trader interface for Alpaca
type AlpacaTrader struct {
	apiKey    string
	secretKey string
	baseURL   string
	client    *http.Client
}

// NewAlpacaTrader creates a new AlpacaTrader
func NewAlpacaTrader(apiKey, secretKey string, paper bool) *AlpacaTrader {
	baseURL := LiveBaseURL
	if paper {
		baseURL = PaperBaseURL
	}
	return &AlpacaTrader{
		apiKey:    apiKey,
		secretKey: secretKey,
		baseURL:   baseURL,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

// GetBalance returns account balance
func (t *AlpacaTrader) GetBalance() (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", t.baseURL+"/v2/account", nil)
	if err != nil {
		return nil, err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Alpaca API error: %s", string(body))
	}

	var account struct {
		Equity        string `json:"equity"`
		Cash          string `json:"cash"`
		BuyingPower   string `json:"buying_power"`
		DayTradeCount int    `json:"daytrade_count"`
		PatternDayTrader bool `json:"pattern_day_trader"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&account); err != nil {
		return nil, err
	}

	equity, _ := strconv.ParseFloat(account.Equity, 64)
	cash, _ := strconv.ParseFloat(account.Cash, 64)
	buyingPower, _ := strconv.ParseFloat(account.BuyingPower, 64)

	// Map to standard fields
	return map[string]interface{}{
		"total_equity":      equity,
		"available_balance": buyingPower, // Or cash? Buying power is what matters for trading.
		"wallet_balance":    cash,
		"total_pnl":         0.0, // Calculated by manager usually
		"day_trade_count":   account.DayTradeCount,
		"pattern_day_trader": account.PatternDayTrader,
	}, nil
}

// GetPositions returns all positions
func (t *AlpacaTrader) GetPositions() ([]map[string]interface{}, error) {
	req, err := http.NewRequest("GET", t.baseURL+"/v2/positions", nil)
	if err != nil {
		return nil, err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Alpaca API error: %s", string(body))
	}

	var positions []struct {
		Symbol           string `json:"symbol"`
		Qty              string `json:"qty"`
		Side             string `json:"side"` // "long" or "short"
		AvgEntryPrice    string `json:"avg_entry_price"`
		CurrentPrice     string `json:"current_price"`
		UnrealizedPL     string `json:"unrealized_pl"`
		UnrealizedPLPC   string `json:"unrealized_plpc"`
		MarketValue      string `json:"market_value"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&positions); err != nil {
		return nil, err
	}

	var result []map[string]interface{}
	for _, p := range positions {
		qty, _ := strconv.ParseFloat(p.Qty, 64)
		entryPrice, _ := strconv.ParseFloat(p.AvgEntryPrice, 64)
		markPrice, _ := strconv.ParseFloat(p.CurrentPrice, 64)
		unrealizedPnl, _ := strconv.ParseFloat(p.UnrealizedPL, 64)

		result = append(result, map[string]interface{}{
			"symbol":           p.Symbol,
			"side":             p.Side, // "long" or "short"
			"positionAmt":      qty,
			"entryPrice":       entryPrice,
			"markPrice":        markPrice,
			"unRealizedProfit": unrealizedPnl,
			"leverage":         1.0, // Stocks usually 1x (or margin)
			"liquidationPrice": 0.0, // Not applicable for standard stocks usually
		})
	}
	return result, nil
}

// OpenLong opens a long position (Buy)
func (t *AlpacaTrader) OpenLong(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	// Stocks leverage is managed by account type, usually 2x overnight, 4x day. We send order.
	return t.placeOrder(symbol, quantity, "buy")
}

// OpenShort opens a short position (Sell)
func (t *AlpacaTrader) OpenShort(symbol string, quantity float64, leverage int) (map[string]interface{}, error) {
	return t.placeOrder(symbol, quantity, "sell")
}

// CloseLong closes a long position (Sell)
func (t *AlpacaTrader) CloseLong(symbol string, quantity float64) (map[string]interface{}, error) {
	if quantity == 0 {
		return t.closePosition(symbol)
	}
	return t.placeOrder(symbol, quantity, "sell")
}

// CloseShort closes a short position (Buy)
func (t *AlpacaTrader) CloseShort(symbol string, quantity float64) (map[string]interface{}, error) {
	if quantity == 0 {
		return t.closePosition(symbol)
	}
	return t.placeOrder(symbol, quantity, "buy")
}

func (t *AlpacaTrader) placeOrder(symbol string, qty float64, side string) (map[string]interface{}, error) {
	url := t.baseURL + "/v2/orders"

	// Format qty? Alpaca supports fractional shares.
	// For simplicity, convert to string.

	orderReq := map[string]interface{}{
		"symbol":        symbol,
		"qty":           fmt.Sprintf("%f", qty),
		"side":          side,
		"type":          "market",
		"time_in_force": "day", // or gtc
	}

	body, _ := json.Marshal(orderReq)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	t.addHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Alpaca order error: %s", string(respBody))
	}

	var order struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"orderId": order.ID,
		"symbol":  order.Symbol,
		"status":  order.Status,
	}, nil
}

func (t *AlpacaTrader) closePosition(symbol string) (map[string]interface{}, error) {
	// DELETE /v2/positions/{symbol}
	url := fmt.Sprintf("%s/v2/positions/%s", t.baseURL, symbol)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Alpaca close position error: %s", string(body))
	}

	var order struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"orderId": order.ID,
		"symbol":  order.Symbol,
		"status":  order.Status,
	}, nil
}

// SetLeverage - Alpaca leverage is fixed by account type/equity, can't set per symbol
func (t *AlpacaTrader) SetLeverage(symbol string, leverage int) error {
	return nil
}

// SetMarginMode - Alpaca is account-wide margin
func (t *AlpacaTrader) SetMarginMode(symbol string, isCrossMargin bool) error {
	return nil
}

// GetMarketPrice
func (t *AlpacaTrader) GetMarketPrice(symbol string) (float64, error) {
	// Use Data API (market/alpaca_provider.go logic logic, but simpler here for last quote)
	// GET /v2/stocks/{symbol}/quotes/latest
	// Or use Alpaca Data API if enabled.
	// But trader usually uses Trading API.
	// Trading API has /v2/last_quote if using free data?
	// Actually for free plan, use IEX via Data API.
	// Re-use logic from market/alpaca_provider.go?
	// Or simply fetch from positions if holding?
	// For general price check, we should use Data API.
	// But Trader needs to know current price.

	// Let's assume we can use the same keys for Data API.
	// https://data.alpaca.markets/v2/stocks/{symbol}/quotes/latest

	dataURL := "https://data.alpaca.markets/v2/stocks/" + symbol + "/quotes/latest?feed=iex"
	req, err := http.NewRequest("GET", dataURL, nil)
	if err != nil {
		return 0, err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to get quote")
	}

	var result struct {
		Quote struct {
			Ap float64 `json:"ap"` // Ask price
			Bp float64 `json:"bp"` // Bid price
		} `json:"quote"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	return (result.Quote.Ap + result.Quote.Bp) / 2, nil
}

// SetStopLoss
func (t *AlpacaTrader) SetStopLoss(symbol string, positionSide string, quantity, stopPrice float64) error {
	// Alpaca supports Bracket Orders or separate Stop orders
	// Here we just place a stop order
	// Side needs to be opposite of position
	side := "sell"
	if positionSide == "SHORT" {
		side = "buy"
	}

	url := t.baseURL + "/v2/orders"
	orderReq := map[string]interface{}{
		"symbol":        symbol,
		"qty":           fmt.Sprintf("%f", quantity),
		"side":          side,
		"type":          "stop",
		"stop_price":    fmt.Sprintf("%f", stopPrice),
		"time_in_force": "gtc",
	}

	body, _ := json.Marshal(orderReq)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	t.addHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set stop loss")
	}
	return nil
}

// SetTakeProfit
func (t *AlpacaTrader) SetTakeProfit(symbol string, positionSide string, quantity, takeProfitPrice float64) error {
	// Similar to Stop Loss but type=limit or take_profit (if supported)
	// Alpaca standard is Limit order for TP
	side := "sell"
	if positionSide == "SHORT" {
		side = "buy"
	}

	url := t.baseURL + "/v2/orders"
	orderReq := map[string]interface{}{
		"symbol":        symbol,
		"qty":           fmt.Sprintf("%f", quantity),
		"side":          side,
		"type":          "limit",
		"limit_price":   fmt.Sprintf("%f", takeProfitPrice),
		"time_in_force": "gtc",
	}

	body, _ := json.Marshal(orderReq)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	t.addHeaders(req)
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to set take profit")
	}
	return nil
}

// CancelStopLossOrders
func (t *AlpacaTrader) CancelStopLossOrders(symbol string) error {
	return t.CancelAllOrders(symbol) // Simplified
}

// CancelTakeProfitOrders
func (t *AlpacaTrader) CancelTakeProfitOrders(symbol string) error {
	return t.CancelAllOrders(symbol) // Simplified
}

// CancelAllOrders
func (t *AlpacaTrader) CancelAllOrders(symbol string) error {
	// DELETE /v2/orders
	// Can filter by symbol? Not directly in DELETE /v2/orders (cancels ALL).
	// We need to list then cancel.
	// GET /v2/orders?status=open&symbols=symbol

	// Simplified: Cancel ALL for symbol
	// Note: Alpaca DELETE /v2/orders cancels ALL open orders.
	// To cancel by symbol, we need to fetch and cancel individually.

	req, err := http.NewRequest("GET", t.baseURL+"/v2/orders?status=open&symbols="+symbol, nil)
	if err != nil {
		return err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var orders []struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return err
	}

	for _, o := range orders {
		delReq, _ := http.NewRequest("DELETE", t.baseURL+"/v2/orders/"+o.ID, nil)
		t.addHeaders(delReq)
		t.client.Do(delReq)
	}

	return nil
}

// CancelStopOrders
func (t *AlpacaTrader) CancelStopOrders(symbol string) error {
	return t.CancelAllOrders(symbol)
}

// FormatQuantity
func (t *AlpacaTrader) FormatQuantity(symbol string, quantity float64) (string, error) {
	// Alpaca supports fractional shares, but let's stick to 4 decimals
	return fmt.Sprintf("%.4f", quantity), nil
}

// GetOrderStatus
func (t *AlpacaTrader) GetOrderStatus(symbol string, orderID string) (map[string]interface{}, error) {
	req, err := http.NewRequest("GET", t.baseURL+"/v2/orders/"+orderID, nil)
	if err != nil {
		return nil, err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var order struct {
		ID           string `json:"id"`
		Status       string `json:"status"`
		Symbol       string `json:"symbol"`
		FilledAvgPrice string `json:"filled_avg_price"`
		FilledQty    string `json:"filled_qty"`
		Side         string `json:"side"`
		Type         string `json:"type"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&order); err != nil {
		return nil, err
	}

	avgPrice, _ := strconv.ParseFloat(order.FilledAvgPrice, 64)
	filledQty, _ := strconv.ParseFloat(order.FilledQty, 64)

	// Convert status to standard format (NEW, FILLED, CANCELED)
	status := "NEW"
	if order.Status == "filled" {
		status = "FILLED"
	} else if order.Status == "canceled" || order.Status == "expired" || order.Status == "rejected" {
		status = "CANCELED"
	}

	return map[string]interface{}{
		"orderId":     order.ID,
		"symbol":      order.Symbol,
		"status":      status,
		"avgPrice":    avgPrice,
		"executedQty": filledQty,
		"side":        order.Side,
		"type":        order.Type,
		"time":        order.CreatedAt,
		"commission":  0.0, // Alpaca usually 0 commission
	}, nil
}

// GetClosedPnL
func (t *AlpacaTrader) GetClosedPnL(startTime time.Time, limit int) ([]types.ClosedPnLRecord, error) {
	// Alpaca doesn't have a direct "Closed PnL" endpoint for individual trades easily.
	// But we can fetch /v2/account/portfolio/history or /v2/activities
	// GET /v2/account/activities/FILL
	return []types.ClosedPnLRecord{}, nil
}

// GetOpenOrders
func (t *AlpacaTrader) GetOpenOrders(symbol string) ([]types.OpenOrder, error) {
	req, err := http.NewRequest("GET", t.baseURL+"/v2/orders?status=open&symbols="+symbol, nil)
	if err != nil {
		return nil, err
	}
	t.addHeaders(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var orders []struct {
		ID        string `json:"id"`
		Symbol    string `json:"symbol"`
		Side      string `json:"side"`
		Type      string `json:"type"`
		Qty       string `json:"qty"`
		LimitPrice string `json:"limit_price"`
		StopPrice string `json:"stop_price"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&orders); err != nil {
		return nil, err
	}

	var result []types.OpenOrder
	for _, o := range orders {
		qty, _ := strconv.ParseFloat(o.Qty, 64)
		price, _ := strconv.ParseFloat(o.LimitPrice, 64)
		stopPrice, _ := strconv.ParseFloat(o.StopPrice, 64)

		result = append(result, types.OpenOrder{
			OrderID:   o.ID,
			Symbol:    o.Symbol,
			Side:      o.Side,
			Type:      o.Type,
			Quantity:  qty,
			Price:     price,
			StopPrice: stopPrice,
			Status:    "NEW",
		})
	}

	return result, nil
}

func (t *AlpacaTrader) addHeaders(req *http.Request) {
	req.Header.Add("APCA-API-KEY-ID", t.apiKey)
	req.Header.Add("APCA-API-SECRET-KEY", t.secretKey)
}
