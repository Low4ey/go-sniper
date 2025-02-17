package models

type LastPriceDexResponse struct {
	SchemaVersion string `json:"schemaVersion"`
	Pairs         []Pair `json:"pairs"`
}

type Pair struct {
	ChainID       string       `json:"chainId"`
	DexID         string       `json:"dexId"`
	URL           string       `json:"url"`
	PairAddress   string       `json:"pairAddress"`
	Labels        []string     `json:"labels,omitempty"`
	BaseToken     Token        `json:"baseToken"`
	QuoteToken    Token        `json:"quoteToken"`
	PriceNative   string       `json:"priceNative"`
	PriceUSD      string       `json:"priceUsd"`
	Txns          Transactions `json:"txns"`
	Volume        Volume       `json:"volume"`
	PriceChange   PriceChange  `json:"priceChange"`
	Liquidity     Liquidity    `json:"liquidity"`
	FDV           float64      `json:"fdv"`
	MarketCap     float64      `json:"marketCap"`
	PairCreatedAt int64        `json:"pairCreatedAt"`
	Info          Info         `json:"info"`
}

type Token struct {
	Address string `json:"address"`
	Name    string `json:"name"`
	Symbol  string `json:"symbol"`
}

type Transactions struct {
	M5  TxnDetails `json:"m5"`
	H1  TxnDetails `json:"h1"`
	H6  TxnDetails `json:"h6"`
	H24 TxnDetails `json:"h24"`
}

type TxnDetails struct {
	Buys  int `json:"buys"`
	Sells int `json:"sells"`
}

type Volume struct {
	H24 float64 `json:"h24"`
	H6  float64 `json:"h6"`
	H1  float64 `json:"h1"`
	M5  float64 `json:"m5"`
}

type PriceChange struct {
	M5  float64 `json:"m5"`
	H1  float64 `json:"h1"`
	H6  float64 `json:"h6"`
	H24 float64 `json:"h24"`
}

type Liquidity struct {
	USD   float64 `json:"usd"`
	Base  float64 `json:"base"`
	Quote float64 `json:"quote"`
}

type Info struct {
	ImageURL  string    `json:"imageUrl"`
	Header    string    `json:"header"`
	OpenGraph string    `json:"openGraph"`
	Websites  []Website `json:"websites,omitempty"`
	Socials   []Social  `json:"socials"`
}

type Website struct {
	Label string `json:"label"`
	URL   string `json:"url"`
}

type Social struct {
	Type string `json:"type"`
	URL  string `json:"url"`
}
