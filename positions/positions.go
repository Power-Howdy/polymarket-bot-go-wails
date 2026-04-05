// Package positions tracks open positions, fill history, and P&L.
package positions

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// Side represents a trade direction.
type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

// Fill represents a completed order execution.
type Fill struct {
	ID        string
	TokenID   string
	Market    string
	Outcome   string
	Side      Side
	Price     float64
	Size      float64
	Cost      float64 // Price * Size
	Timestamp time.Time
}

func (f Fill) String() string {
	return fmt.Sprintf("[%s] %s %-4s %.4f @ $%.4f (cost $%.2f) %s",
		f.Timestamp.Format("15:04:05"), f.Market, f.Side, f.Size, f.Price,
		f.Cost, f.Outcome)
}

// Position represents an aggregated holding in a single token.
type Position struct {
	TokenID     string
	Market      string
	Outcome     string
	TotalSize   float64 // net tokens held
	AvgBuyPrice float64 // average cost basis
	TotalCost   float64 // total USDC spent buying
	RealizedPnL float64 // from sells
}

// UnrealizedPnL calculates open P&L given a current market price.
func (p *Position) UnrealizedPnL(currentPrice float64) float64 {
	return (currentPrice - p.AvgBuyPrice) * p.TotalSize
}

// CurrentValue returns the mark-to-market value of the position.
func (p *Position) CurrentValue(currentPrice float64) float64 {
	return p.TotalSize * currentPrice
}

func (p Position) String() string {
	return fmt.Sprintf("%-50s | size=%.2f avgCost=$%.4f realized=$%.2f",
		truncate(p.Market+" ("+p.Outcome+")", 50),
		p.TotalSize, p.AvgBuyPrice, p.RealizedPnL)
}

// ─── Tracker ──────────────────────────────────────────────────────────────────

// Tracker maintains fill history and aggregated positions.
type Tracker struct {
	mu        sync.RWMutex
	fills     []Fill
	positions map[string]*Position // keyed by tokenID
	fillCount int
}

// New creates a new Tracker.
func New() *Tracker {
	return &Tracker{
		positions: make(map[string]*Position),
	}
}

// RecordFill adds a fill and updates the relevant position.
func (t *Tracker) RecordFill(tokenID, market, outcome string, side Side, price, size float64) Fill {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.fillCount++
	f := Fill{
		ID:        fmt.Sprintf("fill-%04d", t.fillCount),
		TokenID:   tokenID,
		Market:    market,
		Outcome:   outcome,
		Side:      side,
		Price:     price,
		Size:      size,
		Cost:      price * size,
		Timestamp: time.Now(),
	}
	t.fills = append(t.fills, f)

	pos, ok := t.positions[tokenID]
	if !ok {
		pos = &Position{
			TokenID: tokenID,
			Market:  market,
			Outcome: outcome,
		}
		t.positions[tokenID] = pos
	}

	switch side {
	case Buy:
		// Update average cost basis
		newTotalCost := pos.TotalCost + f.Cost
		newTotalSize := pos.TotalSize + size
		if newTotalSize > 0 {
			pos.AvgBuyPrice = newTotalCost / newTotalSize
		}
		pos.TotalCost = newTotalCost
		pos.TotalSize = newTotalSize

	case Sell:
		realized := (price - pos.AvgBuyPrice) * size
		pos.RealizedPnL += realized
		pos.TotalSize -= size
		if pos.TotalSize < 0 {
			pos.TotalSize = 0
		}
	}

	return f
}

// Position returns the current position for a token (nil if none).
func (t *Tracker) Position(tokenID string) *Position {
	t.mu.RLock()
	defer t.mu.RUnlock()
	p := t.positions[tokenID]
	if p == nil {
		return nil
	}
	cp := *p
	return &cp
}

// AllPositions returns all positions with non-zero size, sorted by market name.
func (t *Tracker) AllPositions() []Position {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var out []Position
	for _, p := range t.positions {
		if p.TotalSize > 0.001 {
			out = append(out, *p)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Market < out[j].Market
	})
	return out
}

// RecentFills returns the last n fills (most recent first).
func (t *Tracker) RecentFills(n int) []Fill {
	t.mu.RLock()
	defer t.mu.RUnlock()
	if n > len(t.fills) {
		n = len(t.fills)
	}
	out := make([]Fill, n)
	for i := 0; i < n; i++ {
		out[i] = t.fills[len(t.fills)-1-i]
	}
	return out
}

// TotalRealizedPnL sums realized P&L across all positions.
func (t *Tracker) TotalRealizedPnL() float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var total float64
	for _, p := range t.positions {
		total += p.RealizedPnL
	}
	return total
}

// TotalUnrealizedPnL sums unrealized P&L given a price map (tokenID → price).
func (t *Tracker) TotalUnrealizedPnL(prices map[string]float64) float64 {
	t.mu.RLock()
	defer t.mu.RUnlock()
	var total float64
	for id, p := range t.positions {
		if p.TotalSize > 0 {
			if price, ok := prices[id]; ok {
				total += p.UnrealizedPnL(price)
			}
		}
	}
	return total
}

// Summary returns a human-readable summary string.
func (t *Tracker) Summary(prices map[string]float64) string {
	realized := t.TotalRealizedPnL()
	unrealized := t.TotalUnrealizedPnL(prices)
	positions := t.AllPositions()
	fills := len(t.fills)
	return fmt.Sprintf("fills=%d  positions=%d  realized=$%.2f  unrealized=$%.2f  total=$%.2f",
		fills, len(positions), realized, unrealized, realized+unrealized)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
