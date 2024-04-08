package web_app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

func NewHandlerManager(partialsManager IPartialsManager, store IStore, cookieStore *sessions.CookieStore, paystackPublicKey, paystackSecretKey string) *HandlerManager {
	return &HandlerManager{partialsManager, store, cookieStore, paystackPublicKey, paystackSecretKey}
}

func (h *HandlerManager) indexGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: this is the home page
	// TODO: check auth status, then redirect accordingly
	// i.e if authed, redirect to dashboard home
	w.WriteHeader(200)
}

func (h *HandlerManager) loginGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: check if the user is logged in
	// route them to the dashboard if they're logged in
	// TODO: create middleware for automatic redirection based on auth status

	// TODO: Check that each appropriate route returns a content-type
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/login.html", "./web_app/templates/layouts/pre_auth-base.html")
	// TODO: implement a 500 page
	// TODO: handle static files with chi router
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, "/login")
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, "/login")
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	loginInformation, err := h.store.AuthenticateUser(email, password)

	// TODO: handle validation and CSRF

	if err != nil {
		var errorsMap = make(map[string]string)
		w.WriteHeader(http.StatusUnauthorized)

		// this is the default message
		errorsMap["Email"] = "An error occured"

		if err == ErrAccountDoesNotExist {
			errorsMap["Email"] = "Account does not exist"
		}

		if err == ErrUserNotVerified {
			errorsMap["Email"] = "Verify your email before trying to log in"
		}

		if err == ErrPasswordIncorrect {
			errorsMap["Email"] = "Email or password is incorrect"
		}

		tmpl, err := template.ParseFiles("./web_app/templates/login.html", "./web_app/templates/layouts/pre_auth-base.html")

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	// Handling the session authentication

	// TODO: for the session authentication. Implement a
	// middleware function that checks that the cookie is
	// proper. If it isn't, delete it and redirect the user to the
	// login page

	// The best implementation of this feature would be that
	// there's a "next" query param, so that after logging in, the
	// user will be automatically redirected to the URL in "next'
	session, _ := h.cookieStore.Get(r, "session")
	var role string

	if loginInformation.UserIsAdmin {
		role = "Admin"
	} else {
		role = "Basic"
	}

	sessionCookie := NewUserSession(loginInformation.ID, role)
	storeSessionCookie(session, r, w, *&sessionCookie)

	http.Redirect(w, r, "/dashboard/home", http.StatusFound)
}

// TODO: get this right
func (h *HandlerManager) registerGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: check if the user is logged in
	// route them to the dashboard if they're logged in
	// TODO: create middleware for automatic redirection based on auth status

	// TODO: Check that each appropriate route returns a content-type
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")
	// TODO: implement a 500 page
	// TODO: handle static files with chi router
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, "/register")
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, "/register")
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) registerPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	// TODO: Should return a fragment for htmx requests

	r.ParseForm()
	firstName := r.PostFormValue("first-name")
	lastName := r.PostFormValue("last-name")
	emailAddress := r.PostFormValue("email")
	password := r.PostFormValue("password")
	confirmPassword := r.PostFormValue("confirm-password")

	if password != confirmPassword {
		// TODO: implement the validation as HTMX fragment responses to cut on work
		w.WriteHeader(http.StatusUnprocessableEntity)
		tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")

		
		log.Println(err)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors": map[string]string{
				"Password": "Password and Confirm Password fields do not match",
			},
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	// TODO: check that the user doesn't already exist
	_, err := h.store.RegisterUser(firstName, lastName, emailAddress, password)

	if err != nil {
		log.Println(err)
		var errorsMap = make(map[string]string)

		if err == ErrAccountAlreadyExists {
			errorsMap["Email"] = "Account already registered"
			w.WriteHeader(http.StatusUnprocessableEntity)
		} else {
			errorsMap["General"] = "An error occured"
			w.WriteHeader(http.StatusInternalServerError)
		}

		tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	http.Redirect(w, r, "/login", http.StatusFound)
}

func (h *HandlerManager) verifyEmailGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: check if the user is logged in
	// route them to the dashboard if they're logged in
	// TODO: create middleware for automatic redirection based on auth status

	// TODO: Check that each appropriate route returns a content-type
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/verify.html")

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, "/verify")
		return
	}

	err = tmpl.Execute(w, nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, "/verify")
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) forgotPasswordGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: check if the user is logged in
	// route them to the dashboard if they're logged in
	// TODO: create middleware for automatic redirection based on auth status

	// TODO: Check that each appropriate route returns a content-type
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/forgot-password.html")

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	err = tmpl.Execute(nil, w)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) resetPasswordGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: check if the user is logged in
	// route them to the dashboard if they're logged in
	// TODO: create middleware for automatic redirection based on auth status

	// TODO: Check that each appropriate route returns a content-type
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/reset-password.html")

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	err = tmpl.Execute(w, nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) dashboardHomeGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-home.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	session, err := h.cookieStore.Get(r, "session")

	if err != nil {
		// if there is no session, delete the cookie and redirect to the login page
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	sessionCookie, err := getSessionCookie(session)

	if err != nil {
		// cookie doesn't exist
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	userSession, err := GetSession(sessionCookie.SessionID)

	if err != nil {
		// if the session is invalid, delete the cookie and redirect to the login page
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
	}

	homeScreenInformation, err := h.store.GetHomeScreenInformation(userSession.UserID)

	if err == sql.ErrNoRows {
		// then this user doesn't exist, or there's a problem with the data in the database. Either way, we have nothing to show this user
		log.Printf("unusual edge case hit on dashboard/home route. sqlNoRows returned from GetHomeScreenInformation. %s \n", err)
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
	}

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"FirstName":   homeScreenInformation.FirstName,
		"Savings":     humanize.Comma(homeScreenInformation.SavingsBalance),
		"Loans":       homeScreenInformation.LoansBalance,
		"Investments": homeScreenInformation.InvestmentBalance,
		"Activities":  homeScreenInformation.Activities,
		// "ShowModal":   homeScreenInformation.ShowModal,
		"ShowModal": true,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)

}

