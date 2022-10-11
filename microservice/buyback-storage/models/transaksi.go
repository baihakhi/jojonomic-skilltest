package models

type Transaksi struct {
	ReffID       string  `json:"reff_id"`
	Norek        string  `json:"norek"`
	Type         string  `json:"type"`
	GoldWeight   float64 `json:"gold_weight"`
	HargaTopup   float64 `json:"harga_topup"`
	HargaBuyback float64 `json:"harga_buyback"`
	GoldBalance  float64 `json:"gold_balance"`
	CreatedAt    int     `json:"created_at"`
}

func (Transaksi) TableName() string {
	return "tbl_transaksi"
}
