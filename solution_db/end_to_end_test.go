package main

import (
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	rate_limit_interview "kiet/rate-limit/adapter"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRateLimitE2ESQL(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := setupRouter()

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	// clean up data
	tx.Exec("DELETE FROM api_calls WHERE user_id = $1", "user_a")

	for i := 0; i < 5; i++ {
		req, _ := http.NewRequest("GET", "/api?user=user_a", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	}

	req, _ := http.NewRequest("GET", "/api?user=user_a", nil)
	w := &rate_limit_interview.ResponseRecorder{}
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Status)
}
