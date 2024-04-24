package web_app

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/TobiOkanlawon/go-sanatio"
	"github.com/dustin/go-humanize"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

func NewHandlerManager(partialsManager IPartialsManager, store IStore, cookieStore *sessions.CookieStore, paystackPublicKey, paystackSecretKey string) *HandlerManager {
	return &HandlerManager{partialsManager, store, cookieStore, paystackPublicKey, paystackSecretKey}
}

func (h *HandlerManager) indexGetHandler(w http.ResponseWriter, r *http.Request) {

	// For the while that we don't have an introductory website
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

func (h *HandlerManager) loginGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: check if the user is logged in
	// route them to the dashboard if they're logged in
	// TODO: create middleware for automatic redirection based on auth status

	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/login.html", "./web_app/templates/layouts/pre_auth-base.html")

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

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	tmpl, err := template.ParseFiles("./web_app/templates/login.html", "./web_app/templates/layouts/pre_auth-base.html")
	r.ParseForm()
	var errorsMap = make(map[string]string)
	email := r.PostFormValue("email")
	if validateEmail(email) == false {
		errorsMap["Email"] = "Your email is invalid"
		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors": errorsMap,
		})
		return
	}

	// validate password and return if it fails
	passwordValidation := sanatio.NewStringValidator().SetValue(r.PostFormValue("password")).Required()
	passwordErrors := passwordValidation.GetErrors()
	if len(passwordErrors) != 0 {
		errorsMap["Password"] = "You have to insert a password"
		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})

		return
	}

	password, _ := passwordValidation.GetValue()
	loginInformation, err := h.store.AuthenticateUser(email, password)

	// TODO: handle validation and CSRF

	if err != nil {
		var errorsMap = make(map[string]string)
		log.Println(err)
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

	tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")

	r.ParseForm()
	var errorsMap = make(map[string]string)
	firstName := r.PostFormValue("first-name")
	lastName := r.PostFormValue("last-name")
	emailAddress := r.PostFormValue("email")

	if validateEmail(emailAddress) == false {
		errorsMap["Email"] = "Your email is invalid"
		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors": errorsMap,
		})
		return
	}

	password := r.PostFormValue("password")
	passwordValidator := sanatio.NewStringValidator().SetValue(password).Required()
	if len(passwordValidator.GetErrors()) != 0 {
		errorsMap["Password"] = "There's something wrong with your password"
	}

	confirmPassword := r.PostFormValue("confirm-password")
	confirmPasswordValidator := sanatio.NewStringValidator().SetValue(confirmPassword).Required()

	if len(confirmPasswordValidator.GetErrors()) != 0 {
		errorsMap["Password"] = "There's something wrong with your password"
	}

	if password != confirmPassword {
		// TODO: implement this password != confirmPassword as a sanatio validation
		// TODO: implement the validation as HTMX fragment responses to cut on work
		w.WriteHeader(http.StatusUnprocessableEntity)

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	if len(errorsMap) != 0 {
		w.WriteHeader(http.StatusUnprocessableEntity)

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	// TODO: check that the user doesn't already exist
	_, err = h.store.RegisterUser(firstName, lastName, emailAddress, password)

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

	// TODO: not implemented yet, so it's a teapot
	w.WriteHeader(http.StatusTeapot)
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

	userSession, _ := h.getSessionOrLogout(w, r)

	homeScreenInformation, err := h.store.GetHomeScreenInformation(userSession.UserID)

	if err == sql.ErrNoRows {
		// then this user doesn't exist, or there's a problem with the data in the database. Either way, we have nothing to show this user
		log.Printf("unusual edge case hit on dashboard/home route. sqlNoRows returned from GetHomeScreenInformation. %s \n", err)
		h.logout(w, r)
	}

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"FirstName":   homeScreenInformation.FirstName,
		"Savings":     humanize.Comma(homeScreenInformation.SavingsBalance),
		"Loans":       humanize.Comma(homeScreenInformation.LoansBalance),
		"Investments": humanize.Comma(homeScreenInformation.InvestmentBalance),
		"Activities":  homeScreenInformation.Activities,
		// "ShowModal":   homeScreenInformation.ShowModal,
		"ShowModal": false,
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

	userSession, _ := h.getSessionOrLogout(w, r)
	profileInformation, err := h.store.GetProfileScreenInformation(userSession.UserID)

	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

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

	userSession, _ := h.getSessionOrLogout(w, r)

	savingsInformation, err := h.store.GetSavingsScreenInformation(userSession.UserID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Balance": humanize.Comma(int64(savingsInformation.Balance)),
	})

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

	userSession, _ := h.getSessionOrLogout(w, r)

	loansScreenInformation, err := h.store.GetLoansScreenInformation(userSession.UserID)

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Balance":         humanize.Comma(loansScreenInformation.Balance),
		"HasPendingLoans": loansScreenInformation.HasPendingLoans,
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

	userSession, _ := h.getSessionOrLogout(w, r)

	// TODO: check if the user has pending loans, or a pending loan application, then turn the user down

	// TODO: Do so on the backend too, the user shouldn't be able
	// to submit a loan application if they have a pending loan
	information, err := h.store.GetLoanScreenInformation(userSession.UserID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		csrf.TemplateTag: csrf.TemplateField(r),
		"ShowBVNField":   !information.HasValidBVN,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) getLoansPostHandler(w http.ResponseWriter, r *http.Request) {

	userSession, _ := h.getSessionOrLogout(w, r)

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

	// TODO: Add bank statement upload to the loans application page

	bvn := r.PostFormValue("bvn")
	var transformedBvn uint64
	if bvn != "" {
		transformedBvn, err = strconv.ParseUint(bvn, 10, 64)

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			log.Printf("error %q from url %q", err, r.URL.Path)
			return
		}
	}

	// TODO: this transformedBVN implementation is not clean, but it's what I've got for now
	_, err = h.store.CreateLoanApplication(userSession.UserID, amount, duration, transformedBvn)

	// TODO: handle validation and CSRF

	if err != nil {
		var errorsMap = make(map[string]string)
		errorsMap["Error"] = "An error occurred"
		w.WriteHeader(http.StatusUnprocessableEntity)

		log.Println(err)

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

	// TODO: while we are still using the stop-gap implementation, we don't need the userSession. However, we still use it to log out the user
	userSession, _ := h.getSessionOrLogout(w, r)
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-investments.html",
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	information, err := h.store.GetInvestmentsScreenInformation(userSession.UserID)

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Balance": information.Balance,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) investmentsFormGetHandler(w http.ResponseWriter, r *http.Request) {
	h.getSessionOrLogout(w, r)
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-investments-form.html",
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

func (h *HandlerManager) investmentsFormPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	userSession, _ := h.getSessionOrLogout(w, r)
	r.ParseForm()
	var errorsMap = make(map[string]string)

	// TODO: Redo the validation with sanatio

	employmentStatus := r.PostFormValue("employment-status")

	if employmentStatus == "" {
		errorsMap["EmploymentStatus"] = "Something seems wrong with this field"
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	if employmentStatus == "salaried" {
		employmentStatus = "SALARIED"
	} else if employmentStatus == "self-employed" {
		employmentStatus = "SELF-EMPLOYED"
	} else if employmentStatus == "retired" {
		employmentStatus = "RETIRED"
	} else if employmentStatus == "unemployed" {
		employmentStatus = "UNEMPLOYED"
	} else {
		errorsMap["EmploymentStatus"] = "You have selected an invalid employment status"
	}

	yearOfEmploymentValue := r.PostFormValue("date-of-employment")

	if yearOfEmploymentValue == "" {
		errorsMap["DateOfEmployment"] = "This field is required"
	}

	convertedYearOfEmployment, err := time.Parse("YYYY-MM-DD", yearOfEmploymentValue)

	employerName := r.PostFormValue("employer-name")

	if employerName == "" {
		// TODO: make a validation abstraction
		errorsMap["EmployerName"] = "This field is required"
	}

	amount, err := strconv.ParseUint(r.PostFormValue("investment-amount"), 10, 64)

	if err != nil {
		errorsMap["InvestmentAmount"] = "Something seems wrong with this field"
	}

	// implement a maximum for this field
	tenure, err := strconv.ParseUint(r.PostFormValue("investment-tenure"), 10, 64)

	if err != nil {
		errorsMap["InvestmentTenure"] = "Something seems wrong with this field"
	}

	taxIdentificationNumber, err := strconv.ParseUint(r.PostFormValue("TIN"), 10, 64)

	if err != nil {
		errorsMap["TIN"] = "Something seems wrong with this field"
	}

	bankAccountName := r.PostFormValue("bank-account-name")

	if bankAccountName == "" {
		errorsMap["BankAccountName"] = "This field is required"
	}

	bankAccountNumber, err := strconv.ParseUint(r.PostFormValue("bank-account-number"), 10, 64)

	if err != nil {
		errorsMap["BankAccountNumber"] = "Something seems wrong with this field"
	}

	_, err = h.store.CreateInvestmentApplication(userSession.UserID, employmentStatus, convertedYearOfEmployment, employerName, amount, tenure, taxIdentificationNumber, bankAccountName, bankAccountNumber)

	// TODO: handle validation and CSRF

	if err != nil {
		errorsMap["Error"] = "An error occurred"
		w.WriteHeader(http.StatusUnprocessableEntity)

		log.Println(err)

		tmpl, err := template.ParseFiles(
			"./web_app/templates/layouts/dashboard-base.html",
			"./web_app/templates/dashboard-investments-form.html",
		)

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors":         errorsMap,
			csrf.TemplateTag: csrf.TemplateField(r),
		})
		return
	}

	// TODO: there should be some sort of response sent back to the user when they finish filling the form

	http.Redirect(w, r, "/dashboard/investments", http.StatusFound)
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

	userSession, _ := h.getSessionOrLogout(w, r)

	familyVaultInformation, err := h.store.GetFamilyVaultScreenInformation(userSession.UserID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Balance":        familyVaultInformation.Balance,
		"Plans":          familyVaultInformation.FamilyVaultBasicPlans,
		csrf.TemplateTag: csrf.TemplateField(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) familyVaultPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	// TODO: Should return a fragment for htmx requests

	userSession, _ := h.getSessionOrLogout(w, r)

	r.ParseForm()
	familyName := r.PostFormValue("family-name")
	familyMember := r.PostFormValue("family-member")
	amount, err := strconv.ParseFloat(r.PostFormValue("amount"), 64)
	savingsFrequency := r.PostFormValue("savings-frequency")
	savingsDuration, err := strconv.ParseInt(r.PostFormValue("duration"), 10, 64)

	// TODO: should handle form validation here

	if err != nil {
		// TODO: return a validation error to the client here.
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	information, err := h.store.CreateNewFamilyVault(userSession.UserID, familyName, familyMember, amount, savingsFrequency, savingsDuration)

	if err != nil {
		log.Printf("An error occured while trying to create a family vault %s", err)
		h.logout(w, r)
		return
	}
	// TODO: redirect on success
	// TODO: abstract the URLs
	http.Redirect(w, r, fmt.Sprintf("/dashboard/savings/family-vault/%d", information.PlanID), http.StatusOK)
}

func (h *HandlerManager) familyVaultPlanGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings-family-plan.html",
	}

	userSession, _ := h.getSessionOrLogout(w, r)

	planID := chi.URLParam(r, "planID")
	convertedPlanID, err := strconv.Atoi(planID)

	if err != nil {
		// TODO: return a 404 here
		http.Error(w, "Something went wrong", http.StatusNotFound)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	familyVaultPlanInformation, err := h.store.GetFamilyVaultPlanScreenInformation(userSession.UserID, convertedPlanID)

	if err != nil {
		// TODO: here we would want to check the type of error and check for why it's having that error then display an appropriate error to the user with an appropriate error code
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information":     familyVaultPlanInformation,
		"Balance":         familyVaultPlanInformation.Balance,
		"csrfToken":       csrf.Token(r),
		"ReferenceNumber": h.generatePaymentUUID(),
		"PlanID":          planID,
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) soloSavingsAddFunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	userSession, _ := h.getSessionOrLogout(w, r)

	decoder := json.NewDecoder(r.Body)
	var data SoloSaverAddFundsRequestType

	err := decoder.Decode(&data)

	// we receive the data from the frontend, then we update the DB with the information as a pending payment.

	// when that payment gets verified, then it becomes a
	// successful payment.

	// but it hardly gets verified in this flow

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	// the planID field is 99909990 by convention. It doesn't mean anything for soloSavings
	_, err = h.store.CreatePayment(userSession.UserID, 99909990, data.ReferenceNumber, "SOLO_SAVINGS", data.Amount)

	if err != nil {
		http.Error(w, "Something went wrong while trying to save your transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) familySavingsAddFunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	userSession, _ := h.getSessionOrLogout(w, r)

	decoder := json.NewDecoder(r.Body)
	var data SoloSaverAddFundsRequestType

	err := decoder.Decode(&data)

	planID := chi.URLParam(r, "planID")
	convertedPlanID, err := strconv.Atoi(planID)

	// we receive the data from the frontend, then we update the DB with the information as a pending payment.

	// when that payment gets verified, then it becomes a
	// successful payment.

	// but it hardly gets verified in this flow

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	// the planID field is 99909990 by convention. It doesn't mean anything for soloSavings
	_, err = h.store.CreatePayment(userSession.UserID, uint(convertedPlanID), data.ReferenceNumber, "FAMILY_SAVINGS", data.Amount)

	if err != nil {
		http.Error(w, "Something went wrong while trying to save your transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) targetSavingsAddFunds(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	userSession, _ := h.getSessionOrLogout(w, r)

	decoder := json.NewDecoder(r.Body)
	var data SoloSaverAddFundsRequestType

	err := decoder.Decode(&data)

	planID := chi.URLParam(r, "planID")
	convertedPlanID, err := strconv.Atoi(planID)

	// we receive the data from the frontend, then we update the DB with the information as a pending payment.

	// when that payment gets verified, then it becomes a
	// successful payment.

	// but it hardly gets verified in this flow

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusBadRequest)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	// the planID field is 99909990 by convention. It doesn't mean anything for soloSavings
	_, err = h.store.CreatePayment(userSession.UserID, uint(convertedPlanID), data.ReferenceNumber, "TARGET_SAVINGS", data.Amount)

	if err != nil {
		http.Error(w, "Something went wrong while trying to save your transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HandlerManager) soloSavingsGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	templateFiles := []string{
		"./web_app/templates/layouts/dashboard-base.html",
		"./web_app/templates/dashboard-savings-solo.html",
	}

	userSession, _ := h.getSessionOrLogout(w, r)
	savingsInformation, err := h.store.GetSoloSaverScreenInformation(userSession.UserID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information":       savingsInformation,
		"Balance":           humanize.Comma(int64(savingsInformation.Balance)),
		"csrfToken":         csrf.Token(r),
		"ReferenceNumber":   h.generatePaymentUUID(),
		"PublicKey":         h.paystackPublicKey,
		"HasPendingPayment": savingsInformation.HasPendingPayment,
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

	userSession, _ := h.getSessionOrLogout(w, r)
	targetSavingsInformation, err := h.store.GetTargetSavingsScreenInformation(userSession.UserID)

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Information":    targetSavingsInformation,
		csrf.TemplateTag: csrf.TemplateField(r),
	})

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
}

func (h *HandlerManager) targetSavingsPostHandler(w http.ResponseWriter, r *http.Request) {
	// w.Header().Add("Content-Type", "text/html")
	// userSession := h.getSessionOrLogout(w, r)

	// r.ParseForm()

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

func (h *HandlerManager) adminHomeGetHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")

	templateFiles := []string{
		"./web_app/templates/admin/base.html",
		"./web_app/templates/admin/index.html",
	}

	userSession, err := h.getAdminSessionOrLogout(w, r)

	if err == ErrNotAdmin {
		// This block is very important.
		// Without this block, users will still be able to see the content of the page
		return
	}

	information, err := h.store.GetAdminHomeScreenInformation(userSession.UserID)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	// TODO: check that the user is an admin while getting their sessionCookie

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"LoanRequests":        information.LoanRequests,
		"InvestmentsRequests": information.InvestmentsRequests,
		"WithdrawalRequests":  information.WithdrawalRequests,
	})

	if err != nil {
		http.Error(w, "Something went wrong while trying to load the admin home page", http.StatusInternalServerError)
	}
}

func (h *HandlerManager) logoutGetHandler(w http.ResponseWriter, r *http.Request) {
	h.logout(w, r)
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

func (h *HandlerManager) getSessionOrLogout(w http.ResponseWriter, r *http.Request) (UserSession, error) {
	session, err := h.cookieStore.Get(r, "session")

	if err != nil {
		// session could not be decoded, for some reason
		// TODO: Find out what it means for the session to not be decoded
		log.Printf("session could not be decoded: %s \n", err)
	}

	sessionCookie, err := getSessionCookie(session)

	if err != nil {
		// cookie doesn't exist
		h.logout(w, r)
		return UserSession{}, err
	}

	userSession, err := GetSession(sessionCookie.SessionID)

	if err != nil {
		// if the session is invalid, delete the cookie and redirect to the login pag

		h.logout(w, r)
		return UserSession{}, err
	}

	return userSession, nil
}

func (h *HandlerManager) getAdminSessionOrLogout(w http.ResponseWriter, r *http.Request) (UserSession, error) {
	session, err := h.cookieStore.Get(r, "session")

	if err != nil {
		log.Printf("session could not be decoded %s \n", err)
		return UserSession{}, err
	}

	sessionCookie, err := getSessionCookie(session)

	if err != nil {
		// cookie doesn't exist
		h.logout(w, r)
		return UserSession{}, err
	}

	userSession, err := GetSession(sessionCookie.SessionID)

	if userSession.Role != "admin" {
		http.Error(w, "Forbidden: not an admin", http.StatusForbidden)
		return UserSession{}, ErrNotAdmin
	}

	if err != nil {
		h.logout(w, r)
		return UserSession{}, err
	}

	return userSession, nil
}

func (h *HandlerManager) logout(w http.ResponseWriter, r *http.Request) {

	session, err := h.cookieStore.Get(r, "session")

	if err != nil {
		// session could not be decoded
		log.Printf("session could not be decoded: %s \n", err)
	}

	session.Options.MaxAge = -1
	session.Save(r, w)
	http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
}

// generates UUIDs for payment related purposes
func (h *HandlerManager) generatePaymentUUID() uuid.UUID {

	// This is the better solution, but we can use the less well engineered one for now
	// generate a UUID, check the database that it doesn't already exist
	// then return it if it doesn't, if it does, start again

	// There's an edge case here, it's hard to hit, but there's a possible race condition on the UUID and check thing, so a lock would be the actually totally correct solution. However, since it's an almost impossible edge case, it'll just be overengineering

	return uuid.New()
}
