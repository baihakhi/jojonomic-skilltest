package models

type Request struct {
	AdminID      string `json:"admin_id"`
	ReffID       string `json:"reff_id,omitempty"`
	HargaTopup   int64  `json:"harga_topup"`
	HargaBuyback int64  `json:"harga_buyback"`
}

type Response struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
	ReffId  string `json:"reff_id,omitempty"`
}

// func (r *Response) GetErrorResponse(reffID string, message string)
