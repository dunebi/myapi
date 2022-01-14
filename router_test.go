package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRouter(t *testing.T) {
	router := SetupRouter()

	w := httptest.NewRecorder()
	request, _ := http.NewRequest("GET", "/hi", nil)
	router.ServeHTTP(w, request)

	assert.Equal(t, http.StatusOK, w.Code)
}
