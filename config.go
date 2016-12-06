package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

const (
	configFileName = "rfb.json"
)

// Settings is a main struct for settings
type Settings struct {
	APIKey        string            `json:"api-key"`
	Addr          string            `json:"addr"`
	Couchbase     CouchbaseSettings `json:"couchbase"`
	StaticDirPath string            `json:"static-dir-path"`
}

// CouchbaseSettings is a sub truct for couchbase settings
type CouchbaseSettings struct {
	Cluster string `json:"cluster"`
	Bucket  string `json:"bucket"`
	Secret  string `json:"secret"`
}

// LoadConfig function load a config file
func LoadConfig() {
	settings.APIKey = ""
	settings.Addr = ":8088"
	settings.Couchbase.Cluster = "couchbase://couchbase"
	settings.Couchbase.Bucket = "default"
	settings.Couchbase.Secret = ""

	f, err := os.Open(configFileName)
	if err != nil {
		// not found config file or config file don't be a read
		return
	}
	defer f.Close()

	data, err := ioutil.ReadAll(f)
	if err != nil {
		return
	}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		log.Printf("Config file unmarshal failled! Use default settings.")
	}
}

// SaveConfig function save a config file
func SaveConfig() {
	f, err := os.Create(configFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	data, err := json.MarshalIndent(&settings, "", "    ")
	if err != nil {
		log.Fatal(err)
	}
	_, err = f.Write(data)
	if err != nil {
		log.Fatal(err)
	}
}
