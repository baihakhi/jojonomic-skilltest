package models

import "time"

type Harga struct {
	AdminId      string    `json:"admin_id"`
	ReffId       string    `json:"reff_id"`
	HargaTopup   float64   `json:"harga_topup"`
	HargaBuyback float64   `json:"harga_buyback"`
	CreatedAt    time.Time `json:"created_at"`
}

func (Harga) TableName() string {
	return "tbl_harga"
}
