package main

import (
	"github.com/ovh/jerem/src/cmd"

	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		log.WithError(err).Fatal("Could not execute jerem")
	}
}
