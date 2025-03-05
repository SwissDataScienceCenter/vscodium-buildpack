package internal

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func Run(mount_dir string, extensions []string) error {

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)

	codium_path := filepath.Join(exPath, "..", "bin", "codium-server")
	vscode_dir := filepath.Join(mount_dir, ".vscode")
	extension_dir := filepath.Join(vscode_dir, "extensions")
	err = os.MkdirAll(extension_dir, 0755)
	if err != nil {
		log.Fatalf("Couldn't create directory %s: %s", extension_dir, err)
	}

	for _, extension := range extensions {
		log.Printf("Installing VSCode extension: %s", extension)
		cmd := exec.Command(
			codium_path,
			"--extensions-dir",
			extension_dir,
			"--server-data-dir",
			vscode_dir,
			"--install-extension",
			extension,
		)
		out, err := cmd.Output()
		if err != nil {
			log.Fatalf("Couldn't install extension %s(%s): %s", extension, err, out)
		}
	}
	return nil
}
