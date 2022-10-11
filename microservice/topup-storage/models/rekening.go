package models

import "time"

type Rekening struct {
	ReffID       string    `json:"reff_id"`
	Norek        string    `json:"norek"`
	CustomerName string    `json:"customer_name"`
	GoldBalance  float64   `json:"gold_balance"`
	CreateAt     time.Time `json:"created_at"`
}

func (Rekening) TableName() string {
	return "tbl_rekening"
}
