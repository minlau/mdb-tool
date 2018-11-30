package main

import (
	"flag"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
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

	log.Info().Msgf("starting web service with port: %d", *port)

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	initHandlers(r, databaseStore)

	err = r.Run(fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Error().Err(err).Msg("failed to start web service. closing app")
		return
	}
}
