package models

type HoldingRecord struct {
	ID               *int    `json:"id,omitempty"`
	Time             int     `json:"time"`
	Token            string  `json:"token"`
	TokenName        string  `json:"tokenName"`
	Balance          float64 `json:"balance"`
	SolPaid          float64 `json:"solPaid"`
	SolFeePaid       float64 `json:"solFeePaid"`
	SolPaidUSDC      float64 `json:"solPaidUSDC"`
	SolFeePaidUSDC   float64 `json:"solFeePaidUSDC"`
	PerTokenPaidUSDC float64 `json:"perTokenPaidUSDC"`
	Slot             int     `json:"slot"`
	Program          string  `json:"program"`
}
