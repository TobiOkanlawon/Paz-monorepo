package web_app

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
	isValidMac := validateMAC([]byte(body), []byte(paystackSignature), []byte("secret_key"))

	if !isValidMac {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	// this is a finnicky solution, but I'm choosing it, in this moment, over the brittle one
	var requestBody map[string]any

	err := json.NewDecoder(r.Body).Decode(&requestBody)

	if err != nil {
		log.Printf("error marshalling paystack verification request body: %s \n", r.Body)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	data := requestBody["data"].(map[string]any)
	referenceNumber := data["offline_reference"].(string)
	w.WriteHeader(http.StatusOK)
	transactionInformation, err := h.store.GetPaystackVerificationInformation(referenceNumber)

	if err == ErrReferenceNumberDoesNotExist {
		// TODO: Perhaps, we should have a database table to handle suspect transactions. We can store this transaction there, for resolution purposes.

		// and then we early return
		return
	}

	// TODO: change this to a switch case
	if requestBody["event"] == "paymentrequest.success" {
		// We handle successes and failures, we can handle the rest as time goes on
		amount, err := strconv.ParseUint(data["amount"].(string), 10, 64)

		if err != nil {
			log.Printf("The conversion of the amount passed into the payment object from paystack has failed: %s", err)
		}
		
		_, err = h.store.UpdateSoloSaverPaymentInformation(amount, transactionInformation.CustomerID, transactionInformation.ReferenceNumber)

		if err != nil {
			log.Printf("Couldn't update payment information with error %s", err)
		}
		return
	}

	if requestBody["event"] == "paymentrequest.failure" {
		_, err = h.store.UpdateSoloSaverPaymentFailure(transactionInformation.ReferenceNumber)

		if err != nil {
			log.Printf("Couldn't update payment information with error %s", err)
		}
		return
	}

	return
}

func validateMAC(message, messageMAC, signingKey []byte) bool {
	// the signingKey is, in this case, the secret key from Paystack
	mac := hmac.New(sha512.New, signingKey)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}


func handlePaymentTransaction() {
	// TODO: once we can verify that the webhooks work, then we
	// should be able to implement this properly
}
