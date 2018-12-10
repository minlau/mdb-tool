package main

import (
	"flag"
	"fmt"
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	configFilePath := flag.String("config", "databases.json", "databases config file path")
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
	databaseConfigs := GetDatabaseConfigsFromDataSources(config.DataSources)
	databaseStore.AddDatabases(databaseConfigs)

	log.Info().Msg("finished databases initialization")

	log.Info().
		Int("port", *port).
		Msg("starting web service")

	r := chi.NewRouter()
	r.Use(middleware.DefaultCompress)
	r.Use(ZeroLogLogger)
	r.Use(middleware.Recoverer)
	initHandlers(r, databaseStore)
	err = http.ListenAndServe(fmt.Sprintf(":%d", *port), r)
	if err != nil {
		log.Error().Err(err).Msg("failed to start web service. closing app")
		return
	}
}
