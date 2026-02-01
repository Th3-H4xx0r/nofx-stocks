package market

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	alpacaDataURL      = "https://data.alpaca.markets/v2/stocks/%s/bars"
	alpacaPaperDataURL = "https://data.alpaca.markets/v2/stocks/%s/bars" // Data API is same for paper/live usually, but depends on subscription. Free plan uses IEX data.
)

type AlpacaBar struct {
	T time.Time `json:"t"` // Timestamp
	O float64   `json:"o"` // Open
	H float64   `json:"h"` // High
	L float64   `json:"l"` // Low
	C float64   `json:"c"` // Close
	V float64   `json:"v"` // Volume
}

type AlpacaBarsResponse struct {
	Bars          []AlpacaBar `json:"bars"`
	NextPageToken string      `json:"next_page_token"`
}

// GetAlpacaKlines fetches kline data from Alpaca API
// apiKey/secretKey are required for authentication
func GetAlpacaKlines(symbol, timeframe, apiKey, secretKey string, limit int) ([]Kline, error) {
	// Map timeframe to Alpaca format (1Min, 5Min, 15Min, 1Hour, 1Day)
	alpacaTimeframe, err := mapToAlpacaTimeframe(timeframe)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf(alpacaDataURL, symbol)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	q := req.URL.Query()
	q.Add("timeframe", alpacaTimeframe)
	q.Add("limit", fmt.Sprintf("%d", limit))
	q.Add("adjustment", "raw") // or "split"
	q.Add("feed", "iex")       // "iex" for free, "sip" for paid
	req.URL.RawQuery = q.Encode()

	req.Header.Add("APCA-API-KEY-ID", apiKey)
	req.Header.Add("APCA-API-SECRET-KEY", secretKey)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Alpaca API error: %s", string(body))
	}

	var result AlpacaBarsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	klines := make([]Kline, len(result.Bars))
	for i, bar := range result.Bars {
		klines[i] = Kline{
			OpenTime:  bar.T.UnixMilli(),
			Open:      bar.O,
			High:      bar.H,
			Low:       bar.L,
			Close:     bar.C,
			Volume:    bar.V,
			CloseTime: bar.T.Add(parseAlpacaTimeframeDuration(timeframe)).UnixMilli(),
		}
	}

	return klines, nil
}

// GetAlpacaKlinesRange fetches klines in a specific time range
func GetAlpacaKlinesRange(symbol, timeframe, apiKey, secretKey string, start, end time.Time) ([]Kline, error) {
	alpacaTimeframe, err := mapToAlpacaTimeframe(timeframe)
	if err != nil {
		return nil, err
	}

	var allKlines []Kline
	pageToken := ""

	for {
		url := fmt.Sprintf(alpacaDataURL, symbol)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		q := req.URL.Query()
		q.Add("timeframe", alpacaTimeframe)
		q.Add("start", start.Format(time.RFC3339))
		q.Add("end", end.Format(time.RFC3339))
		q.Add("limit", "10000") // Max limit per page
		q.Add("adjustment", "raw")
		q.Add("feed", "iex")
		if pageToken != "" {
			q.Add("page_token", pageToken)
		}
		req.URL.RawQuery = q.Encode()

		req.Header.Add("APCA-API-KEY-ID", apiKey)
		req.Header.Add("APCA-API-SECRET-KEY", secretKey)

		client := &http.Client{Timeout: 10 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("Alpaca API error: %s", string(body))
		}

		var result AlpacaBarsResponse
		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			return nil, err
		}

		for _, bar := range result.Bars {
			allKlines = append(allKlines, Kline{
				OpenTime:  bar.T.UnixMilli(),
				Open:      bar.O,
				High:      bar.H,
				Low:       bar.L,
				Close:     bar.C,
				Volume:    bar.V,
				CloseTime: bar.T.Add(parseAlpacaTimeframeDuration(timeframe)).UnixMilli(),
			})
		}

		if result.NextPageToken == "" {
			break
		}
		pageToken = result.NextPageToken
	}

	return allKlines, nil
}

func mapToAlpacaTimeframe(timeframe string) (string, error) {
	switch timeframe {
	case "1m":
		return "1Min", nil
	case "5m":
		return "5Min", nil
	case "15m":
		return "15Min", nil
	case "1h":
		return "1Hour", nil
	case "1d":
		return "1Day", nil
	default:
		return "", fmt.Errorf("unsupported timeframe: %s", timeframe)
	}
}

func parseAlpacaTimeframeDuration(tf string) time.Duration {
	switch tf {
	case "1m":
		return time.Minute
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	case "1d":
		return 24 * time.Hour
	default:
		return time.Minute
	}
}
