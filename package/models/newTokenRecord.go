package models

type NewTokenRecord struct {
	ID      *int   `json:"id,omitempty"`
	Time    int    `json:"time"`
	Name    string `json:"name"`
	Mint    string `json:"mint"`
	Creator string `json:"creator"`
}
