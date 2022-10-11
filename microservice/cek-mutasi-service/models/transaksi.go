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

type ListTransaksi []*Transaksi

func (lt ListTransaksi) ToResponseItems() []*TransaksiItem {
	list := make([]*TransaksiItem, len(lt))
	for i, v := range lt {
		list[i] = &TransaksiItem{
			CreatedAt:    int32(v.CreatedAt),
			Type:         v.Type,
			GoldWeight:   v.GoldWeight,
			GoldBalance:  v.GoldBalance,
			HargaTopup:   v.HargaTopup,
			HargaBuyback: v.HargaBuyback,
		}
	}

	return list
}
