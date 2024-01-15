package web_app

import (
	"errors"
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type RouteURL string
var ErrorEmptyRouteMap error = errors.New("passed empty RouteMap as arg")

type RouteTuple struct {
	// TODO: Make this into an enum to mamke it safer
	Method  string
	Handler func(w http.ResponseWriter, r *http.Request)
}

type RouteMap map[RouteURL][]RouteTuple

func generateRouter(routes RouteMap) *chi.Mux {
	r := chi.NewRouter()
	// TODO: the logger should also be injected
	r.Use(middleware.Logger)

	for url, route_tuple := range routes {
		for _, tuple  := range route_tuple {
			r.Method(tuple.Method, string(url), http.HandlerFunc(tuple.Handler))
		}
	}

	return r
}

func StartConfigurableWebAppServer(routes RouteMap) (http.Handler, error) {
	// r := func (w http.ResponseWriter, r *http.Request) {
	// w.Write([]byte("Hello, World!"))
	// }

	if len(routes) == 0 {
		return nil, ErrorEmptyRouteMap
	}
	
	r := generateRouter(routes)

	return r, nil
}

func WebAppServer() (http.Handler, error){
	handler, err := StartConfigurableWebAppServer(make(RouteMap))

	if err != nil {
		return nil, err
	}

	return handler, nil
}
