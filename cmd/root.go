package cmd

import (
	"github.com/spf13/cobra"
	"time"
)

var (
	httpProcessTimeout time.Duration
)

var RootCmd = &cobra.Command{
	Use:   "goboilerplate",
	Short: "Short description.",
}

func InitApp() {
	// TODO: configurable through env
	httpProcessTimeout = 2 * time.Second
}
