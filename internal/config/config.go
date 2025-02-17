package config

// Detailed Explanations:
//
// prio_level:
//  priorityLevel: Allows you to set a custom priority level for the fee. If not specified,
//  the API will use the Medium (50th percentile) level. The levels and their corresponding percentiles are:
//      Min: 0th percentile
//      Low: 25th percentile
//      Medium: 50th percentile
//      High: 75th percentile
//      VeryHigh: 95th percentile
//      UnsafeMax: 100th percentile (use with caution).
//
// legacy_not_allowed:
//  Sorted from high risk to lower risk - however all of them are still risky!
//  1. Freeze Authority Still Enabled: The developers/issuer can freeze or revert transactions,
//     which can indicate a lack of decentralization.
//  2. Single Holder Ownership: A single wallet holds a large portion of the coins, risking market manipulation.
//  3. High Holder Concentration: A few holders own a large percentage, risking price fluctuations.
//  4. Large Amount of LP Unlocked: If many liquidity pool tokens are unlocked, providers might withdraw them abruptly.
//  5. Low Liquidity: Few coins available can cause extreme price changes.
//  6. Copycat Token: A token copied from another without innovation.
//  7. Low Amount of LP Providers: Few liquidity providers can destabilize the market if they withdraw.

type LiquidityPoolConfig struct {
	RadiyumProgramID string
	WsolPcMint       string
}

type TxConfig struct {
	FetchTxMaxRetries      int // Maximum number of retries for fetching transactions
	FetchTxInitialDelay    int // Initial delay (in milliseconds) before fetching LP creation transaction details
	SwapTxInitialDelay     int // Initial delay (in milliseconds) before first buy
	GetTimeout             int // Timeout (in milliseconds) for API requests
	ConcurrentTransactions int // Number of simultaneous transactions
	RetryDelay             int // Delay (in milliseconds) between retries
}

type SwapConfig struct {
	VerboseLog                      bool
	PrioFeeMaxLamports              int    // Maximum priority fee in lamports (e.g., 1000000 = 0.001 SOL)
	PrioLevel                       string // Priority level (e.g., "veryHigh")
	Amount                          string // Swap amount (as string to preserve precision, e.g., "10000000" for 0.01 SOL)
	SlippageBps                     string // Slippage in basis points (e.g., "200" for 2%)
	DbNameTrackerHoldings           string // Sqlite Database location for tracking holdings
	TokenNotTradable400ErrorRetries int    // Number of retries if the token is not tradable yet
	TokenNotTradable400ErrorDelay   int    // Delay (in milliseconds) between retries for tradability check
}

type SellConfig struct {
	PriceSource        string // Price source identifier (e.g., "dex" for Dexscreener)
	PrioFeeMaxLamports int    // Maximum priority fee in lamports
	PrioLevel          string // Priority level (e.g., "veryHigh")
	SlippageBps        string // Slippage in basis points
	AutoSell           bool   // Automatically trigger stop loss and take profit
	StopLossPercent    int    // Stop loss percentage
	TakeProfitPercent  int    // Take profit percentage
	TrackPublicWallet  string // Public wallet address to track (if any)
}

type RugCheckConfig struct {
	VerboseLog     bool
	SimulationMode bool
	// Dangerous
	AllowMintAuthority   bool // Allow mint authority (should be false)
	AllowNotInitialized  bool // Allow uninitialized token accounts (should be false)
	AllowFreezeAuthority bool // Allow freeze authority (should be false)
	AllowRugged          bool
	// Critical
	AllowMutable                bool
	BlockReturningTokenNames    bool
	BlockReturningTokenCreators bool
	BlockSymbols                []string
	BlockNames                  []string
	AllowInsiderTopholders      bool // Allow insider accounts among top holders
	MaxAlowedPctTopholders      int  // Maximum allowed percentage that an individual top holder may have
	ExcludeLPFromTopholders     bool // Exclude Liquidity Pools from top holders check
	// Warning
	MinTotalMarkets         int
	MinTotalLPProviders     int
	MinTotalMarketLiquidity int
	// Misc
	IgnorePumpFun    bool
	MaxScore         int      // Set to 0 to ignore scoring
	LegacyNotAllowed []string // List of legacy conditions that are not allowed
}

type Config struct {
	LiquidityPool LiquidityPoolConfig
	Tx            TxConfig
	Swap          SwapConfig
	Sell          SellConfig
	RugCheck      RugCheckConfig
}

var ConfigVal = Config{
	LiquidityPool: LiquidityPoolConfig{
		RadiyumProgramID: "675kPX9MHTjS2zt1qfr1NYHuzeLXfQM9H24wFSUt1Mp8",
		WsolPcMint:       "So11111111111111111111111111111111111111112",
	},
	Tx: TxConfig{
		FetchTxMaxRetries:      10,
		FetchTxInitialDelay:    3000,  // 3 seconds
		SwapTxInitialDelay:     1000,  // 1 second
		GetTimeout:             10000, // 10 seconds
		ConcurrentTransactions: 1,
		RetryDelay:             500, // 0.5 seconds
	},
	Swap: SwapConfig{
		VerboseLog:                      false,
		PrioFeeMaxLamports:              1000000, // 0.001 SOL
		PrioLevel:                       "veryHigh",
		Amount:                          "10000000", // 0.01 SOL
		SlippageBps:                     "200",      // 2%
		DbNameTrackerHoldings:           "src/tracker/holdings.db",
		TokenNotTradable400ErrorRetries: 5,
		TokenNotTradable400ErrorDelay:   2000, // 2 seconds
	},
	Sell: SellConfig{
		PriceSource:        "dex",
		PrioFeeMaxLamports: 1000000, // 0.001 SOL
		PrioLevel:          "veryHigh",
		SlippageBps:        "200", // 2%
		AutoSell:           true,
		StopLossPercent:    10,
		TakeProfitPercent:  100,
		TrackPublicWallet:  "",
	},
	RugCheck: RugCheckConfig{
		VerboseLog:     false,
		SimulationMode: true,
		// Dangerous
		AllowMintAuthority:   false,
		AllowNotInitialized:  false,
		AllowFreezeAuthority: false,
		AllowRugged:          false,
		// Critical
		AllowMutable:                false,
		BlockReturningTokenNames:    true,
		BlockReturningTokenCreators: true,
		BlockSymbols:                []string{"XXX"},
		BlockNames:                  []string{"XXX"},
		AllowInsiderTopholders:      false,
		MaxAlowedPctTopholders:      1,
		ExcludeLPFromTopholders:     false,
		// Warning
		MinTotalMarkets:         999,
		MinTotalLPProviders:     999,
		MinTotalMarketLiquidity: 1000000,
		// Misc
		IgnorePumpFun: true,
		MaxScore:      1,
		LegacyNotAllowed: []string{
			"Low Liquidity",
			"Freeze Authority still enabled",
			"Single holder ownership",
			"High holder concentration",
			"Freeze Authority still enabled",
			"Large Amount of LP Unlocked",
			"Low Liquidity",
			"Copycat token",
			"Low amount of LP Providers",
		},
	},
}
