package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

type RateLimit struct {
	Limit    int
	Duration time.Duration
}

var rateLimits = map[string]RateLimit{}

var db *sql.DB
var mu sync.Mutex

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
	}

	user := os.Getenv("DB_USER")
	dbname := os.Getenv("DB_NAME")
	password := os.Getenv("DB_PASSWORD")
	sslmode := os.Getenv("DB_SSLMODE")

	connStr := fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s", user, dbname, sslmode, password)
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		panic(err)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatalf("CONFIG_PATH environment variable is required")
	}

	viper.SetConfigFile(configPath)
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file: %v", err)
	}

	var config struct {
		RateLimits map[string]RateLimit `mapstructure:"rate_limits"`
	}

	if err := viper.Unmarshal(&config); err != nil {
		log.Fatalf("Error unmarshalling config: %v", err)
	}

	for key, value := range config.RateLimits {
		value.Duration *= time.Second
		config.RateLimits[key] = value
	}

	rateLimits = config.RateLimits

	// migration
	createTable()
}

func createTable() {
	query := `
    CREATE TABLE IF NOT EXISTS api_calls (
        id SERIAL PRIMARY KEY,
        user_id TEXT,
        timestamp TIMESTAMP
    );`
	_, err := db.Exec(query)
	if err != nil {
		panic(err)
	}
}

func main() {
	// init gin server
	r := gin.Default()
	r.GET("/api", rateLimitMiddleware(), apiHandler)
	r.Run(":8080")
}

func rateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.Query("user")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "USER_ID_REQUIRED"})
			return
		}

		rateLimit, exists := rateLimits[userID]
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "NOT_CONFIG_RATE_LIMIT_FOR_THIS_USER"})
			return
		}

		mu.Lock()
		defer mu.Unlock()

		now := time.Now()
		cutoff := now.Add(-rateLimit.Duration)

		tx, err := db.Begin()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FAILED_TO_BEGIN_TRANSACTION"})
			return
		}

		query := `DELETE FROM api_calls WHERE timestamp < $1`
		_, err = tx.Exec(query, cutoff)
		if err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FAILED_TO_CLEAN_UP_DATA"})
			return
		}

		query = `SELECT COUNT(*) FROM api_calls WHERE user_id = $1`
		var count int
		err = tx.QueryRow(query, userID).Scan(&count)
		if err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FAILED_TO_COUNT_API_CALL"})
			return
		}

		if count >= rateLimit.Limit {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "RATE_LIMIT_EXCEEDED"})
			return
		}

		query = `INSERT INTO api_calls (user_id, timestamp) VALUES ($1, $2)`
		_, err = tx.Exec(query, userID, now)
		if err != nil {
			tx.Rollback()
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FAILED_TO_RECORD_API_CALL"})
			return
		}

		err = tx.Commit()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "FAILED_TO_COMMIT_TRANSACTION"})
			return
		}
	}
}

func apiHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"shopping_list": []string{"cheese", "milk"}})
}
