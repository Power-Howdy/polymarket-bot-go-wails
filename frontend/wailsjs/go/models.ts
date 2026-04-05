export namespace app {
	
	export class BacktestRequest {
	    strategy: string;
	    ticks: number;
	    vol: number;
	    spreadBps: number;
	    orderSize: number;
	    slippage: number;
	    feeBps: number;
	    compare: boolean;
	
	    static createFrom(source: any = {}) {
	        return new BacktestRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategy = source["strategy"];
	        this.ticks = source["ticks"];
	        this.vol = source["vol"];
	        this.spreadBps = source["spreadBps"];
	        this.orderSize = source["orderSize"];
	        this.slippage = source["slippage"];
	        this.feeBps = source["feeBps"];
	        this.compare = source["compare"];
	    }
	}
	export class StrategyResult {
	    name: string;
	    pnl: number;
	    trades: number;
	    winRate: number;
	
	    static createFrom(source: any = {}) {
	        return new StrategyResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.pnl = source["pnl"];
	        this.trades = source["trades"];
	        this.winRate = source["winRate"];
	    }
	}
	export class FillSummary {
	    side: string;
	    price: number;
	    size: number;
	    cost: number;
	
	    static createFrom(source: any = {}) {
	        return new FillSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.side = source["side"];
	        this.price = source["price"];
	        this.size = source["size"];
	        this.cost = source["cost"];
	    }
	}
	export class BacktestResult {
	    strategy: string;
	    pnl: number;
	    trades: number;
	    wins: number;
	    losses: number;
	    winRate: number;
	    maxDrawdown: number;
	    equityCurve: number[];
	    fills: FillSummary[];
	    comparisons?: StrategyResult[];
	
	    static createFrom(source: any = {}) {
	        return new BacktestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.strategy = source["strategy"];
	        this.pnl = source["pnl"];
	        this.trades = source["trades"];
	        this.wins = source["wins"];
	        this.losses = source["losses"];
	        this.winRate = source["winRate"];
	        this.maxDrawdown = source["maxDrawdown"];
	        this.equityCurve = source["equityCurve"];
	        this.fills = this.convertValues(source["fills"], FillSummary);
	        this.comparisons = this.convertValues(source["comparisons"], StrategyResult);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class MarketRow {
	    question: string;
	    outcome: string;
	    bid: number;
	    ask: number;
	    mid: number;
	    spread: number;
	    volume: number;
	    signal: string;
	
	    static createFrom(source: any = {}) {
	        return new MarketRow(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.question = source["question"];
	        this.outcome = source["outcome"];
	        this.bid = source["bid"];
	        this.ask = source["ask"];
	        this.mid = source["mid"];
	        this.spread = source["spread"];
	        this.volume = source["volume"];
	        this.signal = source["signal"];
	    }
	}
	export class Position {
	    tokenId: string;
	    market: string;
	    outcome: string;
	    totalSize: number;
	    avgBuyPrice: number;
	    realizedPnL: number;
	    unrealized: number;
	
	    static createFrom(source: any = {}) {
	        return new Position(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tokenId = source["tokenId"];
	        this.market = source["market"];
	        this.outcome = source["outcome"];
	        this.totalSize = source["totalSize"];
	        this.avgBuyPrice = source["avgBuyPrice"];
	        this.realizedPnL = source["realizedPnL"];
	        this.unrealized = source["unrealized"];
	    }
	}
	export class RiskStats {
	    totalExposure: number;
	    realizedPnL: number;
	    tradesTotal: number;
	    tradesBuys: number;
	    tradesSells: number;
	
	    static createFrom(source: any = {}) {
	        return new RiskStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.totalExposure = source["totalExposure"];
	        this.realizedPnL = source["realizedPnL"];
	        this.tradesTotal = source["tradesTotal"];
	        this.tradesBuys = source["tradesBuys"];
	        this.tradesSells = source["tradesSells"];
	    }
	}
	export class Signal {
	    side: string;
	    market: string;
	    price: number;
	    size: number;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new Signal(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.side = source["side"];
	        this.market = source["market"];
	        this.price = source["price"];
	        this.size = source["size"];
	        this.reason = source["reason"];
	    }
	}
	export class StateSnapshot {
	    markets: MarketRow[];
	    positions: Position[];
	    recentSignals: Signal[];
	    risk: RiskStats;
	    uptime: string;
	    lastUpdate: string;
	    strategy: string;
	    dryRun: boolean;
	    errors: string[];
	
	    static createFrom(source: any = {}) {
	        return new StateSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.markets = this.convertValues(source["markets"], MarketRow);
	        this.positions = this.convertValues(source["positions"], Position);
	        this.recentSignals = this.convertValues(source["recentSignals"], Signal);
	        this.risk = this.convertValues(source["risk"], RiskStats);
	        this.uptime = source["uptime"];
	        this.lastUpdate = source["lastUpdate"];
	        this.strategy = source["strategy"];
	        this.dryRun = source["dryRun"];
	        this.errors = source["errors"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace config {
	
	export class RiskConfig {
	    max_exposure_usdc: number;
	    max_market_exposure_usdc: number;
	    stop_loss_usdc: number;
	
	    static createFrom(source: any = {}) {
	        return new RiskConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.max_exposure_usdc = source["max_exposure_usdc"];
	        this.max_market_exposure_usdc = source["max_market_exposure_usdc"];
	        this.stop_loss_usdc = source["stop_loss_usdc"];
	    }
	}
	export class StrategyConfig {
	    name: string;
	    market_keywords: string[];
	    max_markets: number;
	    spread_bps: number;
	    order_size_usdc: number;
	    momentum_threshold: number;
	    momentum_lookback: number;
	
	    static createFrom(source: any = {}) {
	        return new StrategyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.market_keywords = source["market_keywords"];
	        this.max_markets = source["max_markets"];
	        this.spread_bps = source["spread_bps"];
	        this.order_size_usdc = source["order_size_usdc"];
	        this.momentum_threshold = source["momentum_threshold"];
	        this.momentum_lookback = source["momentum_lookback"];
	    }
	}
	export class Config {
	    clob_host: string;
	    gamma_host: string;
	    private_key: string;
	    api_key: string;
	    api_secret: string;
	    api_passphrase: string;
	    chain_id: number;
	    strategy: StrategyConfig;
	    risk: RiskConfig;
	    poll_interval: string;
	    dry_run: boolean;
	    log_level: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.clob_host = source["clob_host"];
	        this.gamma_host = source["gamma_host"];
	        this.private_key = source["private_key"];
	        this.api_key = source["api_key"];
	        this.api_secret = source["api_secret"];
	        this.api_passphrase = source["api_passphrase"];
	        this.chain_id = source["chain_id"];
	        this.strategy = this.convertValues(source["strategy"], StrategyConfig);
	        this.risk = this.convertValues(source["risk"], RiskConfig);
	        this.poll_interval = source["poll_interval"];
	        this.dry_run = source["dry_run"];
	        this.log_level = source["log_level"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

