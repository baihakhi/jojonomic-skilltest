package models

type Request struct {
	Norek     string `json:"norek"`
	StartDate int32  `json:"start_date"`
	EndDate   int32  `json:"end_date"`
}

type TransaksiItem struct {
	CreatedAt    int32   `json:"date"`
	Type         string  `json:"type"`
	GoldWeight   float64 `json:"gram"`
	HargaTopup   float64 `json:"harga_topup"`
	HargaBuyback float64 `json:"harga_buyback"`
	GoldBalance  float64 `json:"saldo"`
}

type Response struct {
	Error   bool             `json:"error"`
	Data    []*TransaksiItem `json:"data,omitempty"`
	Message string           `json:"message,omitempty"`
}
