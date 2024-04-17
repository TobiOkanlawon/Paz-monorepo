package web_app

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/google/uuid"
)

type PaystackGenericData struct {
	Event string `json:"event"`
}

type PaystackPaymentSuccessful struct {
	Data PaystackPaymentSuccessfulDataObject
}

type PaystackPaymentSuccessfulDataObject struct {
	Amount          uint64    `json:"amount"`
	ReferenceNumber uuid.UUID `json:"offline_reference"`
}

func (h *HandlerManager) paystackVerificationWebhook(w http.ResponseWriter, r *http.Request) {
	// This is a POST endpoint that we'll give to paystack as a
	// webhook to verify payments

	// First, we verify that the request is coming from paystack,
	// by checking the hash key sent in the request

	paystackSignature := r.Header.Get("x-paystack-signature")
	bodyAsBytes, err := io.ReadAll(r.Body)

	if err != nil {
		log.Printf("an error occured while trying to read body %s", err)
	}

	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyAsBytes))

	isValidMac := validateMAC(bodyAsBytes, []byte(paystackSignature), []byte("secret_key"))
	if !isValidMac {
		w.WriteHeader(http.StatusForbidden)
		return
	}

	var jsonBody PaystackGenericData
	err = json.NewDecoder(r.Body).Decode(&jsonBody)
	r.Body = ioutil.NopCloser(bytes.NewBuffer(bodyAsBytes))

	if err != nil {
		log.Printf("error while trying to decode data %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: a good solution is to store the secret key as its byte representation to avoid the constant transformation

	// TODO: change this to a switch case
	if jsonBody.Event == "paymentrequest.success" {
		// marshal the rest of the data into the paymentSuccess object
		var data PaystackPaymentSuccessful
		err := json.NewDecoder(r.Body).Decode(&data)

		if err != nil {
			log.Printf("error while trying to decode paymentrequest.success: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// We handle successes and failures, we can handle the rest as time goes on

		if err != nil {
			log.Printf("The conversion of the amount passed into the payment object from paystack has failed: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = h.store.UpdateSoloSaverPaymentInformation(data.Data.Amount, data.Data.ReferenceNumber)

		if err != nil {
			log.Printf("Couldn't update payment information with error %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		return
	}

	// if requestBody["event"] == "paymentrequest.failure" {
	// 	_, err = h.store.UpdateSoloSaverPaymentFailure(transactionInformation.ReferenceNumber)

	// 	if err != nil {
	// 		log.Printf("Couldn't update payment information with error %s", err)
	// 	}
	// 	return
	// }

	return
}

func validateMAC(message, messageMAC, signingKey []byte) bool {
	// the signingKey is, in this case, the secret key from Paystack
	mac := hmac.New(sha512.New, signingKey)
	mac.Write(message)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(messageMAC, expectedMAC)
}
