package web_app

import (
	"encoding/gob"
	"errors"
	"net/http"
	"os"

	"log"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
)

type RouteURL string

type UserCookie struct {
	ID           uint
	FirstName    string
	LastName     string
}

var ErrorEmptyRouteMap error = errors.New("passed empty RouteMap as arg")

type RouteTuple struct {
	// TODO: Make this into an enum to mamke it safer
	Method  string
	Handler func(w http.ResponseWriter, r *http.Request)
}

type RouteMap map[RouteURL][]RouteTuple

func NewRouter(routes RouteMap) *chi.Mux {
	r := chi.NewRouter()
	// TODO: the logger should also be injected
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	for url, route_tuple := range routes {
		for _, tuple := range route_tuple {
			r.Method(tuple.Method, string(url), http.HandlerFunc(tuple.Handler))
		}
	}

	return r
}

func StartConfigurableWebAppServer(routes RouteMap, secretKey []byte) (http.Handler, error) {
	if len(routes) == 0 {
		return nil, ErrorEmptyRouteMap
	}

	r := NewRouter(routes)
	// TODO: test and DI the fileServer
	fs := http.FileServer(http.Dir("./web_app/templates/static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// TODO: consider having all templates parsed before server initialization
	// tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

	return r, nil
}

// TODO: write tests for the handlers getting passed in
func WebAppServer(secretKey []byte) (http.Handler, error) {
	partialsManager := GetPartialsManager(os.DirFS("./partials"))
	store, err := NewSqlStore("test.db")
	if err != nil {
		log.Fatalf("failed to initialize store %s", err)
	}

	cookieStore := sessions.NewCookieStore(secretKey)
	gob.Register(&UserCookie{})
	// store.Options.HttpOnly = true
	// TODO: set a config argument that has amongst its fields (DEBUG), for local development mode
	// set this field based on if it is dev or prod
	// store.Options.Secure = true

	handlerManager := NewHandlerManager(partialsManager, store, cookieStore)
	routeMap := make(RouteMap)
	// TODO: handle post-slashes
	routeMap[RouteURL("/")] = []RouteTuple{
		{"GET", handlerManager.indexGetHandler},
	}
	routeMap[RouteURL("/dashboard/home")] = []RouteTuple{
		{"GET", handlerManager.dashboardHomeGetHandler},
	}
	routeMap[RouteURL("/login")] = []RouteTuple{
		{"GET", handlerManager.loginGetHandler},
		{"POST", handlerManager.loginPostHandler},
	}
	routeMap[RouteURL("/register")] = []RouteTuple{
		{"GET", handlerManager.registerGetHandler},
		{"POST", handlerManager.registerPostHandler},
	}
	routeMap[RouteURL("/verify")] = []RouteTuple{
		{"GET", handlerManager.verifyEmailGetHandler},
	}
	routeMap[RouteURL("/forgot-password")] = []RouteTuple{
		{"GET", handlerManager.forgotPasswordGetHandler},
	}
	routeMap[RouteURL("/dashboard/profile")] = []RouteTuple{
		{"GET", handlerManager.profileGetHandler},
	}
	routeMap[RouteURL("/dashboard/savings")] = []RouteTuple{
		{"GET", handlerManager.savingsGetHandler},
	}
	routeMap[RouteURL("/dashboard/loans")] = []RouteTuple{
		{"GET", handlerManager.loansGetHandler},
	}
	routeMap[RouteURL("/dashboard/investments")] = []RouteTuple{
		{"GET", handlerManager.investmentsGetHandler},
	}
	routeMap[RouteURL("/dashboard/fragments/bvn")] = []RouteTuple{
		{"GET", handlerManager.bvnModalGetHandler},
		{"POST", handlerManager.addBVNPostHandler},
	}
	routeMap[RouteURL("/dashboard/savings/family-vault")] = []RouteTuple{
		{"GET", handlerManager.familyVaultGetHandler},
	}
	routeMap[RouteURL("/dashboard/savings/target-saver")] = []RouteTuple{
		{"GET", handlerManager.targetSavingsGetHandler},
	}
	routeMap[RouteURL("/dashboard/thrift")] = []RouteTuple{
		{"GET", handlerManager.thriftGetHandler},
	}
	routeMap[RouteURL("/dashboard/thrift/new")] = []RouteTuple{
		{"GET", handlerManager.thriftNewGetHandler},
	}
	routeMap[RouteURL("/dashboard/thrift/{thriftID}")] = []RouteTuple{
		{"GET", handlerManager.thriftPlanGetHandler},
	}

	handler, err := StartConfigurableWebAppServer(routeMap, secretKey)

	if err != nil {
		return nil, err
	}

	return handler, nil
}
