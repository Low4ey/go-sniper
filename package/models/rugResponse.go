package models

type RugResponse struct {
	Mint         string `json:"mint"`
	TokenProgram string `json:"tokenProgram"`
	Creator      string `json:"creator"`
	Token        struct {
		MintAuthority   interface{} `json:"mintAuthority"`
		Supply          int         `json:"supply"`
		Decimals        int         `json:"decimals"`
		IsInitialized   bool        `json:"isInitialized"`
		FreezeAuthority interface{} `json:"freezeAuthority"`
	} `json:"token"`
	TokenExtensions interface{} `json:"token_extensions"`
	TokenMeta       struct {
		Name            string `json:"name"`
		Symbol          string `json:"symbol"`
		Uri             string `json:"uri"`
		Mutable         bool   `json:"mutable"`
		UpdateAuthority string `json:"updateAuthority"`
	} `json:"tokenMeta"`
	TopHolders []struct {
		Address        string `json:"address"`
		Amount         int    `json:"amount"`
		Decimals       int    `json:"decimals"`
		Pct            int    `json:"pct"`
		UiAmount       int    `json:"uiAmount"`
		UiAmountString string `json:"uiAmountString"`
		Owner          string `json:"owner"`
		Insider        bool   `json:"insider"`
	} `json:"topHolders"`
	FreezeAuthority interface{} `json:"freezeAuthority"`
	MintAuthority   interface{} `json:"mintAuthority"`
	Risks           []struct {
		Name        string `json:"name"`
		Value       string `json:"value"`
		Description string `json:"description"`
		Score       int    `json:"score"`
		Level       string `json:"level"`
	} `json:"risks"`
	Score    int `json:"score"`
	FileMeta struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		Symbol      string `json:"symbol"`
		Image       string `json:"image"`
	} `json:"fileMeta"`
	LockerOwners map[string]interface{} `json:"lockerOwners"`
	Lockers      map[string]interface{} `json:"lockers"`
	LpLockers    interface{}            `json:"lpLockers"`
	Markets      []struct {
		Pubkey     string `json:"pubkey"`
		MarketType string `json:"marketType"`
		MintA      string `json:"mintA"`
		MintB      string `json:"mintB"`
		MintLP     string `json:"mintLP"`
		LiquidityA string `json:"liquidityA"`
		LiquidityB string `json:"liquidityB"`
	} `json:"markets"`
	TotalMarketLiquidity int  `json:"totalMarketLiquidity"`
	TotalLPProviders     int  `json:"totalLPProviders"`
	Rugged               bool `json:"rugged"`
}
