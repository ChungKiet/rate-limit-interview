package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"os"
	"time"
)

// read from config
var rateLimits = map[string]RateLimit{}

type RateLimit struct {
	Limit    int
	Duration time.Duration
}

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file")
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
}

func main() {
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

		redisConn, err := redis.Dial("tcp", ":6379")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "REDIS_CONNECTION_FAILED"})
			return
		}
		defer redisConn.Close()

		key := "ratelimit:" + userID
		count, err := redis.Int(redisConn.Do("GET", key))
		if err != nil && err != redis.ErrNil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "GET_RATE_LIMIT_COUNTING_FAILED"})
			return
		}

		if count >= rateLimit.Limit {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "RATE_LIMIT_EXCEEDED"})
			return
		}

		_, err = redisConn.Do("INCR", key)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "INCREASE_RATE_LIMIT_FAILED"})
			return
		}

		if count == 0 {
			_, err = redisConn.Do("EXPIRE", key, int(rateLimit.Duration.Seconds()))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "SET_RATE_LIMIT_EXPIRATION_FAILED"})
				return
			}
		}

		//c.Next()
	}
}

func apiHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"shopping_list": []string{"cheese", "milk"}})
}
