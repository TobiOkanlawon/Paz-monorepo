package web_app

import (
	"errors"
	"time"

	// "github.com/bradfitz/gomemcache/memcache"
	"github.com/google/uuid"
	// "github.com/gorilla/sessions"
)

// The flow is that we get the sessionID from the cookie on the
// frontend, we pass it into the GetSession function, which returns
// the UserSession object that contains the ID that we can then use to
// query the DB.

// As for the session cookies, they have two expiry times. A
// frequently refreshed on, and a maximum expiry time which is the
// maximum time that a session can live.

type UserCookie struct {
	SessionID uuid.UUID
	Expiry    time.Time
}

type UserSession struct {
	UserID        uint
	SessionID     uuid.UUID
	MaximumExpiry time.Time
	Role          string
	AuthenticationStatus bool
}

var ErrNotAdmin = errors.New("user is not an admin user")

// Role is Admin, or Basic. This is used to restrict access to the admin routes

var sessionStore = make(map[uuid.UUID]UserSession)

func NewUserSession(userID uint, role string) UserCookie {
	userSession := UserSession{}
	// check that the sessionID doesn't exist already (I think
	// this might be rare, but rare isn't impossible)

	userSession.SessionID = generateID()
	userSession.UserID = userID
	userSession.Role = role
	// We are assuming that this is only used when we want to log in
	// TODO: there should be a refreshSession feature
	userSession.AuthenticationStatus = true
	// expiry should be set to about 5 minutes, this being a high
	// value application
	userSession.MaximumExpiry = time.Now().Add(5 * time.Hour)

	sessionStore[userSession.SessionID] = userSession

	cookie := UserCookie{}
	cookie.SessionID = userSession.SessionID
	cookie.Expiry = time.Now().Add(8 * time.Minute)
	// cookie.Expiry

	return cookie
}

func generateID() uuid.UUID {
	sessionID := uuid.New()

	_, ok := sessionStore[sessionID]

	if ok {
		return generateID()
	}

	return sessionID
}

func GetSession(sessionID uuid.UUID) (UserSession, error) {
	id, ok := sessionStore[sessionID]

	if !ok {
		// TODO: abstract the error definition
		return UserSession{}, errors.New("Session doesn't exist")
	}

	if id.MaximumExpiry.Before(time.Now()) {
		return UserSession{}, errors.New("User logged out")
	}

	if id.AuthenticationStatus == false {
		return UserSession{}, errors.New("user is logged out")
	}

	// TODO: this implementation doesn't handle the shorter expiry
	// it should increment the time

	return id, nil
}
