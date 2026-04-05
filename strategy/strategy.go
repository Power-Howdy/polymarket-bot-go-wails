package strategy

import (
	"fmt"
	"log"
	"strings"

	"polymarket-bot/client"
)

// Signal represents a trading action.
type Signal struct {
	TokenID string
	Market  string
	Side    string  // "BUY" or "SELL" or "NONE"
	Price   float64
	Size    float64
	Reason  string
}

func (s Signal) String() string {
	if s.Side == "NONE" {
		return fmt.Sprintf("[HOLD] %s @ %.3f — %s", s.Market, s.Price, s.Reason)
	}
	return fmt.Sprintf("[%s] %s tokenID=%s @ $%.3f size=%.2f — %s",
		s.Side, s.Market, s.TokenID[:8]+"…", s.Price, s.Size, s.Reason)
}

// MarketSnapshot captures a moment-in-time market state for a strategy.
type MarketSnapshot struct {
	Market    client.GammaMarket
	OrderBook *client.OrderBook
	TokenID   string // YES token
	Outcome   string
}

// Strategy is the interface all strategies implement.
type Strategy interface {
	Name() string
	Analyze(snap MarketSnapshot, history []float64) Signal
}

// ─── Monitor ─────────────────────────────────────────────────────────────────

// MonitorStrategy just logs — no orders.
type MonitorStrategy struct{}

func NewMonitor() *MonitorStrategy { return &MonitorStrategy{} }

func (m *MonitorStrategy) Name() string { return "monitor" }

func (m *MonitorStrategy) Analyze(snap MarketSnapshot, _ []float64) Signal {
	ob := snap.OrderBook
	return Signal{
		TokenID: snap.TokenID,
		Market:  snap.Market.Question,
		Side:    "NONE",
		Price:   ob.MidPrice(),
		Reason: fmt.Sprintf("bid=%.3f ask=%.3f spread=%.4f vol=$%.0f",
			ob.BestBid(), ob.BestAsk(), ob.Spread(), snap.Market.Volume),
	}
}

// ─── Spread Maker ────────────────────────────────────────────────────────────

// SpreadStrategy posts bids and asks around the mid-price.
type SpreadStrategy struct {
	SpreadBps float64 // e.g. 50 = 0.5%
	OrderSize float64 // USDC per side
}

func NewSpread(spreadBps, orderSize float64) *SpreadStrategy {
	return &SpreadStrategy{SpreadBps: spreadBps, OrderSize: orderSize}
}

func (s *SpreadStrategy) Name() string { return "spread" }

func (s *SpreadStrategy) Analyze(snap MarketSnapshot, _ []float64) Signal {
	ob := snap.OrderBook
	mid := ob.MidPrice()
	if mid == 0 {
		return Signal{Side: "NONE", Reason: "no mid price"}
	}

	halfSpread := (s.SpreadBps / 10_000.0) / 2.0
	bidPrice := mid - halfSpread
	askPrice := mid + halfSpread

	// Clamp to valid range
	bidPrice = clamp(bidPrice, 0.01, 0.99)
	askPrice = clamp(askPrice, 0.01, 0.99)

	// Simple heuristic: if current ask > our ask target, there's room to place a sell
	if ob.BestAsk() > askPrice+0.01 {
		size := s.OrderSize / askPrice
		return Signal{
			TokenID: snap.TokenID,
			Market:  snap.Market.Question,
			Side:    "SELL",
			Price:   askPrice,
			Size:    size,
			Reason:  fmt.Sprintf("mid=%.3f placing ask at %.3f (spread %.0fbps)", mid, askPrice, s.SpreadBps),
		}
	}

	// If current bid < our bid target, there's room to place a buy
	if ob.BestBid() < bidPrice-0.01 {
		size := s.OrderSize / bidPrice
		return Signal{
			TokenID: snap.TokenID,
			Market:  snap.Market.Question,
			Side:    "BUY",
			Price:   bidPrice,
			Size:    size,
			Reason:  fmt.Sprintf("mid=%.3f placing bid at %.3f (spread %.0fbps)", mid, bidPrice, s.SpreadBps),
		}
	}

	return Signal{
		Side:   "NONE",
		Price:  mid,
		Market: snap.Market.Question,
		Reason: fmt.Sprintf("spread tight (bid=%.3f ask=%.3f), no edge", ob.BestBid(), ob.BestAsk()),
	}
}

// ─── Momentum ────────────────────────────────────────────────────────────────

// MomentumStrategy buys rising markets and sells falling ones.
type MomentumStrategy struct {
	Threshold float64 // fractional move, e.g. 0.03 = 3%
	Lookback  int     // ticks to compare against
	OrderSize float64
}

func NewMomentum(threshold float64, lookback int, orderSize float64) *MomentumStrategy {
	return &MomentumStrategy{Threshold: threshold, Lookback: lookback, OrderSize: orderSize}
}

func (m *MomentumStrategy) Name() string { return "momentum" }

func (m *MomentumStrategy) Analyze(snap MarketSnapshot, history []float64) Signal {
	ob := snap.OrderBook
	current := ob.MidPrice()

	if len(history) < m.Lookback {
		return Signal{Side: "NONE", Market: snap.Market.Question,
			Price: current, Reason: "insufficient history"}
	}

	past := history[len(history)-m.Lookback]
	if past == 0 {
		return Signal{Side: "NONE", Market: snap.Market.Question,
			Price: current, Reason: "zero baseline price"}
	}

	change := (current - past) / past
	log.Printf("  momentum: %s change=%.2f%% (%.3f → %.3f)",
		snap.Market.Question, change*100, past, current)

	if change >= m.Threshold {
		return Signal{
			TokenID: snap.TokenID,
			Market:  snap.Market.Question,
			Side:    "BUY",
			Price:   ob.BestAsk(), // take liquidity
			Size:    m.OrderSize / ob.BestAsk(),
			Reason:  fmt.Sprintf("momentum UP %.2f%% over %d ticks", change*100, m.Lookback),
		}
	}

	if change <= -m.Threshold {
		return Signal{
			TokenID: snap.TokenID,
			Market:  snap.Market.Question,
			Side:    "SELL",
			Price:   ob.BestBid(), // take liquidity
			Size:    m.OrderSize / ob.BestBid(),
			Reason:  fmt.Sprintf("momentum DOWN %.2f%% over %d ticks", change*100, m.Lookback),
		}
	}

	return Signal{
		Side:   "NONE",
		Price:  current,
		Market: snap.Market.Question,
		Reason: fmt.Sprintf("change=%.2f%% below threshold %.2f%%", change*100, m.Threshold*100),
	}
}

// ─── Factory ─────────────────────────────────────────────────────────────────

// New creates a strategy by name.
func New(name string, spreadBps, orderSize, momentumThreshold float64, momentumLookback int) (Strategy, error) {
	switch strings.ToLower(name) {
	case "monitor":
		return NewMonitor(), nil
	case "spread":
		return NewSpread(spreadBps, orderSize), nil
	case "momentum":
		return NewMomentum(momentumThreshold, momentumLookback, orderSize), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %q (choose: monitor, spread, momentum)", name)
	}
}

func clamp(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
