// Package dashboard renders a live terminal UI using only the standard library.
package dashboard

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"polymarket-bot/positions"
	"polymarket-bot/risk"
	"polymarket-bot/strategy"
)

// MarketRow is a single row in the markets table.
type MarketRow struct {
	Question string
	Outcome  string
	Bid      float64
	Ask      float64
	Mid      float64
	Spread   float64
	Volume   float64
	Signal   string
}

// State holds all data the dashboard displays.
type State struct {
	mu         sync.RWMutex
	Markets    []MarketRow
	Positions  []positions.Position
	RecentSigs []strategy.Signal
	RiskStats  risk.Stats
	Prices     map[string]float64 // tokenID → current price
	StartTime  time.Time
	LastUpdate time.Time
	Strategy   string
	DryRun     bool
	Errors     []string
}

// Mu exposes the read-write mutex for external snapshot consumers.
func (s *State) Mu() *sync.RWMutex { return &s.mu }

// NewState creates an empty dashboard state.
func NewState(stratName string, dryRun bool) *State {
	return &State{
		Prices:    make(map[string]float64),
		StartTime: time.Now(),
		Strategy:  stratName,
		DryRun:    dryRun,
	}
}

// UpdateMarkets replaces the market rows.
func (s *State) UpdateMarkets(rows []MarketRow) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Markets = rows
	s.LastUpdate = time.Now()
}

// UpdatePositions replaces the positions list.
func (s *State) UpdatePositions(pos []positions.Position) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Positions = pos
}

// AddSignal prepends a signal to the recent signals list (capped at 8).
func (s *State) AddSignal(sig strategy.Signal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RecentSigs = append([]strategy.Signal{sig}, s.RecentSigs...)
	if len(s.RecentSigs) > 8 {
		s.RecentSigs = s.RecentSigs[:8]
	}
}

// UpdateRisk updates the risk stats snapshot.
func (s *State) UpdateRisk(stats risk.Stats) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.RiskStats = stats
}

// AddError records an error message (capped at 5).
func (s *State) AddError(msg string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Errors = append([]string{msg}, s.Errors...)
	if len(s.Errors) > 5 {
		s.Errors = s.Errors[:5]
	}
}

// ─── Renderer ─────────────────────────────────────────────────────────────────

// Dashboard renders the state to a terminal.
type Dashboard struct {
	state  *State
	out    io.Writer
	stopCh chan struct{}
	width  int
}

// New creates a new Dashboard writing to stdout.
func New(state *State) *Dashboard {
	return &Dashboard{
		state:  state,
		out:    os.Stdout,
		stopCh: make(chan struct{}),
		width:  100,
	}
}

// Start begins rendering in a background goroutine (refresh every second).
func (d *Dashboard) Start() {
	go d.loop()
}

// Stop halts rendering.
func (d *Dashboard) Stop() {
	close(d.stopCh)
}

func (d *Dashboard) loop() {
	t := time.NewTicker(time.Second)
	defer t.Stop()
	for {
		select {
		case <-t.C:
			d.Render()
		case <-d.stopCh:
			return
		}
	}
}

