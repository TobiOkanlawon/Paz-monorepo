package web_app

import "testing"

func TestEmailValidation(t *testing.T) {
	t.Run("returns false if sent an empty email string", func(t *testing.T) {
		got := validateEmail("")

		if got != false {
			t.Error("expected validation to fail when passed an empty string")
		}
	})

	t.Run("returns true if passed a valid email", func (t *testing.T) {
		got := validateEmail("email@domain.com")

		if got != true {
			t.Error("expected validation to pass when passed a valid email")
		}
	})

	t.Run("fails when passed a reall weird string", func(t *testing.T) {
		email := "jasdlkfas@ajskdjf"
		got := validateEmail(email)

		if got != false {
			t.Errorf("expected validation to fail when passed a weird string %s", email)
		}
	})

	t.Run("fails when passed a single space", func (t *testing.T) {
		email := " "
		got := validateEmail(email)

		if got != false {
			t.Error("expected validation to fail when passed an empty space")
		}
	})
}
