package main

import (
	"os"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	httpProcessTimeout time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "goboilerplate",
	Short: "Short description.",
}

func initApp() {
	// TODO: configurable through env
	httpProcessTimeout = 2 * time.Second
}

func main() {
	levelStr := strings.ToLower(os.Getenv("LOG_LEVEL"))
	if levelStr == "" {
		levelStr = "info"
	}
	level, err := log.ParseLevel(levelStr)
	if err != nil {
		log.Fatal("LOG_LEVEL is not well-set:", level)
	}

	setupLogs(level)

	cobra.OnInitialize(initApp)

	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
