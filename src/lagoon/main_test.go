package main

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	router   *gin.Engine
	recorder *httptest.ResponseRecorder
)

func TestMain(m *testing.M) {
	router = setupRouter()
	recorder = httptest.NewRecorder()

	os.Exit(m.Run())
}

func TestRedirectionAtRoot(t *testing.T) {
	defer recorder.Flush()

	req, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 301, recorder.Code)
	assert.Equal(t, "/lagoon/ui", recorder.Header().Get("Location"))
}

func TestRedirectionAtContextRoot(t *testing.T) {
	defer recorder.Flush()

	req, _ := http.NewRequest("GET", contextPath, nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 301, recorder.Code)
	assert.Equal(t, "/lagoon/", recorder.Header().Get("Location"))
}

func TestRedirectionAtContextRootWithSlash(t *testing.T) {
	defer recorder.Flush()

	req, _ := http.NewRequest("GET", contextPath+"/", nil)
	router.ServeHTTP(recorder, req)

	assert.Equal(t, 301, recorder.Code)
	assert.Equal(t, "/lagoon/ui", recorder.Header().Get("Location"))
}
