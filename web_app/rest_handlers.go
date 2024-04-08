package web_app

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"net/http"
)

func (h *HandlerManager) paystackVerificationWebhook(w http.ResponseWriter, r *http.Request) {
	// This is a POST endpoint that we'll give to paystack as a
	// webhook to verify payments

	// First, we verify that the request is coming from paystack,
	// by checking the hash key sent in the request
	
	paystackSignature := r.Header.Get("x-paystack-signature")
	var body string
	decoder := json.NewDecoder(r.Body)
	decoder.Decode(&body)
	// TODO: a good solution is to store the secret key as its byte representation to avoid the constant transformation
	isValidMac := validMAC([]byte(body), []byte(paystackSignature), []byte(h.paystackSecretKey))

	if !isValidMac {
		w.WriteHeader(http.StatusOK)
	}

	// Then we check that the type of transaction that the webhook
	// is trying to represent, and pass it to the appropriate
	// function.

	// we always have to acknowledge that we have received the ping
	w.WriteHeader(http.StatusOK)
}

func validMAC(message, messageMAC, signingKey []byte) bool {
	// the signingKey is, in this case, the secret key from Paystack
	mac := hmac.New(sha512.New, signingKey)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}

func handlePaymentTransaction() {
	
}
