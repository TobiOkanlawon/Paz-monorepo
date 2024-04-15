package web_app

import (
	"encoding/gob"
	"errors"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/sessions"
)

type RouteURL string

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

func StartConfigurableWebAppServer(routes RouteMap, secretKey []byte) (handler http.Handler, cleanup func() error, err error) {

	// Removed file logging because it wasn't working properly on the server and the journalctl logging seemed to be doing much better

	// f, err := os.OpenFile("log.txt", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	// if err != nil {
	// 	log.Fatalf("error opening log file: %v", err)
	// }

	// log.SetOutput(f)
	// log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.LstdFlags)

	if len(routes) == 0 {
		return nil, func() error { return nil }, ErrorEmptyRouteMap
	}

	r := NewRouter(routes)
	// TODO: test and DI the fileServer
	fs := http.FileServer(http.Dir("./web_app/templates/static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	// TODO: consider having all templates parsed before server initialization
	// tpl = template.Must(template.ParseGlob("templates/*.gohtml"))

	return r, func() error { return nil }, nil
}

// TODO: write tests for the handlers getting passed in
func WebAppServer(secretKey []byte, paystackPublicKey, paystackSecretKey string) (handler http.Handler, cleanUp func() error, err error) {
	partialsManager := GetPartialsManager(os.DirFS("./partials"))
	db := DB{}
	db.Connect()

	cookieStore := sessions.NewCookieStore(secretKey)
	gob.Register(&UserCookie{})
	// store.Options.HttpOnly = true
	// TODO: set a config argument that has amongst its fields (DEBUG), for local development mode
	// set this field based on if it is dev or prod
	// store.Options.Secure = true

	handlerManager := NewHandlerManager(partialsManager, &db, cookieStore, paystackPublicKey, paystackSecretKey)
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
	routeMap[RouteURL("/dashboard/loans/get-loan")] = []RouteTuple{
		{"GET", handlerManager.getLoansGetHandler},
		{"POST", handlerManager.getLoansPostHandler},
	}
	routeMap[RouteURL("/dashboard/investments")] = []RouteTuple{
		{"GET", handlerManager.investmentsGetHandler},
	}
	routeMap[RouteURL("/dashboard/investments/form")] = []RouteTuple{
		{"GET", handlerManager.investmentsFormGetHandler},
		{"POST", handlerManager.investmentsFormPostHandler},
	}
	routeMap[RouteURL("/dashboard/fragments/bvn")] = []RouteTuple{
		{"GET", handlerManager.bvnModalGetHandler},
		{"POST", handlerManager.addBVNPostHandler},
	}
	routeMap[RouteURL("/dashboard/savings/family-vault")] = []RouteTuple{
		{"GET", handlerManager.familyVaultGetHandler},
		{"POST", handlerManager.familyVaultPostHandler},
	}
	routeMap[RouteURL("/dashboard/savings/family-vault/{planID}")] = []RouteTuple{
		{"GET", handlerManager.familyVaultPlanGetHandler},
	}
	routeMap[RouteURL("/dashboard/savings/target-savings")] = []RouteTuple{
		{"GET", handlerManager.targetSavingsGetHandler},
	}
	routeMap[RouteURL("/dashboard/savings/solo-saver")] = []RouteTuple{
		{"GET", handlerManager.soloSavingsGetHandler},
		{"POST", handlerManager.soloSavingsAddFunds},
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
	routeMap[RouteURL("/utility/paystack-verification-webhook")] = []RouteTuple{
		{"POST", handlerManager.paystackVerificationWebhook},
	}
	routeMap[RouteURL("/dashboard/logout")] = []RouteTuple{
		{"GET", handlerManager.logoutGetHandler},
	}

	handler, cleanUp, err = StartConfigurableWebAppServer(routeMap, secretKey)

	if err != nil {
		return nil, func() error { return nil }, err
	}

	cleanUpFunction := func() error {
		firstError := cleanUp()
		secondError := db.Conn.Close()
		if firstError != nil {
			return firstError
		}
		if secondError != nil {
			return secondError
		}
		return nil
	}

	return handler, cleanUpFunction, nil
}
