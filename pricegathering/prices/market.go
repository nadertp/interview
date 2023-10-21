package prices

import (
	"encoding/json"
	"log"
	"pricegathering/websocket"
	"strconv"
	"strings"
	"time"
)

var OkxURL = "wss://wsaws.okx.com:8443/ws/v5/public"

// MarketInfo mapped from table <markets_marketinfo>
type MarketInfo struct {
	ID                                 int64   `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Exchange                           int32   `gorm:"column:exchange;not null" json:"exchange"`
	PricePrecision                     float64 `gorm:"column:price_precision;not null" json:"price_precision"`
	VolumePrecision                    float64 `gorm:"column:volume_precision;not null" json:"volume_precision"`
	TickerData                         bool    `gorm:"column:ticker_data;not null" json:"ticker_data"`
	BooktickerData                     bool    `gorm:"column:bookticker_data;not null" json:"bookticker_data"`
	OrderbookData                      bool    `gorm:"column:orderbook_data;not null" json:"orderbook_data"`
	BaseAsset                          string  `gorm:"column:base_asset;not null" json:"base_asset"`
	QuoteAsset                         string  `gorm:"column:quote_asset;not null" json:"quote_asset"`
	DefaultSettlementExchangeAccountID *int64  `gorm:"column:default_settlement_exchange_account_id" json:"default_settlement_exchange_account_id"`
}

type ChannelParams struct {
	Channel string `json:"channel"`
	InstId  string `json:"instId"`
}

type Market interface {
	GetMarketInfos() []*MarketInfo
	GetBookTickerChannel() chan BookTicker

	FetchData()
	ProcessData()
}
type OkxSpot struct {
	MarketInfos     []string
	BookTicker      chan BookTicker
	Ticker          chan Ticker
	WebSocketClient []*websocket.WebSocket
}

func StoreData(m Market) {

	// const numWorkers = 10

	// for i := 0; i < numWorkers; i++ {
	// 	go func(m Market) {
	// 		for bookTicker := range m.GetBookTickerChannel() {
	// 			caches.StoreHashSetInRedis(bookTicker.Stream_name+"_"+bookTicker.Symbol, strings.ToLower(string(bookTicker.Exchange)), bookTicker)
	// 		}
	// 	}(m)
	// }
}

func (m *OkxSpot) GetMarketInfos() []string {
	return m.MarketInfos
}

func (m *OkxSpot) GetBookTickerChannel() chan BookTicker {
	return m.BookTicker
}

func (m *OkxSpot) GetSubscribeMessage(marketInfos []string) interface{} {
	var channelParams []interface{}
	for _, market := range marketInfos {

		channelParams = append(channelParams, ChannelParams{
			Channel: "bbo-tbt",
			InstId:  market,
		})

		channelParams = append(channelParams, ChannelParams{
			Channel: "tickers",
			InstId:  market,
		})

	}
	return struct {
		Op   string        `json:"op"`
		Args []interface{} `json:"args"`
	}{
		Op:   "subscribe",
		Args: channelParams,
	}
}

func (m *OkxSpot) FetchData() {
	marketPerConnection := 2

	for i := 0; i < len(m.MarketInfos); i += marketPerConnection {
		end := i + marketPerConnection
		if end > len(m.MarketInfos) {
			end = len(m.MarketInfos)
		}
		channels := m.MarketInfos[i:end]
		log.Println(m.GetSubscribeMessage(channels))
		m.WebSocketClient = append(m.WebSocketClient, websocket.NewWebSocket(OkxURL, m.GetSubscribeMessage(channels)))
	}
}

func (m *OkxSpot) ProcessData() {
	for _, w := range m.WebSocketClient {
		var messageData struct {
			Arg  map[string]interface{}   `json:"arg"`
			Data []map[string]interface{} `json:"data"`
		}
		go func(w *websocket.WebSocket) {
			for message := range w.Data {
				err := json.Unmarshal(message, &messageData)
				if err != nil {
					log.Printf("Error unmarshaling JSON message: %v", err)
					continue
				}
				if messageData.Data == nil {
					continue
				}
				arg := messageData.Arg
				data := messageData.Data

				switch arg["channel"] {
				case "bbo-tbt":
					m.BookTicker <- m.getBookTicker(data[0])
				case "tickers":
					m.Ticker <- m.getTicker(data[0])
				}
			}
		}(w)
	}
}

func (m *OkxSpot) getTicker(data map[string]interface{}) Ticker {
	exchangeTime, _ := strconv.ParseInt(data["ts"].(string), 10, 64)
	lastPrice, _ := strconv.ParseFloat(data["last"].(string), 64)
	return Ticker{
		Exchange:             "okx_spot",
		Stream_name:          "ticker",
		Symbol:               strings.ToLower(data["instId"].(string)),
		Exchange_time:        exchangeTime,
		Update_time:          time.Now().UnixMilli(),
		Price_change:         0,
		Price_change_percent: 0,
		Last_price:           lastPrice,
	}

}

func (m *OkxSpot) getBookTicker(data map[string]interface{}) BookTicker {
	asks := data["asks"].([]interface{})
	ask := asks[0].([]interface{})
	bids := data["bids"].([]interface{})
	bid := bids[0].([]interface{})
	best_bid_price, _ := strconv.ParseFloat(bid[0].(string), 64)
	best_bid_qty, _ := strconv.ParseFloat(bid[1].(string), 64)
	best_ask_price, _ := strconv.ParseFloat(ask[0].(string), 64)
	best_ask_qty, _ := strconv.ParseFloat(ask[1].(string), 64)
	exchangeTime, _ := strconv.ParseInt(data["ts"].(string), 10, 64)
	return BookTicker{
		Exchange:       "okx_spot",
		Stream_name:    "bookTicker",
		Symbol:         strings.ToLower(data["instId"].(string)),
		Exchange_time:  exchangeTime,
		Update_time:    time.Now().UnixMilli(),
		Best_bid_price: best_bid_price,
		Best_bid_qty:   best_bid_qty,
		Best_ask_price: best_ask_price,
		Best_ask_qty:   best_ask_qty,
	}
}
