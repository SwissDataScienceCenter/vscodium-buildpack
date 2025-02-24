package vscodiumbuildpack_test

import (
	"errors"
	"testing"

	vscodium "github.com/SwissDataScienceCenter/vscodium-buildpack"
	"github.com/SwissDataScienceCenter/vscodium-buildpack/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDetect(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		workingDir string

		versionParser *fakes.VersionParser
		detect        packit.DetectFunc
	)

	it.Before(func() {
		workingDir = t.TempDir()

		versionParser = &fakes.VersionParser{}
		versionParser.ResolveVersionCall.Returns.ResultVersion = "1.96.*"

		detect = vscodium.Detect(versionParser)
	})

	it("returns a plan that provides vscodium", func() {
		result, err := detect(packit.DetectContext{
			WorkingDir: workingDir,
		})
		Expect(err).NotTo(HaveOccurred())
		Expect(result.Plan).To(Equal(packit.BuildPlan{
			Provides: []packit.BuildPlanProvision{
				{Name: "vscodium"},
			},
			Requires: []packit.BuildPlanRequirement{
				{
					Name: "vscodium",
					Metadata: vscodium.BuildPlanMetadata{
						Version:       "1.96.*",
						VersionSource: "buildpack.toml",
						Launch:        true,
					},
				},
			},
		}))
	})

	context("failure cases", func() {

		context("version parsing fails", func() {
			it.Before(func() {
				versionParser.ResolveVersionCall.Returns.Err = errors.New("parsing version failed")
			})

			it("returns an error", func() {
				_, err := detect(packit.DetectContext{
					WorkingDir: workingDir,
				})

				Expect(err).To(MatchError(ContainSubstring("parsing version failed")))
			})
		})
	})
}
