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

var ErrorEmptyRouteMap error = errors.New("passed empty RouteMap as arg")

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
	r := chi.NewRouter()
	
	// TODO: the logger should also be injected
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	// TODO: handle post-slashes
	r.Get("/", handlerManager.indexGetHandler)
	r.Get("/dashboard/home", handlerManager.dashboardHomeGetHandler)
	r.Get("/login", handlerManager.loginGetHandler)
	r.Post("/login", handlerManager.loginPostHandler)
	r.Get("/register", handlerManager.registerGetHandler)
	r.Post("/register", handlerManager.registerPostHandler)
	r.Get("/verify", handlerManager.verifyEmailGetHandler)
	r.Get("/forgot-password", handlerManager.forgotPasswordGetHandler)
	r.Get("/dashboard/profile", handlerManager.profileGetHandler)
	r.Get("/dashboard/savings", handlerManager.savingsGetHandler)
	r.Get("/dashboard/loans", handlerManager.loansGetHandler)
	r.Get("/dashboard/loans/get-loan", handlerManager.getLoansGetHandler)
	r.Post("/dashboard/loans/get-loan", handlerManager.getLoansPostHandler)
	r.Get("/dashboard/investments", handlerManager.investmentsGetHandler)
	r.Get("/dashboard/investments/form", handlerManager.investmentsFormGetHandler)
	r.Post("/dashboard/investments/form", handlerManager.investmentsFormPostHandler)
	r.Get("/dashboard/fragments/bvn", handlerManager.bvnModalGetHandler)
	r.Post("/dashboard/fragments/bvn", handlerManager.addBVNPostHandler)
	r.Get("/dashboard/savings/family-vault", handlerManager.familyVaultGetHandler)
	r.Post("/dashboard/savings/family-vault", handlerManager.familyVaultPostHandler)
	r.Get("/dashboard/savings/family-vault/{planID}", handlerManager.familyVaultGetHandler)
	r.Get("/dashboard/savings/target-savings", handlerManager.targetSavingsGetHandler)
	r.Get("/dashboard/savings/solo-saver", handlerManager.soloSavingsGetHandler)
	r.Post("/dashboard/savings/solo-saver", handlerManager.soloSavingsAddFunds)
	r.Get("/dashboard/thrift", handlerManager.thriftGetHandler)
	r.Get("/dashboard/thrift/new", handlerManager.thriftNewGetHandler)
	r.Get("/dashboard/thrift/{thriftID}", handlerManager.thriftPlanGetHandler)
	r.Post("/utility/paystack-verification-webhook", handlerManager.paystackVerificationWebhook)
	r.Get("/dashboard/logout", handlerManager.loginGetHandler)
	r.Get("/admin", handlerManager.adminHomeGetHandler)

	fs := http.FileServer(http.Dir("./web_app/templates/static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

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

	return r, cleanUpFunction, nil
}
