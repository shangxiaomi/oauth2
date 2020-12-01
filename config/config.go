package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
)

var cfg App

func Setup() {
	file, err := ioutil.ReadFile("app.yaml")
	if err != nil {
		log.Fatal("error: %v", err)
	}
	err = yaml.Unmarshal(file, &cfg)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
}
