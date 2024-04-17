package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	web_backend "github.com/TobiOkanlawon/PazBackend/web_app"
)

// TODO: Make this an env variable as well
const port string = "8001"

func main() {
	secretKey, ok := os.LookupEnv("SECRET_KEY")
	if !ok {
		log.Fatalf("did not find the secret key")
	}
	paystackSecretKey, ok := os.LookupEnv("PAYSTACK_SECRET_KEY")
	if !ok {
		log.Fatalf("did not find the secret key")
	}
	paystackPublicKey, ok := os.LookupEnv("PAYSTACK_PUBLIC_KEY")
	if !ok {
		log.Fatalf("did not find the secret key")
	}
	handlerFunc, cleanUp, err := web_backend.WebAppServer([]byte(secretKey), paystackPublicKey, paystackSecretKey)
	defer cleanUp()
	if err != nil {
		log.Fatalf("error with setting up server %s \n", err)
	}
	fmt.Printf("Running the server at port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), handlerFunc))
}
