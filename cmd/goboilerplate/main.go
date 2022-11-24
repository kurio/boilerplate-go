package main

import (
	"github.com/sirupsen/logrus"
)

func main() {
	if err := rootCMD.Execute(); err != nil {
		logrus.Fatalf("error executing command: %+v", err)
	}
}
