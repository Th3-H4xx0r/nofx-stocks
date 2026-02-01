package ml

import (
	"math/rand"
	"strings"
	"time"
)

// Prediction result
type Prediction struct {
	Symbol     string  `json:"symbol"`
	Direction  string  `json:"direction"` // "bullish", "bearish", "neutral"
	Confidence float64 `json:"confidence"` // 0.0 - 1.0
	PriceTarget float64 `json:"price_target"`
	ModelName  string  `json:"model_name"`
	Timestamp  int64   `json:"timestamp"`
}

// Predictor interface
type Predictor interface {
	Predict(symbol string, data interface{}) (*Prediction, error)
}

// MockPredictor for testing
type MockPredictor struct{}

func (p *MockPredictor) Predict(symbol string, data interface{}) (*Prediction, error) {
	// Random prediction for now
	rand.Seed(time.Now().UnixNano())
	directions := []string{"bullish", "bearish", "neutral"}
	dir := directions[rand.Intn(len(directions))]
	conf := 0.5 + rand.Float64()*0.4

	return &Prediction{
		Symbol:     symbol,
		Direction:  dir,
		Confidence: conf,
		ModelName:  "MockLSTM",
		Timestamp:  time.Now().UnixMilli(),
	}, nil
}

// SentimentAnalyzer interface
type SentimentAnalyzer interface {
	Analyze(text string) (float64, error) // Returns score -1.0 to 1.0
}

// SimpleSentimentAnalyzer (keyword based)
type SimpleSentimentAnalyzer struct{}

func (s *SimpleSentimentAnalyzer) Analyze(text string) (float64, error) {
	// Very basic keyword matching
	// In production this should call an external NLP service or use a local model
	bullish := []string{"surge", "jump", "high", "gain", "bull", "growth", "positive", "up"}
	bearish := []string{"drop", "fall", "low", "loss", "bear", "decline", "negative", "down", "crash"}

	score := 0.0
	lowerText := strings.ToLower(text)

	for _, word := range bullish {
		if strings.Contains(lowerText, word) {
			score += 0.1
		}
	}
	for _, word := range bearish {
		if strings.Contains(lowerText, word) {
			score -= 0.1
		}
	}

	if score > 1.0 { score = 1.0 }
	if score < -1.0 { score = -1.0 }

	return score, nil
}
