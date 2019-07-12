package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	handler "github.com/kurio/boilerplate-go/internal/http"
)

var serverCmd = &cobra.Command{
	Use: "http",
	Run: func(cmd *cobra.Command, args []string) {
		address := viper.GetString("app.port")
		if address == "" {
			address = ":8080"
		}

		e := handler.NewServer()

		// Start server
		go func(address string) {
			if err := e.Start(address); err != nil {
				e.Logger.Info("shutting down the server")
			}
		}(address)

		// Wait for interrupt signal to gracefully shutdown the server with
		// a timeout of 10 seconds.
		quit := make(chan os.Signal)
		signal.Notify(quit, os.Interrupt)
		<-quit

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := e.Shutdown(ctx); err != nil {
			e.Logger.Fatal(err)
		}
	},
}

func init() {
	RootCMD.AddCommand(serverCmd)
}
