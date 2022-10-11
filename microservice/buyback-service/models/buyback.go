package models

type Params struct {
	GoldWeight         float64 `json:"gram"`
	Amount             int64   `json:"harga"`
	Norek              string  `json:"norek"`
	ReffID             string  `json:"reff_id,omitempty"`
	HargaTopup         int64   `json:"harga_topup"`
	HargaBuyback       int64   `json:"harga_buyback"`
	CurrentGoldBalance float64 `json:"current_gold_balance"`
}