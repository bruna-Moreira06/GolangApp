package main

import (
	"encoding/json"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

func yml2json() {

	yfile, err := os.ReadFile("openapi.yml")

	if err != nil {
		log.Fatal(err)
	}

	var data any

	err = yaml.Unmarshal(yfile, &data)

	if err != nil {
		log.Fatal(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")
	enc.Encode(data)
}