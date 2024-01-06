package web_app

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	t.Run("returns the page at /", func(t *testing.T) {
		routes := make(map[string][]byte)
		homeRouteBody := "Hello, World!"
		routes["/"] = []byte(homeRouteBody)
		request, _ := http.NewRequest(http.MethodGet, "/", nil)
		response := httptest.NewRecorder()

		handler := StartConfigurableWebAppServer(routes)
		
		handler(response, request)

		got := response.Body.String()
		want := homeRouteBody

		if got != want {
			t.Errorf("want %s, but got %s", want, got)
		}
		
	})
}
