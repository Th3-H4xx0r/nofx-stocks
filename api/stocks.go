package api

import (
	"fmt"
	"net/http"
	"nofx/logger"
	"nofx/market"
	"nofx/ml"
	"sync"

	"github.com/gin-gonic/gin"
)

type StockSuggestion struct {
	Symbol          string  `json:"symbol"`
	Price           float64 `json:"price"`
	Change24h       float64 `json:"change_24h"`
	RSI             float64 `json:"rsi"`
	VWAP            float64 `json:"vwap"`
	NewsSentiment   float64 `json:"news_sentiment"`
	MLPrediction    string  `json:"ml_prediction"`
	Confidence      float64 `json:"confidence"`
	Reasoning       string  `json:"reasoning"`
}

// Popular tech stocks to scan
var targetStocks = []string{
	"AAPL", "MSFT", "GOOGL", "AMZN", "NVDA", "TSLA", "META", "AMD",
	"PLTR", "COIN", "MSTR", "INTC", "NFLX", "QCOM", "HOOD", "PYPL",
}

func (s *Server) handleGetStockSuggestions(c *gin.Context) {
	userID := c.GetString("user_id")

	// Find Alpaca config for user
	exchanges, err := s.store.Exchange().List(userID)
	if err != nil {
		SafeInternalError(c, "Failed to get exchange configs", err)
		return
	}

	var apiKey, secretKey string
	for _, ex := range exchanges {
		if ex.ExchangeType == "alpaca" && ex.Enabled {
			apiKey = string(ex.AlpacaAPIKey)
			secretKey = string(ex.AlpacaSecretKey)
			break
		}
	}

	if apiKey == "" || secretKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Alpaca exchange not configured or enabled"})
		return
	}

	// Create analyzer instance
	predictor := &ml.MockPredictor{}
	// sentiment := &ml.SimpleSentimentAnalyzer{} // Not used for now

	var suggestions []StockSuggestion
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Limit concurrency
	sem := make(chan struct{}, 5)

	for _, symbol := range targetStocks {
		wg.Add(1)
		go func(sym string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			// Fetch market data
			klines, err := market.GetAlpacaKlines(sym, "1d", apiKey, secretKey, 100) // 100 daily bars
			if err != nil {
				logger.Warnf("Failed to fetch data for %s: %v", sym, err)
				return
			}
			if len(klines) < 20 {
				return
			}

			current := klines[len(klines)-1]
			price := current.Close
			prev := klines[len(klines)-2]
			change24h := (price - prev.Close) / prev.Close * 100

			// Calculate indicators
			rsi := market.ExportCalculateRSI(klines, 14)

			// Fetch News (skipped for now)
			newsScore := 0.0

			// Predict
			pred, _ := predictor.Predict(sym, klines)

			sugg := StockSuggestion{
				Symbol:        sym,
				Price:         price,
				Change24h:     change24h,
				RSI:           rsi,
				NewsSentiment: newsScore,
				MLPrediction:  pred.Direction,
				Confidence:    pred.Confidence * 100,
				Reasoning:     fmt.Sprintf("RSI is %.2f, ML model suggests %s trend.", rsi, pred.Direction),
			}

			mu.Lock()
			suggestions = append(suggestions, sugg)
			mu.Unlock()
		}(symbol)
	}

	wg.Wait()

	c.JSON(http.StatusOK, suggestions)
}
