package web_app

import (
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"

	// "time"

	"fmt"

	"github.com/gorilla/csrf"
	"github.com/gorilla/sessions"
)

type HandlerManager struct {
	partialsManager IPartialsManager
	store           IStore
	cookieStore     *sessions.CookieStore
}

type LoginData struct {
	Errors    map[string]string
	csrfField string
}

type RegisterData struct {
	Errors map[string]string
}

type VerificationData struct {
	Message string
}

func NewHandlerManager(partialsManager IPartialsManager, store IStore, cookieStore *sessions.CookieStore) *HandlerManager {
	return &HandlerManager{partialsManager, store, cookieStore}
}

func (h *HandlerManager) indexGetHandler(w http.ResponseWriter, r *http.Request) {
	// TODO: this is the home page
	// TODO: check auth status, then redirect accordingly
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

// TODO: get this right
func (h *HandlerManager) loginPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	user, err := h.store.AuthenticateUser(email, password)

	// TODO: Check for the particular error
	// TODO: handle incorrect authentication details
	// TODO: handle validation and CSRF
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		tmpl, err := template.ParseFiles("./web_app/templates/login.html", "./web_app/templates/layouts/pre_auth-base.html")

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
			"Errors": map[string]string{
				"email": "Error",
			},
		})
		return
	}

	// Handling the session authentication
	session, _ := h.cookieStore.Get(r, "session")
	storeUser(session, r, w, *user)

	// TODO: redirect on success
	http.Redirect(w, r, "/dashboard/home", 301)
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

	w.WriteHeader(200)
}

func (h *HandlerManager) registerPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	// TODO: Should return a fragment for htmx requests
	r.ParseForm()
	firstName := r.PostFormValue("first-name")
	lastName := r.PostFormValue("last-name")
	emailAddress := r.PostFormValue("email")
	password := r.PostFormValue("password")
	_ = r.PostFormValue("confirm-password")

	// TODO: handle password validation

	// TODO: check that the user doesn't already exist
	_, err := h.store.RegisterUser(firstName, lastName, emailAddress, password)

	// TODO: Check for the particular error
	// TODO: handle incorrect authentication details
	// TODO: handle validation and CSRF
	if err != nil {
		w.WriteHeader(401)
		tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", RegisterData{
			Errors: map[string]string{
				"email": "error with email",
			},
		})
		return
	}
	// TODO: redirect on success
	// TODO: abstract the URLs
	http.Redirect(w, r, "/login", 301)
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
		log.Fatal(err)
	}

	userCookie, err := getCookie(session)

	if err != nil {
		log.Fatal(err)
	}

	userId := userCookie.ID
	accountInformation, err := h.store.GetAccountInformation(userId)

	if err != nil {
		log.Fatal(err)
	}

	activities, err := h.store.GetActivities(userId)
	fmt.Println(activities)

	err = tmpl.ExecuteTemplate(w, "base", map[string]interface{}{
		"Savings":     accountInformation.SavingsAmount,
		"Loans":       accountInformation.LoansAmount,
		"Investments": accountInformation.InvestmentsAmount,
		"Activities":  activities,
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

	session, err := h.cookieStore.Get(r, "session")

	// TODO: validate the session

	if err != nil {
		log.Fatal(err)
	}

	userFromSession, err := getCookie(session)

	user, err := h.store.GetUserInformation(userFromSession.ID)

	if err != nil {
		log.Fatal(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// update the session, just in case there has been a change
	session.Values["user"] = user
	session.Save(r, w)

	if err != nil {
		log.Fatal(err)
	}

	data := map[string]interface{}{
		"FirstName":    user.FirstName,
		"LastName":     user.LastName,
		"EmailAddress": user.EmailAddress,
		"PhoneNumber":  user.PhoneNumber,
		"DateOfBirth":  user.DateOfBirth,
		"Sex":          user.Sex,
	}

	err = tmpl.ExecuteTemplate(w, "base", data)

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

	err := tmpl.ExecuteTemplate(w, "base", nil)

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

	session, _ := h.cookieStore.Get(r, "session")
	user, err := getCookie(session)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("failed to get user cookie")
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	_, err = h.store.GetLoansInformation(user.ID)

	// data := map[string]interface{}{
	// 	"Balance":  loansData.Balance,
	// 	"Activity": loansData.Activity,
	// }

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err = tmpl.ExecuteTemplate(w, "base", nil)

	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		log.Printf("error %q from url %q", err, r.URL.Path)
		return
	}

	w.WriteHeader(200)
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

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", nil)

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

	tmpl := template.Must(template.ParseFiles(templateFiles...))

	err := tmpl.ExecuteTemplate(w, "base", nil)

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

func (h *HandlerManager) thriftNewPostHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/html")
	// TODO: Should return a fragment for htmx requests
	r.ParseForm()
	groupTitle := r.PostFormValue("title")
	description := r.PostFormValue("description")
	amount, err := strconv.ParseFloat(r.PostFormValue("amount"), 64)

	// TODO: should handle form validation here

	if err != nil {
		// TODO: return a validation error to the client here.
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}

	numbersOfMembers, err := strconv.Atoi(r.PostFormValue("number-of-members"))

	if err != nil {
		// TODO: return a validation error to the client here.
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
	}
	frequency := r.PostFormValue("frequency")

	session, err := h.cookieStore.Get(r, "session")

	if err != nil {
		log.Fatal(err)
	}

	userCookie, err := getCookie(session)

	if err != nil {
		log.Fatal(err)
	}

	userId := userCookie.ID

	_, err = h.store.CreateNewThrift(userId, groupTitle, description, amount, numbersOfMembers, frequency)

	if err != nil {
		w.WriteHeader(401)
		tmpl, err := template.ParseFiles("./web_app/templates/register.html", "./web_app/templates/layouts/pre_auth-base.html")

		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
		}

		tmpl.ExecuteTemplate(w, "base", RegisterData{
			Errors: map[string]string{
				"email": "error with email",
			},
		})
		return
	}
	// TODO: redirect on success
	// TODO: abstract the URLs
	http.Redirect(w, r, "/login", 301)
}

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

// getCookie returns a user from session s
// on error returns an empty user
func getCookie(s *sessions.Session) (*UserCookie, error) {
	val := s.Values["user"]
	var cookie = &UserCookie{}
	cookie, ok := val.(*UserCookie)
	if !ok {
		return &UserCookie{}, errors.New("invalid cookiel")
	}
	return cookie, nil
}

func storeUser(s *sessions.Session, r *http.Request, w http.ResponseWriter, user User) error {
	var userCookie UserCookie

	userCookie.ID = user.ID
	userCookie.FirstName = user.FirstName
	userCookie.LastName = user.LastName

	s.Values["user"] = userCookie
	s.Save(r, w)

	return nil
}
