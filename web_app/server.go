package web_app

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func generateRouter(){
	r := chi.NewRouter()
	r.Use(middleware.Logger())
}

func StartConfigurableWebAppServer(routes map[string][]byte) http.HandlerFunc{
	
}