func (h *HandlerManager) profileGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-profile.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	// 	session, err := h.cookieStore.Get(r, "session")

	// 	// TODO: validate the session

	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// userFromSession, err := getCookie(session)
	var userID uint = 1
	profileInformation, err := h.store.GetProfileScreenInformation(userID)

	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// 	// update the session, just in case there has been a change
	// 	session.Values["user"] = user
	// 	session.Save(r, w)

	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information":    profileInformation,
		csrf.TemplateTag: csrf.TemplateField(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) savingsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	var userID uint = 1
	savingsInformation, err := h.store.GetSavingsScreenInformation(userID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", savingsInformation)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) loansGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-loans.html",
	}

	session, err := h.cookieStore.Get(r, "session")
	cookie, err := getSessionCookie(session)

	if err != nil {
		// TODO: implement a redirect instead
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		log.Printf("failed to get user cookie")
		return
	}

	userSession, err := GetSession(cookie.SessionID)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		log.Printf("session retrieval error: %s", err)
		return
	}

	loansScreenInformation, err := h.store.GetLoansScreenInformation(userSession.UserID)

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information": loansScreenInformation,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) getLoansGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-get-loans.html",
	}

	session, _ := h.cookieStore.Get(r, "session")
	cookie, err := getSessionCookie(session)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		log.Printf("failed to get user cookie")
		return
	}

	userSession, err := GetSession(cookie.SessionID)

	if err != nil {
		http.Redirect(w, r, "/login", http.StatusUnauthorized)
		log.Printf("some session problem: %s", err)
		return
	}

	information, err := h.store.GetLoanScreenInformation(userSession.UserID)

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"showBVNField": !information.HasValidBVN,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) getLoansPostHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("Content-Type", "text/html")
	r.ParseForm()
	amount, err := strconv.ParseUint(r.FormValue("loan-amount"), 10, 64)

	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}
	// TODO: here, if the conversion fails, perhaps because of malicious entry, there won't be any repercussions
	// TODO: FIX THIS
	duration, err := strconv.ParseUint(r.FormValue("term-duration"), 10, 64)
	
	if err != nil {
		w.WriteHeader(http.StatusUnprocessableEntity)
	}

	session, err := h.cookieStore.Get(r, "session")

	if err != nil {
		log.Fatal(err)
	}

	sessionCookie, err := getSessionCookie(session)

	if err != nil {
		log.Fatal(err)
	}

	userSession, err := GetSession(sessionCookie.SessionID)

	if err != nil {
		// if the session is invalid, delete the cookie and redirect to the login page
		session.Options.MaxAge = -1
		session.Save(r, w)
		http.Redirect(w, r, "/dashboard/login", http.StatusUnauthorized)
	}

	_, err = h.store.CreateLoanApplication(userSession.UserID, amount , duration)

	// TODO: handle validation and CSRF

	
	log.Println("We are here");

	if err != nil {
		var errorsMap = make(map[string]string)
		errorsMap["Error"] = "An error occurred"
		w.WriteHeader(http.StatusUnprocessableEntity)

		log.Println(err);

		// this is the default message
		// errorsMap[""] = "An error occured"

		// if err == ErrAccountDoesNotExist {
		// 	errorsMap["Email"] = "Account does not exist"
		// }

		// if err == ErrUserNotVerified {
		// 	errorsMap["Email"] = "Verify your email before trying to log in"
		// }

		// if err == ErrPasswordIncorrect {
		// 	errorsMap["Email"] = "Email or password is incorrect"
		// }

		tmpl, err := template.ParseFiles("./web_app/templates/layouts/dashboard-base.html",
			"./web_app/templates/dashboard-get-loans.html")

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	// TODO: there should be some sort of response sent back to the user thwne they finish filling the form

	http.Redirect(w, r, "/dashboard/loans", http.StatusFound)
}

func (h *HandlerManager) investmentsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-investments.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

