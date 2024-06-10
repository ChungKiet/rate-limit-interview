package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"
	"kiet/rate-limit/adapter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitE2E(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	redisConn, err := redis.Dial("tcp", ":6379")
	if err != nil {
		t.Fatal(err)
	}
	defer redisConn.Close()

	redisConn.Do("DEL", "ratelimit:user_a")

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/api?user=user_a", nil)
		w := &adapter.ResponseRecorder{}
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Status)
	}

	req, _ := http.NewRequest("GET", "/api?user=user_a", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}
