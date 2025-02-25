package vscodiumbuildpack

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Parser struct{}

func NewParser() Parser {
	return Parser{}
}
func (p Parser) ResolveVersion(cnbPath, version string) (string, error) {
	bpTOML, err := os.Open(filepath.Join(cnbPath, "buildpack.toml"))
	if err != nil {
		return "", err
	}

	var buildpackTOML struct {
		Metadata struct {
			DefaultVersions map[string]string `toml:"default-versions"`
			VersionLines    struct {
				Stable string `toml:"stable"`
			} `toml:"version-lines"`
		} `toml:"metadata"`
	}

	_, err = toml.NewDecoder(bpTOML).Decode(&buildpackTOML)
	if err != nil {
		return "", err
	}

	fmt.Printf("%v", buildpackTOML)
	if version == "stable" {
		version = buildpackTOML.Metadata.VersionLines.Stable
	}

	if version == "" {
		version = buildpackTOML.Metadata.DefaultVersions[VSCODIUM]
	}

	return version, nil
}
