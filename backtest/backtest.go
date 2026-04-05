// Package backtest replays historical price data through a strategy and reports results.
package backtest

import (
	"fmt"
	"math"
	"sort"
	"time"

	"polymarket-bot/client"
	"polymarket-bot/positions"
	"polymarket-bot/strategy"
)

// Tick is a single point-in-time market state.
type Tick struct {
	Time    time.Time
	TokenID string
	Market  string
	Outcome string
	Bid     float64
	Ask     float64
	Mid     float64
}

// Result holds the outcome of a backtest run.
type Result struct {
	Strategy    string
	Market      string
	Ticks       int
	Trades      int
	WinTrades   int
	LoseTrades  int
	RealizedPnL float64
	MaxDrawdown float64
	SharpeRatio float64
	WinRate     float64
	Fills       []positions.Fill
	EquityCurve []float64
}

func (r Result) String() string {
	return fmt.Sprintf(
		`Backtest Results — %s on %q
  Ticks:       %d
  Trades:      %d  (W:%d L:%d  WinRate:%.1f%%)
  Realized PnL: $%.4f
  Max Drawdown: $%.4f
  Sharpe Ratio: %.3f`,
		r.Strategy, r.Market,
		r.Ticks, r.Trades, r.WinTrades, r.LoseTrades, r.WinRate*100,
		r.RealizedPnL, r.MaxDrawdown, r.SharpeRatio,
	)
}

// Runner executes a backtest.
type Runner struct {
	strat     strategy.Strategy
	slippage  float64 // fractional, e.g. 0.001 = 0.1%
	feeBps    float64 // maker/taker fee in bps, e.g. 100 = 1%
	orderSize float64 // USDC per order
}

// New creates a backtest runner.
// slippage: fraction added to buy price / subtracted from sell price.
// feeBps: fee in basis points per trade.
func New(strat strategy.Strategy, slippage, feeBps, orderSize float64) *Runner {
	return &Runner{
		strat:     strat,
		slippage:  slippage,
		feeBps:    feeBps,
		orderSize: orderSize,
	}
}

// Run replays ticks through the strategy and returns results.
func (r *Runner) Run(ticks []Tick) Result {
	if len(ticks) == 0 {
		return Result{Strategy: r.strat.Name()}
	}

	tracker := positions.New()
	var history []float64
	var equityCurve []float64
	var pnlHistory []float64
	trades := 0
	wins := 0
	losses := 0
	market := ticks[0].Market

	peakEquity := 0.0
	maxDrawdown := 0.0

	for _, tick := range ticks {
		history = append(history, tick.Mid)

		// Build a synthetic order book snapshot
		ob := &client.OrderBook{
			AssetID: tick.TokenID,
			Bids:    []client.PriceLevel{{Price: fmt.Sprintf("%.4f", tick.Bid)}},
			Asks:    []client.PriceLevel{{Price: fmt.Sprintf("%.4f", tick.Ask)}},
		}

		snap := strategy.MarketSnapshot{
			Market: client.GammaMarket{
				Question: tick.Market,
				Volume:   1000, // synthetic
			},
			OrderBook: ob,
			TokenID:   tick.TokenID,
			Outcome:   tick.Outcome,
		}

		sig := r.strat.Analyze(snap, history)
		if sig.Side == "NONE" {
			equityCurve = append(equityCurve, tracker.TotalRealizedPnL())
			continue
		}

		// Apply slippage and fee
		fillPrice := sig.Price
		feeMultiplier := 1.0 + (r.feeBps/10000.0)
		if sig.Side == "BUY" {
			fillPrice = sig.Price * (1 + r.slippage) * feeMultiplier
		} else {
			fillPrice = sig.Price * (1 - r.slippage) / feeMultiplier
		}

		// Cap size to order budget
		size := r.orderSize / fillPrice

		side := positions.Buy
		if sig.Side == "SELL" {
			side = positions.Sell
		}

		fill := tracker.RecordFill(tick.TokenID, tick.Market, tick.Outcome, side, fillPrice, size)
		trades++
		pnl := tracker.TotalRealizedPnL()

		if fill.Side == positions.Sell {
			pos := tracker.Position(tick.TokenID)
			if pos != nil && pos.RealizedPnL > 0 {
				wins++
			} else {
				losses++
			}
		}

		equityCurve = append(equityCurve, pnl)
		pnlHistory = append(pnlHistory, pnl)

		// Track drawdown
		if pnl > peakEquity {
			peakEquity = pnl
		}
		dd := peakEquity - pnl
		if dd > maxDrawdown {
			maxDrawdown = dd
		}
	}

	realizedPnL := tracker.TotalRealizedPnL()
	winRate := 0.0
	if trades > 0 {
		winRate = float64(wins) / float64(trades)
	}

	return Result{
		Strategy:    r.strat.Name(),
		Market:      market,
		Ticks:       len(ticks),
		Trades:      trades,
		WinTrades:   wins,
		LoseTrades:  losses,
		RealizedPnL: realizedPnL,
		MaxDrawdown: maxDrawdown,
		SharpeRatio: sharpe(pnlHistory),
		WinRate:     winRate,
		Fills:       tracker.RecentFills(trades),
		EquityCurve: equityCurve,
	}
}

