package backtest

import (
	"testing"
	"time"
	"github.com/stretchr/testify/assert"
)

func TestBacktestConfig_AssetClass_Validation(t *testing.T) {
	now := time.Now().Unix()

	// Test Default (Crypto)
	cfg := &BacktestConfig{
		RunID: "test_run",
		Symbols: []string{"btc"},
		StartTS: now - 3600,
		EndTS: now,
	}
	err := cfg.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "crypto", cfg.AssetClass)
	assert.Equal(t, "BTCUSDT", cfg.Symbols[0])

	// Test Stocks
	cfgStocks := &BacktestConfig{
		RunID: "test_run_stocks",
		Symbols: []string{"aapl"},
		AssetClass: "stocks",
		StartTS: now - 3600,
		EndTS: now,
	}
	err = cfgStocks.Validate()
	assert.NoError(t, err)
	assert.Equal(t, "stocks", cfgStocks.AssetClass)
	assert.Equal(t, "AAPL", cfgStocks.Symbols[0]) // Uppercased, no USDT
}

func TestBacktestConfig_ToStrategyConfig_Stocks(t *testing.T) {
	now := time.Now().Unix()
	cfg := &BacktestConfig{
		RunID: "test_run",
		Symbols: []string{"AAPL"},
		AssetClass: "stocks",
		Timeframes: []string{"5m"},
		StartTS: now - 3600,
		EndTS: now,
	}
	// ensure defaults
	cfg.Validate()

	stratCfg := cfg.ToStrategyConfig()
	assert.Equal(t, "stocks", stratCfg.AssetClass)
	assert.Equal(t, false, stratCfg.Indicators.EnableOI)
	assert.Equal(t, false, stratCfg.Indicators.EnableFundingRate)
	assert.Equal(t, true, stratCfg.Indicators.EnableVolume)
	assert.Equal(t, []string{"AAPL"}, stratCfg.CoinSource.StaticCoins)
}

func TestBacktestConfig_ToStrategyConfig_Crypto(t *testing.T) {
	now := time.Now().Unix()
	cfg := &BacktestConfig{
		RunID: "test_run",
		Symbols: []string{"BTC"},
		AssetClass: "crypto",
		Timeframes: []string{"5m"},
		StartTS: now - 3600,
		EndTS: now,
	}
	// ensure defaults
	cfg.Validate()

	stratCfg := cfg.ToStrategyConfig()
	assert.Equal(t, "crypto", stratCfg.AssetClass)
	assert.Equal(t, true, stratCfg.Indicators.EnableOI)
	assert.Equal(t, true, stratCfg.Indicators.EnableFundingRate)
	assert.Equal(t, []string{"BTCUSDT"}, stratCfg.CoinSource.StaticCoins)
}
