package app

import (
	"fmt"

	"polymarket-bot/backtest"
	"polymarket-bot/strategy"
)

// BacktestRequest carries parameters from the frontend.
type BacktestRequest struct {
	Strategy  string  `json:"strategy"`
	Ticks     int     `json:"ticks"`
	Vol       float64 `json:"vol"`
	SpreadBps float64 `json:"spreadBps"`
	OrderSize float64 `json:"orderSize"`
	Slippage  float64 `json:"slippage"`
	FeeBps    float64 `json:"feeBps"`
	Compare   bool    `json:"compare"`
}

// BacktestResult carries backtest output to the frontend.
type BacktestResult struct {
	Strategy    string           `json:"strategy"`
	PnL         float64          `json:"pnl"`
	Trades      int              `json:"trades"`
	Wins        int              `json:"wins"`
	Losses      int              `json:"losses"`
	WinRate     float64          `json:"winRate"`
	MaxDrawdown float64          `json:"maxDrawdown"`
	EquityCurve []float64        `json:"equityCurve"`
	Fills       []FillSummary    `json:"fills"`
	Comparisons []StrategyResult `json:"comparisons,omitempty"`
}

// FillSummary is a single trade fill for display.
type FillSummary struct {
	Side  string  `json:"side"`
	Price float64 `json:"price"`
	Size  float64 `json:"size"`
	Cost  float64 `json:"cost"`
}

// StrategyResult is one entry in a compare run.
type StrategyResult struct {
	Name    string  `json:"name"`
	PnL     float64 `json:"pnl"`
	Trades  int     `json:"trades"`
	WinRate float64 `json:"winRate"`
}

func runBacktest(req BacktestRequest) (*BacktestResult, error) {
	ticks := backtest.GenerateSyntheticTicks(
		"token-abc123",
		"Will BTC exceed $100k by end of 2025?",
		"Yes",
		0.45,
		req.Ticks,
		req.Vol,
		0.02,
	)

	if req.Compare {
		var strats []strategy.Strategy
		for _, name := range []string{"monitor", "spread", "momentum"} {
			s, err := strategy.New(name, req.SpreadBps, req.OrderSize, 0.03, 5)
			if err != nil {
				continue
			}
			strats = append(strats, s)
		}
		results := backtest.CompareStrategies(strats, ticks, req.Slippage, req.FeeBps, req.OrderSize)
		var comps []StrategyResult
		for _, r := range results {
			comps = append(comps, StrategyResult{
				Name: r.Strategy, PnL: r.RealizedPnL,
				Trades: r.Trades, WinRate: r.WinRate,
			})
		}
		return &BacktestResult{Strategy: "compare", Comparisons: comps}, nil
	}

	strat, err := strategy.New(req.Strategy, req.SpreadBps, req.OrderSize, 0.03, 5)
	if err != nil {
		return nil, fmt.Errorf("unknown strategy: %w", err)
	}

	runner := backtest.New(strat, req.Slippage, req.FeeBps, req.OrderSize)
	r := runner.Run(ticks)

	result := &BacktestResult{
		Strategy:    r.Strategy,
		PnL:         r.RealizedPnL,
		Trades:      r.Trades,
		Wins:        r.WinTrades,
		Losses:      r.LoseTrades,
		WinRate:     r.WinRate,
		MaxDrawdown: r.MaxDrawdown,
		EquityCurve: r.EquityCurve,
	}
	for _, f := range r.Fills {
		result.Fills = append(result.Fills, FillSummary{
			Side: string(f.Side), Price: f.Price, Size: f.Size, Cost: f.Cost,
		})
	}
	return result, nil
}
