package internal

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Run(environment map[string]string) error {

	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	exPath := filepath.Dir(ex)

	codium_path := filepath.Join(exPath, "..", "bin", "codium-server")
	workdir, ok := environment["ROOT_DIR"]
	if !ok {
		log.Fatal("ROOT_DIR environment variable not set")
	}

	extensions := []string{"ms-python.python", "ms-toolsai.jupyter"}
	if val, ok := environment["VSCODIUM_EXTENSIONS"]; ok {
		extensions = strings.Split(val, " ")
	}
	vscode_dir := filepath.Join(workdir, ".vscode")
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