// TODO: This is a HTMX route. Check for accuracy later
func (h *HandlerManager) bvnModalGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	fragment, err := template.ParseFiles("./web_app/templates/fragments/bvn-modal.html")

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	err = fragment.Execute(w, nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) addBVNPostHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	bvn := r.FormValue("bvn")
	// TODO: validation
	if bvn == "" {
		// TODO: send an error
	}

	// TODO: store the BVN/send it to somewhere for validation

	w.Header().Add("Content-Type", "text/html")
	fragment, err := template.ParseFiles("./web_app/templates/fragments/verification-success.html")

	// TODO: handle err
	if err != nil {
		log.Fatalf("error with fragment %q", err)
	}

	fragment.Execute(w, VerificationData{
		Message: "Your BVN has been verified",
	})

	w.WriteHeader(201)
}

func (h *HandlerManager) familyVaultGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings-family.html",
	}

	var userID uint = 1
	familyVaultInformation, _ := h.store.GetFamilyVaultScreenInformation(userID)

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Balance": familyVaultInformation.Balance,
		"Plans":   familyVaultInformation.FamilyVaultBasicPlans,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) familyVaultPostHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: Do this
}

func (h *HandlerManager) familyVaultPlanGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings-family-plan.html",
	}

	var userID uint = 1
	planID := chi.URLParam(r, "planID")
	convertedPlanID, err := strconv.Atoi(planID)

	if err != nil {
		// TODO: return a 404 here
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	familyVaultPlanInformation, err := h.store.GetFamilyVaultPlanScreenInformation(userID, convertedPlanID)

	if err != nil {
		// TODO: here we would want to check the type of error and check for why it's having that error then display an appropriate error to the user with an appropriate error code
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information": familyVaultPlanInformation,
		"Balance":     familyVaultPlanInformation.Balance,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) soloSavingsAddFunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	decoder := json.NewDecoder(r.Body)
	var data SoloSaverAddFundsRequestType

	err := decoder.Decode(&data)

	var userID uint = 1

	// we receive the data from the frontend, then we update the DB with the information as a pending payment.

	_, err = h.store.CreateSoloSavingsPendingTransaction(userID, data.Amount, data.ReferenceNumber)

	// when that payment gets verified, then it becomes a
	// successful payment.

	// but it hardly gets verified in this flow

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	log.Printf("%v", data)
	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) soloSavingsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings-solo.html",
	}

	var userID uint = 1
	savingsInformation, err := h.store.GetSoloSaverScreenInformation(userID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information": savingsInformation,
		"Balance":     humanize.Comma(int64(savingsInformation.Balance)),
		"csrfToken":   csrf.Token(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) targetSavingsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings-target.html",
	}

	var userID uint = 1
	targetSavingsInformation, err := h.store.GetTargetSavingsScreenInformation(userID)

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information": targetSavingsInformation,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) thriftGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-thrift.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) thriftNewGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/thrift-new.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

// func (h *HandlerManager) thriftNewPostHandler(w http.ResponseWriter, r *http.Request) {
// 	w.Header().Add("Content-Type", "text/html")
// 	// TODO: Should return a fragment for htmx requests
// 	r.ParseForm()
// 	groupTitle := r.PostFormValue("title")
// 	description := r.PostFormValue("description")
// 	amount, err := strconv.ParseFloat(r.PostFormValue("amount"), 64)

// 	// TODO: should handle form validation here

// 	if err != nil {
// 		// TODO: return a validation error to the client here.
// 		http.Error(w, "Something went wrong", http.StatusInternalServerError)
// 	}

// 	numbersOfMembers, err := strconv.Atoi(r.PostFormValue("number-of-members"))

// 	if err != nil {
// 		// TODO: return a validation error to the client here.
// 		http.Error(w, "Something went wrong", http.StatusInternalServerError)
// 	}
// 	frequency := r.PostFormValue("frequency")

// 	session, err := h.cookieStore.Get(r, "session")

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	userCookie, err := getSessionCookie(session)

// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	userSession, err := GetSession(userCookie.SessionID)

// 	_, err = h.store.CreateNewThrift(userSession.UserID, groupTitle, description, amount, numbersOfMembers, frequency)

// 	if err != nil {
// 		w.WriteHeader(401)
// 		tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")

// 		if err != nil {
// 			http.Error(w, "Something went wrong", http.StatusInternalServerError)
// 		}

// 		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
// 			"Errors": map[string]string{
// 				"email": "error with email",
// 			},
// 		})
// 		return
// 	}
// 	// TODO: redirect on success
// 	// TODO: abstract the URLs
// 	http.Redirect(w, r, "/login", 301)
// }

func (h *HandlerManager) thriftPlanGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/thrift-individual.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

// getSessionCookie returns a user from session s
// on error returns an empty user
func getSessionCookie(s *sessions.Session) (*UserCookie, error) {
	val := s.Values["session_token"]
	var cookie = &UserCookie{}
	cookie, ok := val.(*UserCookie)
	if !ok {
		return &UserCookie{}, errors.New("invalid cookie")
	}
	return cookie, nil
}

func storeSessionCookie(s *sessions.Session, r *http.Request, w http.ResponseWriter, sessionCookie UserCookie) {
	// TODO: https://pkg.go.dev/github.com/gorilla/sessions#CookieStore.MaxAge states that we should check for error while saving in prod mode
	s.Values["session_token"] = sessionCookie
	s.Save(r, w)
}
