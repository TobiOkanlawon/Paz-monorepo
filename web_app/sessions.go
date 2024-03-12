package web_app

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/google/uuid"
	"github.com/gorilla/sessions"
)

type UserSession struct {
	ID             uuid.UUID
	GorillaSession *sessions.Session
	SessionStore   *memcache.Client
	UID            uint
	Expire         time.Time
}

var Session UserSession

func (u *UserSession) Create() {
	// TODO: abstract the memcache URL
	u.SessionStore = memcache.New("127.0.0.1:11211")
	u.ID = uuid.New()
}

func (u *UserSession) GetSession(sessionID string) (UserCookie, error) {
	log.Println("Getting session")
	return UserCookie{}, nil
}

func (u *UserSession) SetSession() {
	log.Println("Setting session")
}

func CheckSession(w http.ResponseWriter, r *http.Request) bool {
	cookieSession, err := r.Cookie("sessionID")
	if err != nil {
		fmt.Println("creating cookie in memcache")
		Session.Create()
		Session.Expire = time.Now().Local()
		Session.Expire.Add(time.Hour)
		Session.SetSession()
	} else {
		fmt.Println("Found cookie, checking against Memcache")
		ValidSession, err := Session.GetSession(cookieSession.Value)
		fmt.Println(ValidSession)
		return err == nil
	}
	return true
}
