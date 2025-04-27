package vscodiumbuildpack

import (
	"os"
	"strings"

	"github.com/paketo-buildpacks/packit/v2"
)

type VersionParser interface {
	ResolveVersion(cnbPath, version string) (resultVersion string, err error)
}

const RenkuFrontendEnvKey = "RENKU_FRONTENDS"
const RenkuFrontendEnvValue = "vscodium"

type BuildPlanMetadata struct {
	Version       string `toml:"version,omitempty"`
	VersionSource string `toml:"version-source,omitempty"`
	Launch        bool   `toml:"launch"`
}

func Detect(versionParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

		val := os.Getenv(RenkuFrontendEnvKey)
		if !strings.Contains(strings.ToLower(val), RenkuFrontendEnvValue) {
			return packit.DetectResult{}, packit.Fail.WithMessage("coulld not find the %s environment variable or its value does not contain %s", RenkuFrontendEnvKey, RenkuFrontendEnvValue)
		}
		plan := packit.DetectResult{
			Plan: packit.BuildPlan{
				Provides: []packit.BuildPlanProvision{
					{Name: VSCODIUM},
				},
			},
		}

		var requirements []packit.BuildPlanRequirement
		version, err := versionParser.ResolveVersion(context.CNBPath, "")
		if err != nil {
			return packit.DetectResult{}, err
		}
		requirements = append(requirements, packit.BuildPlanRequirement{
			Name: VSCODIUM,
			Metadata: BuildPlanMetadata{
				Version:       version,
				VersionSource: "buildpack.toml",
				Launch:        true,
			},
		})
		plan.Plan.Requires = requirements
		return plan, nil

	}
}
