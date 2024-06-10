package main

import (
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	rate_limit_interview "kiet/rate-limit/adapter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/api", rateLimitMiddleware(), apiHandler)
	return r
}

func TestRateLimitMiddlewareSQL(t *testing.T) {
	r := setupRouter()

	w := &rate_limit_interview.ResponseRecorder{}
	req, _ := http.NewRequest("GET", "/api?user=user_a", nil)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	// Ensure the table is clean before testing
	tx.Exec("DELETE FROM api_calls WHERE user_id = $1", "user_a")

	for i := 0; i < 5; i++ {
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Status)
	}

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Status)
}

func TestAPIHandlerSQL(t *testing.T) {
	r := setupRouter()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api?user=user_a", nil)

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	// clean up data
	tx.Exec("DELETE FROM api_calls WHERE user_id = $1", "user_a")

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"shopping_list":["cheese","milk"]}`, w.Body.String())
}
