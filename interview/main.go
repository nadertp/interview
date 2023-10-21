package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
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

var randGen = rand.NewSource(time.Now().UnixNano())
var random = rand.New(randGen)

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

func GeneratePrice(db *gorm.DB, market string, price float64, delay time.Duration) {
	realPrice := price

	for {
		nowPrice := realPrice
		margin := 0.0
		if random.Float64() < 0.05 {
			margin = 50 + random.Float64()
			margin = margin * math.Copysign(1, rand.Float64()-0.5)
			nowPrice = realPrice + realPrice*margin/100
			fmt.Printf("anomaly in %v => %v\n", market, margin)
		} else {
			margin = (random.Float64() * 0.5) * math.Copysign(1, rand.Float64()-0.5)
			nowPrice = realPrice + realPrice*margin/100
			fmt.Println("price generated")
		}
		// fmt.Println(margin)

		p := Price{
			Market:    market,
			Price:     nowPrice,
			CreatedAt: time.Now(),
		}
		db.Create(&p)
		time.Sleep(delay)
	}
}

func main() {
	db, _ := GetDBConnection()
	db.AutoMigrate(&Price{})
	go GeneratePrice(db, "btc-usdt", 27000, time.Second)
	go GeneratePrice(db, "eth-usdt", 1600, time.Second)
	go GeneratePrice(db, "ssv-usdt", 12, time.Second)
	go GeneratePrice(db, "dao-usdt", 0.5, time.Second)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, os.Interrupt)
	s := <-c
	fmt.Println("Got signal:", s)

}
