package main

import (
	"fmt"
	"os"
	"pricegathering/prices"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Price struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement:true" json:"id"`
	Price     float64   `gorm:"column:price;not null" json:"price"`
	Market    string    `gorm:"index:,column:market;" json:"market"`
	CreatedAt time.Time `gorm:"column:created_at;" json:"created_at"`
}

func GetDBConnection() (*gorm.DB, error) {
	// Load the environment variables from the specified .env file
	envFilePath := "./.env"
	err := godotenv.Load(envFilePath)
	if err != nil {
		fmt.Println("Error loading .env file:", err)
	}

	host := os.Getenv("DB_HOST")
	if host == "" {
		host = "postgres"
	}

	port, err := strconv.ParseInt(os.Getenv("DB_PORT"), 10, 64)
	if err != nil {
		port = 5432
	}
	user := os.Getenv("DB_USER")
	if user == "" {
		user = "interview"
	}

	password := os.Getenv("DB_PASSWORD")
	if password == "" {
		password = "interview"
	}

	dbname := "interview"

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d", host, user, password, dbname, port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	fmt.Println("Connected to PostgreSQL")
	return db, nil
}
func main() {
	db, _ := GetDBConnection()

	market := prices.OkxSpot{
		MarketInfos: []string{"DAO-USDT", "SSV-USDT", "BTC-USDT", "ETH-USDT"},
		BookTicker:  make(chan prices.BookTicker),
		Ticker:      make(chan prices.Ticker),
	}

	sampling := map[string]time.Time{
		strings.ToLower("DAO-USDT"): time.Now(),
		strings.ToLower("SSV-USDT"): time.Now(),
		strings.ToLower("BTC-USDT"): time.Now(),
		strings.ToLower("ETH-USDT"): time.Now(),
	}

	market.FetchData()
	go market.ProcessData()
	for {
		select {
		case bt := <-market.BookTicker:
			if time.Now().After(sampling[bt.Symbol].Add(time.Second)) {
				p := Price{
					Market:    bt.Symbol,
					Price:     (bt.Best_ask_price + bt.Best_bid_price) / 2,
					CreatedAt: time.Now(),
				}
				sampling[bt.Symbol] = time.Now()
				db.Create(&p)
			}
		case t := <-market.Ticker:
			if time.Now().After(sampling[t.Symbol].Add(time.Second)) {
				p := Price{
					Market:    t.Symbol,
					Price:     t.Last_price,
					CreatedAt: time.Now(),
				}
				sampling[t.Symbol] = time.Now()
				db.Create(&p)
			}
		}
	}
}
