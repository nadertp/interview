package prices

import (
	"encoding/json"
)

type BookTicker struct {
	Exchange       string
	Stream_name    string
	Symbol         string
	Exchange_time  int64
	Update_time    int64
	Best_bid_price float64
	Best_bid_qty   float64
	Best_ask_price float64
	Best_ask_qty   float64
}

func (b BookTicker) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Exchange       string
		Stream_name    string
		Symbol         string
		Exchange_time  int64
		Update_time    int64
		Best_bid_price float64
		Best_bid_qty   float64
		Best_ask_price float64
		Best_ask_qty   float64
	}{
		Exchange:       "okx_spot",
		Stream_name:    b.Stream_name,
		Symbol:         b.Symbol,
		Exchange_time:  b.Exchange_time,
		Update_time:    b.Update_time,
		Best_bid_price: b.Best_bid_price,
		Best_bid_qty:   b.Best_bid_qty,
		Best_ask_price: b.Best_ask_price,
		Best_ask_qty:   b.Best_ask_qty,
	})
}
