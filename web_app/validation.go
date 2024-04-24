package web_app

import "regexp"

var rxEmail = regexp.MustCompile(".+@.+\\..+")

/*
   Takes an email and returns true or false based on whether it fits a
   particular email format
 */
func validateEmail(email string) (bool) {
	match := rxEmail.Match([]byte(email))
	return match
}
