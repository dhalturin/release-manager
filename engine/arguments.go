package engine

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/dhalturin/release-manager/data"
)

func argumentsLoad() {
	ex, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}

	path, _ := filepath.Abs(filepath.Dir(ex) + "/../../../etc/" + data.Project)

	flag.BoolVar(&data.Arg.Version, "version", false, "Show version")
	flag.StringVar(&data.Arg.ConfigPath, "config-path", path, "Set the configuration file path")
	flag.StringVar(&data.Arg.ConfigFile, "config-file", "config.json", "Set the configuration file name")

	flag.Parse()

	if data.Arg.Version {
		fmt.Printf("Version: %s\n", data.Version)
		os.Exit(0)
	}
}

func init() {
	argumentsLoad()
}
