package bot

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"polymarket-bot/client"
	"polymarket-bot/config"
	"polymarket-bot/dashboard"
	"polymarket-bot/positions"
	"polymarket-bot/risk"
	"polymarket-bot/strategy"
)

// Bot is the main trading engine.
type Bot struct {
	cfg       *config.Config
	client    *client.Client
	strategy  strategy.Strategy
	risk      *risk.Manager
	tracker   *positions.Tracker
	dash      *dashboard.Dashboard
	dashState *dashboard.State
	onUpdate  func() // optional callback after each tick

	historyMu sync.Mutex
	history   map[string][]float64

	stopCh chan struct{}
}

// New creates a bot from config.
func New(cfg *config.Config) (*Bot, error) {
	return NewWithCallback(cfg, nil)
}

// NewWithCallback creates a bot that calls onUpdate after every tick.
func NewWithCallback(cfg *config.Config, onUpdate func()) (*Bot, error) {
	c := client.New(
		cfg.CLOBHost,
		cfg.GammaHost,
		cfg.APIKey,
		cfg.APISecret,
		cfg.APIPassphrase,
	)

	strat, err := strategy.New(
		cfg.Strategy.Name,
		cfg.Strategy.SpreadBps,
		cfg.Strategy.OrderSizeUSDC,
		cfg.Strategy.MomentumThreshold,
		cfg.Strategy.MomentumLookback,
	)
	if err != nil {
		return nil, err
	}

	rm := risk.New(
		cfg.Risk.MaxExposureUSDC,
		cfg.Risk.MaxMarketExposureUSDC,
		cfg.Risk.StopLossUSDC,
	)

	tracker := positions.New()
	state := dashboard.NewState(strat.Name(), cfg.DryRun)
	dash := dashboard.New(state)

	return &Bot{
		cfg:       cfg,
		client:    c,
		strategy:  strat,
		risk:      rm,
		tracker:   tracker,
		dash:      dash,
		dashState: state,
		onUpdate:  onUpdate,
		history:   make(map[string][]float64),
		stopCh:    make(chan struct{}),
	}, nil
}

// State returns the live dashboard state (for the Wails app bridge).
func (b *Bot) State() *dashboard.State { return b.dashState }

// Run starts the main trading loop.
func (b *Bot) Run() error {
	log.Printf("🚀 Bot starting — strategy=%s  poll=%s  dry_run=%v",
		b.strategy.Name(), b.cfg.PollInterval, b.cfg.DryRun)

	b.dash.Start()
	defer b.dash.Stop()

	ticker := time.NewTicker(b.cfg.PollInterval)
	defer ticker.Stop()

	if err := b.tick(); err != nil {
		log.Printf("⚠️  tick error: %v", err)
		b.dashState.AddError(err.Error())
	}

	for {
		select {
		case <-ticker.C:
			if err := b.tick(); err != nil {
				log.Printf("⚠️  tick error: %v", err)
				b.dashState.AddError(err.Error())
			}
		case <-b.stopCh:
			return nil
		}
	}
}

// Stop signals the bot to shut down.
func (b *Bot) Stop() {
	close(b.stopCh)
}

// tick performs one poll-and-act cycle.
func (b *Bot) tick() error {
	markets, err := b.client.GetMarkets(b.cfg.Strategy.MarketKeywords, b.cfg.Strategy.MaxMarkets*3)
	if err != nil {
		return fmt.Errorf("fetching markets: %w", err)
	}

	markets = filterMarkets(markets, b.cfg.Strategy.MaxMarkets)
	if len(markets) == 0 {
		return nil
	}

	var rows []dashboard.MarketRow
	prices := make(map[string]float64)

	for i := range markets {
		m := &markets[i]
		row, tokenID, price := b.processMarket(m)
		if row != nil {
			rows = append(rows, *row)
		}
		if tokenID != "" {
			prices[tokenID] = price
		}
	}

	b.dashState.UpdateMarkets(rows)
	b.dashState.UpdatePositions(b.tracker.AllPositions())
	b.dashState.UpdateRisk(b.risk.Stats())
	for k, v := range prices {
		b.dashState.Prices[k] = v
	}

	if b.onUpdate != nil {
		b.onUpdate()
	}

	return nil
}

