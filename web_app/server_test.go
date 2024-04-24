package web_app

// import (
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// )

// func TestRoutes(t *testing.T) {
// 	makeDefaultRouteMap := func(t *testing.T, routePath string, routeBody string) RouteMap {

// 		t.Helper()

// 		routeMap := make(RouteMap)
// 		if routePath == "" {
// 			return routeMap
// 		}

// 		if routeBody == "" {
// 			return routeMap
// 		}
		
// 		routeTuple := RouteTuple{"GET", func(w http.ResponseWriter, r *http.Request) {
// 			w.Write([]byte(routeBody))
// 		}}
// 		routeMap[RouteURL(routePath)] = []RouteTuple{routeTuple}
// 		return routeMap
// 	}
	
// 	t.Run("returns the page at /", func(t *testing.T) {
// 		homeRouteBody := "Hello, World!"
// 		routes := makeDefaultRouteMap(t, "/", homeRouteBody)
// 		request, _ := http.NewRequest(http.MethodGet, "/", nil)
// 		response := httptest.NewRecorder()

// 		handler, _ := StartConfigurableWebAppServer(routes)

// 		handler.ServeHTTP(response, request)

// 		got := response.Body.String()
// 		want := homeRouteBody

// 		if got != want {
// 			t.Errorf("want %s, but got %s", want, got)
// 		}

// 	})

// 	t.Run("returns a 404 for pages that don't exist", func(t *testing.T) {
// 		homeRouteBody := "Hello, World!"
// 		routes := makeDefaultRouteMap(t, "/", homeRouteBody)
// 		request, _ := http.NewRequest(http.MethodGet, "/does/not/exist", nil)
// 		response := httptest.NewRecorder()

// 		handler, _ := StartConfigurableWebAppServer(routes)
// 		handler.ServeHTTP(response, request)

// 		got := response.Result().StatusCode
// 		want := 404

// 		if got != want {
// 			t.Errorf("doesn't return 404 on non-existent route, returns %d", got)
// 		}
// 	})

// 	t.Run("returns an error on empty RouteMaps", func(t *testing.T) {
// 		routeMap := makeDefaultRouteMap(t, "", "")
// 		_, err := StartConfigurableWebAppServer(routeMap)
// 		if err != ErrorEmptyRouteMap {
// 			t.Fatal("expected an error to be generated while passing an empty routeMap")
// 		}
// 	})
// }
