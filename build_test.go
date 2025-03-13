package vscodiumbuildpack_test

import (
	"bytes"
	"errors"
	"path/filepath"
	"testing"

	vscodiumbuildpack "github.com/SwissDataScienceCenter/vscodium-buildpack"
	"github.com/SwissDataScienceCenter/vscodium-buildpack/fakes"
	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	//nolint Ignore SA1019, informed usage of deprecated package
	"github.com/paketo-buildpacks/packit/v2/paketosbom"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"

	. "github.com/onsi/gomega"
	"github.com/sclevine/spec"
)

func testBuild(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		layersDir    string
		cnbPath      string
		workspaceDir string

		dependencyService *fakes.DependencyService
		calculator        *fakes.ChecksumCalculator
		sbomGenerator     *fakes.SBOMGenerator

		buffer *bytes.Buffer

		buildContext packit.BuildContext
		build        packit.BuildFunc
	)

	it.Before(func() {
		layersDir = t.TempDir()
		cnbPath = t.TempDir()
		workspaceDir = t.TempDir()

		buffer = bytes.NewBuffer(nil)

		dependencyService = &fakes.DependencyService{}
		dependencyService.ResolveCall.Returns.Dependency = postal.Dependency{
			ID:             "vscodium",
			Checksum:       "sha256:some-sha",
			Source:         "some-source",
			SourceChecksum: "sha256:some-source-sha",
			Stacks:         []string{"some-stack"},
			URI:            "some-uri",
			Version:        "1.96.4",
		}
		dependencyService.GenerateBillOfMaterialsCall.Returns.BOMEntrySlice = []packit.BOMEntry{
			{
				Name: "vscodium",
				Metadata: paketosbom.BOMMetadata{
					Version: "vscodium-dependency-version",
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "vscodium-dependency-sha",
					},
					URI: "vscodium-dependency-uri",
				},
			},
		}

		calculator = &fakes.ChecksumCalculator{}
		calculator.SumCall.Returns.String = "some-bin-sha"

		sbomGenerator = &fakes.SBOMGenerator{}
		sbomGenerator.GenerateFromDependencyCall.Returns.SBOM = sbom.SBOM{}

		buildContext = packit.BuildContext{
			BuildpackInfo: packit.BuildpackInfo{
				Name:        "VSCodium Buildpack",
				SBOMFormats: []string{sbom.CycloneDXFormat, sbom.SPDXFormat},
				Version:     "1.2.3",
			},
			CNBPath:    cnbPath,
			WorkingDir: workspaceDir,
			Stack:      "some-stack",
			Platform:   packit.Platform{Path: "platform"},
			Plan: packit.BuildpackPlan{
				Entries: []packit.BuildpackPlanEntry{
					{
						Name: "vscodium",
						Metadata: map[string]interface{}{
							"version-source": "BP_VSCODIUM_VERSION",
							"version":        "1.96.*",
							"launch":         true,
						},
					},
				},
			},
			Layers: packit.Layers{Path: layersDir},
		}

		build = vscodiumbuildpack.Build(
			dependencyService,
			calculator,
			sbomGenerator,
			scribe.NewEmitter(buffer),
			chronos.DefaultClock,
		)

	})

	it("does a build", func() {
		result, err := build(buildContext)
		Expect(err).NotTo(HaveOccurred())

		Expect(len(result.Layers)).To(Equal(1))

		layer := result.Layers[0]
		Expect(layer.Name).To(Equal("vscodium"))
		Expect(layer.Path).To(Equal(filepath.Join(layersDir, "vscodium")))
		Expect(layer.Build).To(BeFalse())
		Expect(layer.Launch).To(BeTrue())
		Expect(layer.SharedEnv).To(Equal(packit.Environment{
			"PATH.append": filepath.Join(layersDir, "vscodium", "bin"),
			"PATH.delim":  ":",
		}))
		Expect(layer.LaunchEnv).To(Equal(packit.Environment{
			"RENKU_SESSION_IP.default":    "0.0.0.0",
			"RENKU_SESSION_PORT.default":  "8000",
			"RENKU_MOUNT_DIR.default":     workspaceDir,
			"RENKU_WORKING_DIR.default":   workspaceDir,
			"RENKU_BASE_URL_PATH.default": "/",
			"VSCODIUM_EXTENSIONS.default": "ms-python.python ms-toolsai.jupyter",
		}))
		Expect(layer.Metadata).To(Equal(map[string]interface{}{
			vscodiumbuildpack.DepKey: "sha256:some-sha",
		}))
		Expect(layer.ExecD).To(Equal([]string{filepath.Join(cnbPath, "bin", "install-extensions")}))

		Expect(result.Launch.BOM).To(Equal([]packit.BOMEntry{
			{
				Name: "vscodium",
				Metadata: paketosbom.BOMMetadata{
					Version: "vscodium-dependency-version",
					Checksum: paketosbom.BOMChecksum{
						Algorithm: paketosbom.SHA256,
						Hash:      "vscodium-dependency-sha",
					},
					URI: "vscodium-dependency-uri",
				},
			},
		}))

		Expect(result.Launch.Processes).To(Equal([]packit.Process{
			{
				Type:    "web",
				Command: "codium-server",
				Args: []string{
					"--server-base-path", "${RENKU_BASE_URL_PATH}/",
					"--host", "${RENKU_SESSION_IP}",
					"--port", "${RENKU_SESSION_PORT}",
					"--extensions-dir", "${RENKU_MOUNT_DIR}/.vscode/extensions",
					"--server-data-dir", "${RENKU_MOUNT_DIR}/.vscode",
					"--without-connection-token",
					"--accept-server-license-terms",
					"--telemetry-level", "off",
					"--default-folder", "${RENKU_WORKING_DIR}",
				},
				Direct:  false,
				Default: true,
			},
		}))

		Expect(filepath.Join(layersDir, "vscodium")).To(BeADirectory())

		Expect(dependencyService.ResolveCall.Receives.Path).To(Equal(filepath.Join(cnbPath, "buildpack.toml")))
		Expect(dependencyService.ResolveCall.Receives.Name).To(Equal("vscodium"))
		Expect(dependencyService.ResolveCall.Receives.Version).To(Equal("1.96.*"))
		Expect(dependencyService.ResolveCall.Receives.Stack).To(Equal("some-stack"))
	})

	context("failure cases", func() {
		context("when the dependency cannot be resolved", func() {
			it.Before(func() {
				dependencyService.ResolveCall.Returns.Error = errors.New("failed to resolve dependency")
			})
			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to resolve dependency"))
			})
		})
		context("when the dependency cannot be installed", func() {
			it.Before(func() {
				dependencyService.DeliverCall.Returns.Error = errors.New("failed to deliver dependency")
			})

			it("returns an error", func() {
				_, err := build(buildContext)
				Expect(err).To(MatchError("failed to deliver dependency"))
			})
		})
	})
}
