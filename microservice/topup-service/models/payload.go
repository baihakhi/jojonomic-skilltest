package models

type Request struct {
	GoldWeight float64 `json:"gram"`
	Amount     float64 `json:"harga"`
	Norek      string  `json:"norek"`
}

type Response struct {
	Error   bool   `json:"error"`
	ReffID  string `json:"reff_id,omitempty"`
	Message string `json:"message,omitempty"`
}
