package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/sembh1998/feed-case/database"
	"github.com/sembh1998/feed-case/events"
	"github.com/sembh1998/feed-case/repository"
	"github.com/sembh1998/feed-case/search"
)

type Config struct {
	PostgresDB          string `envconfig:"POSTGRES_DB"`
	PostgresuUser       string `envconfig:"POSTGRES_USER"`
	PostgresPassword    string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress         string `envconfig:"NATS_ADDRESS"`
	ElasticSearchAddess string `envconfig:"ELASTICSEARCH_ADDRESS"`
}

func newRouter() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/feeds", listFeedHandler).Methods(http.MethodGet)
	router.HandleFunc("/search", searchHandler).Methods(http.MethodGet)
	return
}

func main() {
	var cfg Config
	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatalf("%v", err)
	}

	addr := fmt.Sprintf("postgres://%s:%s@postgres/%s?sslmode=disable", cfg.PostgresuUser, cfg.PostgresPassword, cfg.PostgresDB)

	repo, err := database.NewPostgresRepository(addr)
	if err != nil {
		log.Fatal(err)
	}
	repository.SetRepository(repo)

	es, err := search.NewElastic(fmt.Sprintf("http://%s", cfg.ElasticSearchAddess))
	if err != nil {
		log.Fatal(err)
	}
	search.SetSearchRepository(es)
	defer search.Close()

	n, err := events.NewNats(fmt.Sprintf("nats://%s", cfg.NatsAddress))
	if err != nil {
		log.Fatal(err)
	}
	err = n.OnCreatedFeed(onCreatedFeed)

	if err != nil {
		log.Fatal(err)
	}

	events.SetEventStore(n)

	defer events.Close()

	router := newRouter()
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
