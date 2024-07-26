package cmd

import (
	"log"
	"os"
)

func LoadEnvVars(envNames ...string) []string {
	var envValues []string
	for _, name := range envNames {
		envValues = append(envValues, LoadEnvVar(name))
	}

	return envValues
}

func LoadEnvVar(envName string) string {
	val, ok := os.LookupEnv(envName)
	if !ok {
		log.Fatalf("%s environment variable not set", envName)
	}

	return val
}
