package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/Koshroy/turnover/controllers"
	"github.com/Koshroy/turnover/keystore"
	mware "github.com/Koshroy/turnover/middleware"
)

func main() {
	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Printf("could not parse config properly: %v\n", err)
		return
	}

	store, err := keystore.NewStore(config.Server.PrivateKey, config.Server.PublicKey)
	if err != nil {
		log.Printf("could not read keys properly: %v\n", err)
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(mware.ActivityPubHeaders)

	actorController := controllers.NewActor(config.Server.Scheme, config.Server.Hostname, store)

	r.Get("/actor", actorController.ServeHTTP)

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		panic(err)
	}
}
