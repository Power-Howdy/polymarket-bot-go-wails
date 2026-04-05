// Package app exposes the Wails application methods bound to the frontend.
package app

import (
	"context"
	"fmt"
	"log"
	"time"

	"polymarket-bot/bot"
	"polymarket-bot/config"
	"polymarket-bot/dashboard"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the Wails application context.
type App struct {
	ctx       context.Context
	bot       *bot.Bot
	dashState *dashboard.State
	cfg       *config.Config
	running   bool
}

// New creates a new App instance.
func New() *App {
	return &App{}
}

// Startup is called by Wails when the app starts.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
}

// LoadConfig loads config from a file path.
func (a *App) LoadConfig(path string) (*config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		return nil, fmt.Errorf("loading config: %w", err)
	}
	a.cfg = cfg
	return cfg, nil
}

// StartBot starts the trading bot with the given config.
// cfg is received as JSON from the frontend; PollInterval must be re-parsed from PollIntervalStr.
func (a *App) StartBot(cfg config.Config) error {
	if a.running {
		return fmt.Errorf("bot is already running")
	}
	// Re-parse the duration string that JSON cannot carry as time.Duration.
	if cfg.PollIntervalStr != "" {
		d, err := time.ParseDuration(cfg.PollIntervalStr)
		if err != nil {
			return fmt.Errorf("invalid poll_interval %q: %w", cfg.PollIntervalStr, err)
		}
		cfg.PollInterval = d
	} else {
		cfg.PollInterval = 10 * time.Second
	}
	a.cfg = &cfg
	b, err := bot.NewWithCallback(&cfg, a.onStateUpdate)
	if err != nil {
		return fmt.Errorf("creating bot: %w", err)
	}
	a.bot = b
	a.dashState = b.State()
	a.running = true
	go func() {
		if err := b.Run(); err != nil {
			log.Printf("bot error: %v", err)
			runtime.EventsEmit(a.ctx, "bot:error", err.Error())
		}
		a.running = false
	}()
	return nil
}

// StopBot stops the running bot.
func (a *App) StopBot() {
	if a.bot != nil && a.running {
		a.bot.Stop()
		a.running = false
	}
}

// IsRunning returns whether the bot is active.
func (a *App) IsRunning() bool {
	return a.running
}

// GetState returns the current dashboard state snapshot.
func (a *App) GetState() *StateSnapshot {
	if a.dashState == nil {
		return &StateSnapshot{}
	}
	return snapshotFrom(a.dashState)
}

// onStateUpdate is called after each tick to push state to the frontend.
func (a *App) onStateUpdate() {
	if a.dashState == nil {
		return
	}
	snap := snapshotFrom(a.dashState)
	runtime.EventsEmit(a.ctx, "state:update", snap)
}

// RunBacktest runs a backtest and returns the result.
func (a *App) RunBacktest(req BacktestRequest) (*BacktestResult, error) {
	return runBacktest(req)
}

// SelectConfigFile opens a native file picker and returns the chosen path.
func (a *App) SelectConfigFile() (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "Select Config File",
		Filters: []runtime.FileFilter{
			{DisplayName: "JSON / YAML", Pattern: "*.json;*.yaml;*.yml"},
		},
	})
	if err != nil {
		return "", err
	}
	return path, nil
}
