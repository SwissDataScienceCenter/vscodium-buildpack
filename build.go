package vscodiumbuildpack

import (
	"os"
	"path/filepath"
	"time"

	"github.com/paketo-buildpacks/packit/v2"
	"github.com/paketo-buildpacks/packit/v2/chronos"
	"github.com/paketo-buildpacks/packit/v2/draft"
	"github.com/paketo-buildpacks/packit/v2/postal"
	"github.com/paketo-buildpacks/packit/v2/sbom"
	"github.com/paketo-buildpacks/packit/v2/scribe"
)

type DependencyService interface {
	Resolve(path, name, version, stack string) (postal.Dependency, error)
	Deliver(dependency postal.Dependency, cnbPath, layerPath, platformPath string) error
	GenerateBillOfMaterials(dependencies ...postal.Dependency) []packit.BOMEntry
}

//go:generate faux --interface Calculator --output fakes/calculator.go
type Calculator interface {
	Sum(paths ...string) (string, error)
}

//go:generate faux --interface SBOMGenerator --output fakes/sbom_generator.go
type SBOMGenerator interface {
	GenerateFromDependency(dependency postal.Dependency, dir string) (sbom.SBOM, error)
}

func Build(
	dependencyService DependencyService,
	calculator Calculator,
	sbomGenerator SBOMGenerator,
	logger scribe.Emitter,
	clock chronos.Clock,
) packit.BuildFunc {
	return func(context packit.BuildContext) (packit.BuildResult, error) {
		logger.Title("%s %s", context.BuildpackInfo.Name, context.BuildpackInfo.Version)

		planner := draft.NewPlanner()

		logger.Process("Resolving VSCodium version")
		entry, sortedEntries := planner.Resolve("vscodium", context.Plan.Entries, []interface{}{
			"BP_VSCODIUM_VERSION",
			"buildpack.toml",
		})
		logger.Candidates(sortedEntries)

		entryVersion, _ := entry.Metadata["version"].(string)
		dependency, err := dependencyService.Resolve(filepath.Join(context.CNBPath, "buildpack.toml"), entry.Name, entryVersion, context.Stack)
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.SelectedDependency(entry, dependency, clock.Now())

		layer, err := context.Layers.Get(VSCODIUM)
		if err != nil {
			return packit.BuildResult{}, err
		}

		bom := dependencyService.GenerateBillOfMaterials(dependency)
		launch, build := planner.MergeLayerTypes(VSCODIUM, context.Plan.Entries)

		var buildMetadata packit.BuildMetadata
		if build {
			buildMetadata.BOM = bom
		}

		var launchMetadata packit.LaunchMetadata
		if launch {
			command := "codium-server"
			args := []string{
				"--server-base-path", "${RENKU_BASE_URL_PATH}",
				"--host", "${RENKU_SESSION_IP}",
				"--port", "${RENKU_SESSION_PORT}",
				"--extensions-dir", "${RENKU_MOUNT_DIR}/.vscode/extensions",
				"--server-data-dir", "${RENKU_MOUNT_DIR}/.vscode",
				"--without-connection-token",
				"--accept-server-license-terms",
				"--telemetry-level", "off",
				"--default-folder", "${RENKU_WORKING_DIR}",
			}
			launchMetadata.Processes = []packit.Process{
				{
					Type:    "web",
					Command: command,
					Args:    args,
					Default: true,
					Direct:  false,
				},
			}
			launchMetadata.BOM = bom

		}

		logger.Process("Executing build process")

		// TODO: check layer caching so we don't reinstall if already installed
		layer, err = layer.Reset()
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer.Launch, layer.Build = launch, build

		logger.Subprocess("Installing VSCodium %s", dependency.Version)
		duration, err := clock.Measure(func() error {
			return dependencyService.Deliver(dependency, context.CNBPath, layer.Path, context.Platform.Path)
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		layer.Metadata = map[string]interface{}{
			DepKey: dependency.Checksum,
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		layer.SharedEnv.Append("PATH", filepath.Join(layer.Path, "bin"), string(os.PathListSeparator))

		layer.ExecD = []string{filepath.Join(context.CNBPath, "bin", "install-extensions")}
		layer.LaunchEnv.Default("RENKU_SESSION_IP", "0.0.0.0")
		layer.LaunchEnv.Default("RENKU_SESSION_PORT", "8000")
		layer.LaunchEnv.Default("RENKU_MOUNT_DIR", context.WorkingDir)
		layer.LaunchEnv.Default("RENKU_WORKING_DIR", context.WorkingDir)
		layer.LaunchEnv.Default("RENKU_BASE_URL_PATH", "/")
		layer.LaunchEnv.Default("VSCODIUM_EXTENSIONS", "ms-python.python ms-toolsai.jupyter")

		logger.EnvironmentVariables(layer)

		logger.LaunchProcesses(launchMetadata.Processes)

		logger.GeneratingSBOM(layer.Path)
		var sbomContent sbom.SBOM
		duration, err = clock.Measure(func() error {
			sbomContent, err = sbomGenerator.GenerateFromDependency(dependency, layer.Path)
			return err
		})
		if err != nil {
			return packit.BuildResult{}, err
		}

		logger.Action("Completed in %s", duration.Round(time.Millisecond))
		logger.Break()

		logger.FormattingSBOM(context.BuildpackInfo.SBOMFormats...)
		layer.SBOM, err = sbomContent.InFormats(context.BuildpackInfo.SBOMFormats...)
		if err != nil {
			return packit.BuildResult{}, err
		}

		return packit.BuildResult{
			Layers: []packit.Layer{layer},
			Build:  buildMetadata,
			Launch: launchMetadata,
		}, nil
	}
}
