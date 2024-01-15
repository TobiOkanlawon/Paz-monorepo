package main

import (
	"fmt"
	"log"
	"net/http"

	web_backend "github.com/TobiOkanlawon/PazBackend/web_app"
)

func main() {
	handler_func, err := web_backend.WebAppServer()
	if err != nil {
		log.Fatalf("error with setting up server %s", err)
	}
	fmt.Println("Running the server at port 8001")
	log.Fatal(http.ListenAndServe(":8001", handler_func))
}
