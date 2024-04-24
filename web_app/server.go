package web_app

import (
	"encoding/gob"
	"errors"
	"log"
	"net/http"
	"os"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/gorilla/csrf"
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

	csrfMiddleware := csrf.Protect(
		[]byte(secretKey),
		// TODO: Add secure to the list, base if off a debug environment variable
		// csrf.Secure(),
	)
	
	// TODO: the logger should also be injected
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)

	dashboardSubRouter := chi.NewRouter()
	dashboardSubRouter.Use(csrfMiddleware)
	r.Mount("/dashboard", dashboardSubRouter)
	
	adminSubRouter := chi.NewRouter()
	adminSubRouter.Use(csrfMiddleware)
	r.Mount("/admin", adminSubRouter)
	
	preAuthSubRouter := chi.NewRouter()
	preAuthSubRouter.Use(csrfMiddleware)
	r.Mount("/", preAuthSubRouter)

	apiSubRouter := chi.NewRouter()
	r.Mount("/utility", apiSubRouter)

	// TODO: handle post-slashes
	preAuthSubRouter.Get("/", handlerManager.indexGetHandler)
	preAuthSubRouter.Get("/login", handlerManager.loginGetHandler)
	preAuthSubRouter.Post("/login", handlerManager.loginPostHandler)
	preAuthSubRouter.Get("/register", handlerManager.registerGetHandler)
	preAuthSubRouter.Post("/register", handlerManager.registerPostHandler)
	preAuthSubRouter.Get("/forgot-password", handlerManager.forgotPasswordGetHandler)
	preAuthSubRouter.Get("/verify", handlerManager.verifyEmailGetHandler)

	// TODO: actually, change the dashboard "/" route to redirect to the /home, or the other way
	// But there really should only be one way to do these things
	dashboardSubRouter.Get("/", handlerManager.dashboardHomeGetHandler)
	dashboardSubRouter.Get("/home", handlerManager.dashboardHomeGetHandler)
	dashboardSubRouter.Get("/profile", handlerManager.profileGetHandler)
	dashboardSubRouter.Get("/savings", handlerManager.savingsGetHandler)
	dashboardSubRouter.Get("/loans", handlerManager.loansGetHandler)
	dashboardSubRouter.Get("/loans/get-loan", handlerManager.getLoansGetHandler)
	dashboardSubRouter.Post("/loans/get-loan", handlerManager.getLoansPostHandler)
	dashboardSubRouter.Get("/investments", handlerManager.investmentsGetHandler)
	dashboardSubRouter.Get("/investments/form", handlerManager.investmentsFormGetHandler)
	dashboardSubRouter.Post("/investments/form", handlerManager.investmentsFormPostHandler)
	dashboardSubRouter.Get("/fragments/bvn", handlerManager.bvnModalGetHandler)
	dashboardSubRouter.Post("/fragments/bvn", handlerManager.addBVNPostHandler)
	dashboardSubRouter.Get("/savings/family-vault", handlerManager.familyVaultGetHandler)
	dashboardSubRouter.Post("/savings/family-vault", handlerManager.familyVaultPostHandler)
	dashboardSubRouter.Get("/savings/family-vault/{planID}", handlerManager.familyVaultGetHandler)
	dashboardSubRouter.Get("/savings/target-savings", handlerManager.targetSavingsGetHandler)
	dashboardSubRouter.Get("/savings/solo-saver", handlerManager.soloSavingsGetHandler)
	dashboardSubRouter.Post("/savings/solo-saver", handlerManager.soloSavingsAddFunds)
	dashboardSubRouter.Get("/thrift", handlerManager.thriftGetHandler)
	dashboardSubRouter.Get("/thrift/new", handlerManager.thriftNewGetHandler)
	dashboardSubRouter.Get("/thrift/{thriftID}", handlerManager.thriftPlanGetHandler)	
	dashboardSubRouter.Get("/logout", handlerManager.loginGetHandler)
	
	apiSubRouter.Post("/paystack-verification-webhook", handlerManager.paystackVerificationWebhook)
	
	adminSubRouter.Get("/", handlerManager.adminHomeGetHandler)

	fs := http.FileServer(http.Dir("./web_app/templates/static/"))
	r.Handle("/static/*", http.StripPrefix("/static/", fs))

	cleanUpFunction := func() error {
		firstError := cleanUp()
		secondError := db.Conn.Close()
		if firstError != nil {
			log.Printf("error with cleanup %s \n", firstError)
		}
		if secondError != nil {
			log.Printf("error with cleanup %s \n", secondError)
			os.Exit(1)
		}
		return nil
	}

	return r, cleanUpFunction, nil
}