// GenerateSyntheticTicks creates a realistic synthetic price series for testing.
// It uses a mean-reverting random walk to simulate prediction market dynamics.
func GenerateSyntheticTicks(
	tokenID, market, outcome string,
	startPrice float64,
	n int,
	volatility float64, // per-tick std dev, e.g. 0.01
	meanReversion float64, // pull toward 0.5, e.g. 0.02
) []Tick {
	ticks := make([]Tick, n)
	price := startPrice
	t := time.Now().Add(-time.Duration(n) * time.Minute)

	// Simple LCG for deterministic pseudo-randomness (no math/rand import)
	seed := uint64(12345)
	randNorm := func() float64 {
		// Box-Muller using two LCG values
		seed = seed*6364136223846793005 + 1442695040888963407
		u1 := float64(seed>>33) / float64(1<<31)
		seed = seed*6364136223846793005 + 1442695040888963407
		u2 := float64(seed>>33) / float64(1<<31)
		if u1 < 1e-10 {
			u1 = 1e-10
		}
		// Approximate normal: use u1 directly centered around 0
		z := (u1 - 0.5) * 2.0 * volatility
		_ = u2
		return z + meanReversionPull(price, meanReversion)
	}

	for i := range ticks {
		price += randNorm()
		if price < 0.01 {
			price = 0.01
		}
		if price > 0.99 {
			price = 0.99
		}
		spread := 0.005 + volatility*0.5
		ticks[i] = Tick{
			Time:    t.Add(time.Duration(i) * time.Minute),
			TokenID: tokenID,
			Market:  market,
			Outcome: outcome,
			Bid:     math.Max(0.01, price-spread/2),
			Ask:     math.Min(0.99, price+spread/2),
			Mid:     price,
		}
	}
	return ticks
}

func meanReversionPull(price, strength float64) float64 {
	return (0.5 - price) * strength
}

// CompareStrategies runs multiple strategies on the same ticks and returns sorted results.
func CompareStrategies(strats []strategy.Strategy, ticks []Tick, slippage, feeBps, orderSize float64) []Result {
	results := make([]Result, len(strats))
	for i, s := range strats {
		runner := New(s, slippage, feeBps, orderSize)
		results[i] = runner.Run(ticks)
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].RealizedPnL > results[j].RealizedPnL
	})
	return results
}

// sharpe computes an annualised Sharpe ratio from a P&L series.
// Uses 0 as the risk-free rate.
func sharpe(pnl []float64) float64 {
	if len(pnl) < 2 {
		return 0
	}
	// Compute returns
	rets := make([]float64, len(pnl)-1)
	for i := range rets {
		rets[i] = pnl[i+1] - pnl[i]
	}
	mean := 0.0
	for _, r := range rets {
		mean += r
	}
	mean /= float64(len(rets))

	variance := 0.0
	for _, r := range rets {
		d := r - mean
		variance += d * d
	}
	variance /= float64(len(rets))
	stddev := math.Sqrt(variance)
	if stddev == 0 {
		return 0
	}

	// Annualise assuming 1-minute ticks (525,600 per year)
	annFactor := math.Sqrt(525600.0)
	return (mean / stddev) * annFactor
}
