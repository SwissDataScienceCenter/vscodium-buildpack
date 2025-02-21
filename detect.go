package vscodiumbuildpack

import (
	"github.com/paketo-buildpacks/packit/v2"
)

type VersionParser interface {
	ResolveVersion(cnbPath, version string) (resultVersion string, err error)
}

type BuildPlanMetadata struct {
	Version       string `toml:"version,omitempty"`
	VersionSource string `toml:"version-source,omitempty"`
	Launch        bool   `toml:"launch"`
}

func Detect(versionParser VersionParser) packit.DetectFunc {
	return func(context packit.DetectContext) (packit.DetectResult, error) {

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
