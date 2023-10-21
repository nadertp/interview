package prices

import "encoding/json"

type Ticker struct {
	Exchange             string
	Stream_name          string
	Symbol               string
	Exchange_time        int64
	Update_time          int64
	Price_change         float64
	Price_change_percent float64
	Last_price           float64
}

func (t Ticker) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Exchange             string
		Stream_name          string
		Symbol               string
		Exchange_time        int64
		Update_time          int64
		Price_change         float64
		Price_change_percent float64
		Last_price           float64
	}{
		Exchange:             "okx_spot",
		Stream_name:          t.Stream_name,
		Symbol:               t.Symbol,
		Exchange_time:        t.Exchange_time,
		Update_time:          t.Update_time,
		Price_change:         t.Price_change,
		Price_change_percent: t.Price_change_percent,
		Last_price:           t.Last_price,
	})
}