// processMarket analyzes a single market and acts on signals.
func (b *Bot) processMarket(m *client.GammaMarket) (*dashboard.MarketRow, string, float64) {
	clobMarket, err := b.client.GetCLOBMarket(m.ConditionID)
	if err != nil {
		b.dashState.AddError(fmt.Sprintf("%q: %v", truncate(m.Question, 30), err))
		return nil, "", 0
	}

	var yesToken *client.Token
	for i := range clobMarket.Tokens {
		t := &clobMarket.Tokens[i]
		if strings.EqualFold(t.Outcome, "yes") {
			yesToken = t
			break
		}
	}
	if yesToken == nil && len(clobMarket.Tokens) > 0 {
		yesToken = &clobMarket.Tokens[0]
	}
	if yesToken == nil {
		return nil, "", 0
	}

	ob, err := b.client.GetOrderBook(yesToken.TokenID)
	if err != nil {
		b.dashState.AddError(fmt.Sprintf("orderbook %q: %v", truncate(m.Question, 30), err))
		return nil, "", 0
	}

	b.appendHistory(yesToken.TokenID, ob.MidPrice())
	hist := b.getHistory(yesToken.TokenID)

	snap := strategy.MarketSnapshot{
		Market:    *m,
		OrderBook: ob,
		TokenID:   yesToken.TokenID,
		Outcome:   yesToken.Outcome,
	}

	sig := b.strategy.Analyze(snap, hist)

	row := &dashboard.MarketRow{
		Question: m.Question,
		Outcome:  yesToken.Outcome,
		Bid:      ob.BestBid(),
		Ask:      ob.BestAsk(),
		Mid:      ob.MidPrice(),
		Spread:   ob.Spread(),
		Volume:   float64(m.Volume),
		Signal:   sig.Side,
	}
	if sig.Side != "NONE" {
		row.Signal = fmt.Sprintf("%s @$%.3f", sig.Side, sig.Price)
	}

	b.dashState.AddSignal(sig)

	if sig.Side == "NONE" {
		return row, yesToken.TokenID, ob.MidPrice()
	}

	if err := b.risk.Check(sig); err != nil {
		b.dashState.AddError(fmt.Sprintf("risk: %v", err))
		return row, yesToken.TokenID, ob.MidPrice()
	}

	b.execute(sig, yesToken.Outcome)
	return row, yesToken.TokenID, ob.MidPrice()
}

// execute submits (or simulates) an order and records the fill.
func (b *Bot) execute(sig strategy.Signal, outcome string) {
	if b.cfg.DryRun {
		side := positions.Buy
		if sig.Side == "SELL" {
			side = positions.Sell
		}
		fill := b.tracker.RecordFill(sig.TokenID, sig.Market, outcome, side, sig.Price, sig.Size)
		log.Printf("📝 DRY RUN: %s", fill)
		b.risk.RecordFill(sig, sig.Price)
		return
	}

	if b.cfg.APIKey == "" {
		log.Printf("ℹ️  no API key — skipping live order")
		return
	}

	req := &client.OrderRequest{
		TokenID:   sig.TokenID,
		Price:     sig.Price,
		Size:      sig.Size,
		Side:      sig.Side,
		OrderType: "GTC",
	}

	resp, err := b.client.PlaceOrder(req)
	if err != nil {
		log.Printf("❌ order failed: %v", err)
		b.dashState.AddError(fmt.Sprintf("order: %v", err))
		return
	}

	log.Printf("✅ order placed: id=%s", resp.OrderID)
	side := positions.Buy
	if sig.Side == "SELL" {
		side = positions.Sell
	}
	b.tracker.RecordFill(sig.TokenID, sig.Market, outcome, side, sig.Price, sig.Size)
	b.risk.RecordFill(sig, sig.Price)
}

// ─── History helpers ──────────────────────────────────────────────────────────

const maxHistory = 100

func (b *Bot) appendHistory(tokenID string, price float64) {
	b.historyMu.Lock()
	defer b.historyMu.Unlock()
	h := b.history[tokenID]
	h = append(h, price)
	if len(h) > maxHistory {
		h = h[len(h)-maxHistory:]
	}
	b.history[tokenID] = h
}

func (b *Bot) getHistory(tokenID string) []float64 {
	b.historyMu.Lock()
	defer b.historyMu.Unlock()
	cp := make([]float64, len(b.history[tokenID]))
	copy(cp, b.history[tokenID])
	return cp
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func filterMarkets(markets []client.GammaMarket, max int) []client.GammaMarket {
	var out []client.GammaMarket
	for _, m := range markets {
		if m.Active && !m.Closed && m.Volume > 0 {
			out = append(out, m)
			if len(out) >= max {
				break
			}
		}
	}
	return out
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
