package engine

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/dhalturin/release-manager/data"
)

func configLoad() {
	configFile := data.Arg.ConfigPath + "/" + data.Arg.ConfigFile

	if _, err := os.Stat(configFile); err == nil {
		configFile, err := ioutil.ReadFile(configFile)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(configFile, &data.Config)
	} else {
		log.Fatal(err)
	}
}

func init() {
	configLoad()
}
