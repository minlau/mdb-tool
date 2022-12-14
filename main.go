package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	configFilePath := flag.String("config", "config.json", "databases config file path")
	flag.Parse()

	initLogger()

	log.Info().Msg("starting app")

	config, err := readConfig(*configFilePath)
	if err != nil {
		log.Error().Err(err).Msg("failed to read config. closing app")
		return
	}

	log.Info().Msg("starting databases initialization")

	databaseStore := NewDatabaseStore()
	databaseStore.AddDatabases(config.DatabaseConfigs)
	databaseConfigs, errs := GetDatabaseConfigsFromDataSources(config.DataSources)
	for _, errItem := range errs {
		log.Warn().Err(errItem).Msg("failed to get database configs from db")
	}
	databaseStore.AddDatabases(databaseConfigs)

	log.Info().Msg("finished databases initialization")

	log.Info().Msg("starting handlers initialization")

	r := chi.NewRouter()
	r.Use(middleware.Compress(1))
	r.Use(ZeroLogLogger)
	r.Use(middleware.Recoverer)

	initHandlers(r, databaseStore)

	log.Info().Msg("finished handlers initialization")

	log.Info().
		Int("port", *port).
		Msg("starting http server")

	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
	if err != nil {
		log.Error().Err(err).Msg("failed to start http server")
		return
	}
}
