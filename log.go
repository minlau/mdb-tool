package main

import (
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog"
	"os"
)

func initLogger() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
	zerolog.TimeFieldFormat = "2006-01-02 15:04:05.000000"
}
