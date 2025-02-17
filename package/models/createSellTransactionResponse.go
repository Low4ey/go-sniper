package models

type CreateSellTransactionResponse struct {
	Success bool    `json:"success"`
	Msg     *string `json:"msg,omitempty"`
	Tx      *string `json:"tx,omitempty"`
}
