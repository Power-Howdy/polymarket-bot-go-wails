package app

import (
	"time"

	"polymarket-bot/dashboard"
	"polymarket-bot/positions"
	"polymarket-bot/risk"
	"polymarket-bot/strategy"
)

// StateSnapshot is the JSON-serialisable view of the full dashboard state.
type StateSnapshot struct {
	Markets    []MarketRow  `json:"markets"`
	Positions  []Position   `json:"positions"`
	RecentSigs []Signal     `json:"recentSignals"`
	Risk       RiskStats    `json:"risk"`
	Uptime     string       `json:"uptime"`
	LastUpdate string       `json:"lastUpdate"`
	Strategy   string       `json:"strategy"`
	DryRun     bool         `json:"dryRun"`
	Errors     []string     `json:"errors"`
}

// MarketRow is the frontend market table row.
type MarketRow struct {
	Question string  `json:"question"`
	Outcome  string  `json:"outcome"`
	Bid      float64 `json:"bid"`
	Ask      float64 `json:"ask"`
	Mid      float64 `json:"mid"`
	Spread   float64 `json:"spread"`
	Volume   float64 `json:"volume"`
	Signal   string  `json:"signal"`
}

// Position is the frontend position row.
type Position struct {
	TokenID     string  `json:"tokenId"`
	Market      string  `json:"market"`
	Outcome     string  `json:"outcome"`
	TotalSize   float64 `json:"totalSize"`
	AvgBuyPrice float64 `json:"avgBuyPrice"`
	RealizedPnL float64 `json:"realizedPnL"`
	Unrealized  float64 `json:"unrealized"`
}

// Signal is the frontend signal row.
type Signal struct {
	Side   string  `json:"side"`
	Market string  `json:"market"`
	Price  float64 `json:"price"`
	Size   float64 `json:"size"`
	Reason string  `json:"reason"`
}

// RiskStats is the frontend risk summary.
type RiskStats struct {
	TotalExposure float64 `json:"totalExposure"`
	RealizedPnL   float64 `json:"realizedPnL"`
	TradesTotal   int     `json:"tradesTotal"`
	TradesBuys    int     `json:"tradesBuys"`
	TradesSells   int     `json:"tradesSells"`
}

func snapshotFrom(s *dashboard.State) *StateSnapshot {
	s.Mu().RLock()
	defer s.Mu().RUnlock()

	snap := &StateSnapshot{
		Strategy:   s.Strategy,
		DryRun:     s.DryRun,
		Errors:     s.Errors,
		Uptime:     time.Since(s.StartTime).Round(time.Second).String(),
		LastUpdate: s.LastUpdate.Format("15:04:05"),
	}

	for _, m := range s.Markets {
		snap.Markets = append(snap.Markets, marketRowFrom(m))
	}
	for _, p := range s.Positions {
		unreal := 0.0
		if price, ok := s.Prices[p.TokenID]; ok {
			unreal = p.UnrealizedPnL(price)
		}
		snap.Positions = append(snap.Positions, positionFrom(p, unreal))
	}
	for _, sig := range s.RecentSigs {
		snap.RecentSigs = append(snap.RecentSigs, signalFrom(sig))
	}
	snap.Risk = riskStatsFrom(s.RiskStats)
	return snap
}

func marketRowFrom(m dashboard.MarketRow) MarketRow {
	return MarketRow{
		Question: m.Question, Outcome: m.Outcome,
		Bid: m.Bid, Ask: m.Ask, Mid: m.Mid,
		Spread: m.Spread, Volume: m.Volume, Signal: m.Signal,
	}
}

func positionFrom(p positions.Position, unreal float64) Position {
	return Position{
		TokenID: p.TokenID, Market: p.Market, Outcome: p.Outcome,
		TotalSize: p.TotalSize, AvgBuyPrice: p.AvgBuyPrice,
		RealizedPnL: p.RealizedPnL, Unrealized: unreal,
	}
}

func signalFrom(s strategy.Signal) Signal {
	return Signal{
		Side: s.Side, Market: s.Market,
		Price: s.Price, Size: s.Size, Reason: s.Reason,
	}
}

func riskStatsFrom(r risk.Stats) RiskStats {
	return RiskStats{
		TotalExposure: r.TotalExposure, RealizedPnL: r.RealizedPnL,
		TradesTotal: r.TradesTotal, TradesBuys: r.TradesBuys, TradesSells: r.TradesSells,
	}
}
