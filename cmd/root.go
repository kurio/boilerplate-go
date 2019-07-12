package cmd

import (
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var RootCMD = &cobra.Command{
	Use: "app",
}

func init() {
	initConfig()
	setupLogger()
}

func initConfig() {
	viper.AutomaticEnv()

	replacer := strings.NewReplacer(".", "_")
	viper.SetEnvKeyReplacer(replacer)
}

func setupLogger() {
	log.SetReportCaller(true)

	if viper.GetBool("app.debug") {
		log.SetLevel(log.DebugLevel)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
		})

		log.Warn("app is running is debug mode")
	} else {
		log.SetLevel(log.InfoLevel)
		log.SetFormatter(&log.JSONFormatter{})
	}
}
