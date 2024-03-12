package web_app

import (
	"io/fs"
	// "net/http"
	// "net/http/httptest"
	// "net/url"
	// "strings"

	// // "net/http"
	// // "net/http/httptest"
	"testing"
	// "testing/fstest"
)

func TestHandlers(t *testing.T) {
	// t.Run("test login POST handler", func(t *testing.T) {
	// 	fsMap := fstest.MapFS{}
	// 	manager := NewHandlerManager(GetPartialsManager(fsMap))

	// 	form := url.Values{}
	// 	form.Add("email-address", "tobiinlondon34@gmail.com")
	// 	form.Add("password", "password")

	// 	req, err := http.NewRequest("POST", "/login", strings.NewReader(
	// 		form.Encode(),
	// 	))

	// 	if err != nil {
	// 		t.Fatalf("err from request %q", err)
	// 	}
		
	// 	response := httptest.NewRecorder()

	// 	req.PostForm = form
	// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// 	manager.loginPostHandler(response, req)

	// 	if response.Result().StatusCode != http.StatusOK {
	// 		t.Fatalf("returned something other than 200: %q", response.Result().StatusCode)
	// 	}
		
	// })

	
// 	t.Run("returns a 200 status code ", func(t *testing.T) {
// 		fsMap := fstest.MapFS{}
// 		manager := NewHandlerManager(GetPartialsManager(fsMap))
		
// 		tt := []struct{
// 			f func(http.ResponseWriter, *http.Request)
// 			url string
// 		}{
// 			{manager.homeGetHandler, "/home"},
// 			{manager.loginGetHandler, "/login"},
// 		}

// 		for _, value := range tt {
// 			request, _ := http.NewRequest(http.MethodGet, value.url, nil)
// 			response := httptest.NewRecorder()
// 			value.f(response, request)

// 			got := response.Code
// 			want := 200

// 			if got != want {
// 				t.Errorf("wanted status code 200 in url %q but got %d", value.url, got)
// 			}
// 		}
// 	})
}

type StubPartialsManager struct {
	fileSystem fs.FS
	store *store
}

func (s *StubPartialsManager) GetPartial(name string) (string, error) {
	return "", nil
}

func (s *StubPartialsManager) RegisterPartial(name, templateName string) (error) {
	return nil
}

func (s *StubPartialsManager) Clear() error {
	return nil
}

func TestHandlerManger(t *testing.T) {
	t.Run("successfully creates a handlerManager", func(t *testing.T) {
		testPartialsManager := &StubPartialsManager{}
		testStore, err := NewSqlStore("/tmp/test.db")

		if err != nil {
			t.Fatalf("failed to create store with err %q", err)
		}
		_ =  NewHandlerManager(testPartialsManager, testStore)
	})
	
}
