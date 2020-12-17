package main

import (
	"github.com/kurio/boilerplate-go/cmd"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

func main() {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := log.ParseLevel(levelStr)
	if err != nil {
		log.Fatal("LOG_LEVEL is not well-set:", level)
	}

	cmd.SetupLogs(level)

	cobra.OnInitialize(cmd.InitApp)

	if err := cmd.RootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
