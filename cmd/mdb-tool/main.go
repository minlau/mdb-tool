package main

import (
	"flag"
	"fmt"
	"github.com/minlau/mdb-tool/store"
	"github.com/minlau/mdb-tool/web"
	"github.com/rs/zerolog/log"
	"net/http"
)

func main() {
	port := flag.Int("port", 8080, "server port")
	configFilePath := flag.String("config", "config.json", "databases config file path")
	flag.Parse()

	initLogger()

	log.Info().Msg("starting app")

	cfg, err := LoadConfig(*configFilePath)
	if err != nil {
		log.Error().Err(err).Msg("failed to load config. closing app")
		return
	}

	log.Info().Msg("starting databases initialization")

	databaseStore := store.NewDatabaseStore()
	databaseStore.AddDatabases(cfg.DatabaseConfigs)
	databaseConfigs, errs := store.GetDatabaseConfigsFromDataSources(cfg.DataSources)
	for _, errItem := range errs {
		log.Warn().Err(errItem).Msg("failed to get database configs from db")
	}
	databaseStore.AddDatabases(databaseConfigs)

	log.Info().Msg("finished databases initialization")

	log.Info().Msg("starting handlers initialization")

	r := web.New(databaseStore)

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
