package main

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	mware "github.com/koshroy/turnover/middleware"
	"net/http"
)

func main() {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(mware.ActivityPubHeaders)

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})

	err := http.ListenAndServe(":3000", r)
	if err != nil {
		panic(err)
	}
}
