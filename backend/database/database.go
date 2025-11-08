package database

import (
	"context"
	"final_project/models"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
var Rdb *redis.Client

func DBLoad() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}
	fmt.Println("Database connection established")
	DB = db
}

func Redis() {
	ctx := context.Background()
	dbNum, err := strconv.Atoi(os.Getenv("REDIS_DB"))
	if err != nil {
		dbNum = 0
	}

	Rdb = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       dbNum,
	})

	pong, err := Rdb.Ping(ctx).Result()
	if err != nil {
		panic(err)
	}
	fmt.Println("Redis connected:", pong)
}

func DBMigrate() {
	if err := DB.Debug().AutoMigrate(&models.Bookmark{}, models.Chapter{}, models.Genre{}, models.Novel{}, models.User{}, models.Tag{}, models.RefreshToken{}); err != nil {
		panic("Failed to migrate database")
	}
	fmt.Println("Database migration completed")
}
