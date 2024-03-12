package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	web_backend "github.com/TobiOkanlawon/PazBackend/web_app"
	"github.com/gorilla/csrf"
)

const port string = "8001"

func main() {
	secretKey, ok := os.LookupEnv("SECRET_KEY")
	if !ok {
		log.Fatalf("did not find the secret key")
	}
	handlerFunc, err := web_backend.WebAppServer([]byte(secretKey))
	if err != nil {
		log.Fatalf("error with setting up server %s \n", err)
	}
	fmt.Printf("Running the server at port %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), csrf.Protect(
		[]byte(secretKey),
		csrf.Path("/dashboard"),
	)(handlerFunc)))
}
