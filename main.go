package main

import (
	"flag"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func main() {
	configFilePath := flag.String("config", "databases.json", "databases config file path")

	initLogger()
	log.Debug().Msg("starting app")

	//databaseConfigs, err := getDatabaseConfigsFromDb()
	databaseConfigs, err := getDatabaseConfigsFromFile(*configFilePath)
	if err != nil {
		log.Error().Err(err).Msg("failed to initialize config. Closing app")
		return
	}

	log.Debug().Msg("starting databases initialization")

	databaseStore := NewDatabaseStore()
	err = databaseStore.AddDatabases(databaseConfigs)
	if err != nil {
		log.Error().Err(err).Msg("failed to add databases. Closing app")
		return
	}
	log.Debug().Msg("finished databases initialization")

	log.Debug().Msg("starting web service with port: 8079")

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	initHandlers(r, databaseStore)

	err = r.Run(":8079")
	if err != nil {
		log.Error().Err(err).Msg("failed to start web service. Closing app")
		return
	}
}
