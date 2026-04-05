package risk

import (
	"fmt"
	"sync"

	"polymarket-bot/strategy"
)

// Manager tracks exposure and P&L, gating order signals.
type Manager struct {
	mu sync.Mutex

	maxTotalExposure  float64
	maxMarketExposure float64
	stopLoss          float64

	totalExposure    float64
	marketExposure   map[string]float64
	realizedPnL      float64
	unrealizedPnL    float64
	tradesTotal      int
	tradesBuys       int
	tradesSells      int
}

// New creates a risk manager.
func New(maxTotal, maxMarket, stopLoss float64) *Manager {
	return &Manager{
		maxTotalExposure:  maxTotal,
		maxMarketExposure: maxMarket,
		stopLoss:          stopLoss,
		marketExposure:    make(map[string]float64),
	}
}

// Check returns an error if the signal would violate risk limits.
func (r *Manager) Check(sig strategy.Signal) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if sig.Side == "NONE" {
		return nil
	}

	if r.realizedPnL < r.stopLoss {
		return fmt.Errorf("stop-loss triggered (P&L $%.2f < limit $%.2f)", r.realizedPnL, r.stopLoss)
	}

	orderValue := sig.Price * sig.Size
	if sig.Side == "BUY" {
		newTotal := r.totalExposure + orderValue
		if newTotal > r.maxTotalExposure {
			return fmt.Errorf("total exposure limit: $%.2f + $%.2f > $%.2f",
				r.totalExposure, orderValue, r.maxTotalExposure)
		}
		mkt := r.marketExposure[sig.Market]
		if mkt+orderValue > r.maxMarketExposure {
			return fmt.Errorf("market exposure limit for %q: $%.2f + $%.2f > $%.2f",
				sig.Market, mkt, orderValue, r.maxMarketExposure)
		}
	}

	return nil
}

// RecordFill updates internal state after an order is filled.
func (r *Manager) RecordFill(sig strategy.Signal, fillPrice float64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.tradesTotal++
	cost := fillPrice * sig.Size

	switch sig.Side {
	case "BUY":
		r.tradesBuys++
		r.totalExposure += cost
		r.marketExposure[sig.Market] += cost
	case "SELL":
		r.tradesSells++
		r.totalExposure -= cost
		r.marketExposure[sig.Market] -= cost
		// Simplified P&L: assume we bought at sig.Price, sold at fillPrice
		r.realizedPnL += (fillPrice - sig.Price) * sig.Size
	}
}

// Stats returns a snapshot of current risk metrics.
func (r *Manager) Stats() Stats {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Stats{
		TotalExposure: r.totalExposure,
		RealizedPnL:   r.realizedPnL,
		TradesTotal:   r.tradesTotal,
		TradesBuys:    r.tradesBuys,
		TradesSells:   r.tradesSells,
	}
}

// Stats is a snapshot of risk metrics.
type Stats struct {
	TotalExposure float64
	RealizedPnL   float64
	TradesTotal   int
	TradesBuys    int
	TradesSells   int
}

func (s Stats) String() string {
	return fmt.Sprintf("exposure=$%.2f  P&L=$%.2f  trades=%d (B:%d S:%d)",
		s.TotalExposure, s.RealizedPnL, s.TradesTotal, s.TradesBuys, s.TradesSells)
}
