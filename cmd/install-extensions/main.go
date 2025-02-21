package main

import (
	"log"
	"os"

	"github.com/SwissDataScienceCenter/vscodium-buildpack/cmd/install-extensions/internal"
	"github.com/SwissDataScienceCenter/vscodium-buildpack/cmd/util"
)

func main() {
	err := internal.Run(util.LoadEnvironmentMap(os.Environ()))
	if err != nil {
		log.Fatal(err)
	}
}
