package ginutil

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.GET("/test", Wrap(func(ctx RequestContext[any]) (string, error) {
		return "Hello", nil
	}))
	r.GET("/issue", IssueCredential("TestKey", 10*time.Second, func(code string) (Credential, error) {
		return Credential{
			ID: code, Subject: "TestKey", Audience: []string{"Test-0", "Test-1"},
		}, nil
	}))
	return r
}

func TestWrap(t *testing.T) {
	router := setupRouter()
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(recorder, request)
	t.Log(recorder.Body.String())
}

func TestIssueCredential(t *testing.T) {
	router := setupRouter()
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("GET", "/issue?code=8282", nil)
	if err != nil {
		t.Error(err)
	}
	router.ServeHTTP(recorder, request)
	t.Log(recorder.Body.String())
}
