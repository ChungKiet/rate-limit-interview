package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"kiet/rate-limit/adapter"
	"net/http"
	"testing"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/api", rateLimitMiddleware(), apiHandler)
	return r
}

func TestRateLimitMiddleware(t *testing.T) {
	r := setupRouter()

	w := &adapter.ResponseRecorder{}
	req, _ := http.NewRequest("GET", "/api?user=user_a", nil)

	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		t.Fatal(err)
	}
	defer redisConn.Close()

	// clean up data
	redisConn.Do("DEL", "ratelimit:user_a")

	for i := 0; i < 5; i++ {
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Status)
	}

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Status)
}

func TestAPIHandler(t *testing.T) {
	r := setupRouter()

	w := &adapter.ResponseRecorder{}
	req, _ := http.NewRequest("GET", "/api?user=user_a", nil)

	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		t.Fatal(err)
	}
	defer redisConn.Close()

	// clean up data
	redisConn.Do("DEL", "ratelimit:user_a")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Status)
	assert.Equal(t, []byte(`{"shopping_list":["cheese","milk"]}`), w.Body)
}
