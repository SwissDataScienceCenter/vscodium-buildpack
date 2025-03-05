package main

import (
	"log"
	"os"
	"strings"

	"github.com/SwissDataScienceCenter/vscodium-buildpack/cmd/install-extensions/internal"
)

func main() {
	mount_dir, ok := os.LookupEnv("RENKU_MOUNT_DIR")
	if !ok {
		log.Fatal("ROOT_DIR environment variable not set")
	}

	ext, ok := os.LookupEnv("VSCODIUM_EXTENSIONS")
	if !ok {
		log.Fatal("VSCODIUM_EXTENSIONS env var missing")
	}
	extensions := strings.Split(ext, " ")

	err := internal.Run(mount_dir, extensions)
	if err != nil {
		log.Fatal(err)
	}
}