// Render writes a full dashboard frame to the output.
func (d *Dashboard) Render() {
	d.state.mu.RLock()
	defer d.state.mu.RUnlock()

	s := d.state
	w := d.width
	out := d.out

	// Clear screen (ANSI)
	fmt.Fprint(out, "\033[H\033[2J")

	// ── Header ────────────────────────────────────────────────────────────────
	uptime := time.Since(s.StartTime).Round(time.Second)
	updated := time.Since(s.LastUpdate).Round(time.Second)
	dryTag := ""
	if s.DryRun {
		dryTag = "  ⚠️  DRY RUN"
	}
	fmt.Fprintln(out, box("═", fmt.Sprintf(
		"  🤖 Polymarket Bot  │  strategy: %-10s │  uptime: %-8s │  updated: %s ago%s  ",
		s.Strategy, uptime, updated, dryTag,
	), w))

	// ── Markets table ─────────────────────────────────────────────────────────
	fmt.Fprintln(out, header("MARKETS", w))
	if len(s.Markets) == 0 {
		fmt.Fprintln(out, "  (waiting for market data…)")
	} else {
		fmt.Fprintf(out, "  %-44s %-6s %-6s %-6s %-8s %-10s  %s\n",
			"QUESTION", "BID", "ASK", "MID", "SPREAD", "VOL($)", "SIGNAL")
		fmt.Fprintln(out, "  "+strings.Repeat("─", w-4))
		for _, m := range s.Markets {
			q := truncate(m.Question, 44)
			sig := truncate(m.Signal, 22)
			fmt.Fprintf(out, "  %-44s %-6.3f %-6.3f %-6.3f %-8.4f %-10.0f  %s\n",
				q, m.Bid, m.Ask, m.Mid, m.Spread, m.Volume, sig)
		}
	}

	// ── Positions table ───────────────────────────────────────────────────────
	fmt.Fprintln(out, header("OPEN POSITIONS", w))
	if len(s.Positions) == 0 {
		fmt.Fprintln(out, "  (no open positions)")
	} else {
		fmt.Fprintf(out, "  %-46s %-8s %-10s %-12s  %s\n",
			"MARKET (OUTCOME)", "SIZE", "AVG COST", "REALIZED PnL", "UNREALIZED")
		fmt.Fprintln(out, "  "+strings.Repeat("─", w-4))
		for _, p := range s.Positions {
			label := truncate(p.Market+" ("+p.Outcome+")", 46)
			unreal := 0.0
			if price, ok := s.Prices[p.TokenID]; ok {
				unreal = p.UnrealizedPnL(price)
			}
			fmt.Fprintf(out, "  %-46s %-8.2f %-10.4f $%-11.4f  $%.4f\n",
				label, p.TotalSize, p.AvgBuyPrice, p.RealizedPnL, unreal)
		}
	}

	// ── Recent signals ────────────────────────────────────────────────────────
	fmt.Fprintln(out, header("RECENT SIGNALS", w))
	if len(s.RecentSigs) == 0 {
		fmt.Fprintln(out, "  (no signals yet)")
	} else {
		for _, sig := range s.RecentSigs {
			icon := "  ·"
			switch sig.Side {
			case "BUY":
				icon = "  🟢"
			case "SELL":
				icon = "  🔴"
			}
			fmt.Fprintf(out, "%s %s\n", icon, truncate(sig.String(), w-5))
		}
	}

	// ── Risk summary ──────────────────────────────────────────────────────────
	fmt.Fprintln(out, header("RISK", w))
	rs := s.RiskStats
	pnlSign := "+"
	if rs.RealizedPnL < 0 {
		pnlSign = ""
	}
	fmt.Fprintf(out, "  Exposure: $%-8.2f  P&L: %s$%-8.2f  Trades: %d  (B:%d / S:%d)\n",
		rs.TotalExposure, pnlSign, rs.RealizedPnL, rs.TradesTotal, rs.TradesBuys, rs.TradesSells)

	// ── Errors ────────────────────────────────────────────────────────────────
	if len(s.Errors) > 0 {
		fmt.Fprintln(out, header("RECENT ERRORS", w))
		for _, e := range s.Errors {
			fmt.Fprintf(out, "  ⚠️  %s\n", truncate(e, w-8))
		}
	}

	fmt.Fprintln(out, strings.Repeat("═", w))
	fmt.Fprintln(out, "  Press Ctrl+C to stop")
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func header(title string, w int) string {
	line := fmt.Sprintf("  ── %s ", title)
	if len(line) < w {
		line += strings.Repeat("─", w-len(line))
	}
	return "\n" + line
}

func box(ch, content string, w int) string {
	top := strings.Repeat(ch, w)
	mid := content
	if len(mid) < w-2 {
		mid = mid + strings.Repeat(" ", w-2-len(mid))
	}
	return top + "\n" + mid + "\n" + strings.Repeat(ch, w)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
