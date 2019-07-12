package main

import (
	"github.com/kurio/boilerplate-go/cmd"
	log "github.com/sirupsen/logrus"
)

func main() {
	if err := cmd.RootCMD.Execute(); err != nil {
		log.Fatal(err)
	}
}
