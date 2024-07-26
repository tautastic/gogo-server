package main

import (
	"github.com/tautastic/gogo-server/api"
	"github.com/tautastic/gogo-server/cmd"
)

func main() {
	envVars := cmd.LoadEnvVars("AUTHORIZATION_TOKEN", "KATAGO_PATH")
	inputChan, outputChan, errorChan, readyChan := make(chan string), make(chan string, 100), make(chan error), make(chan bool, 1)
	go cmd.StartProcess(envVars[1], []string{"gtp"}, inputChan, outputChan, errorChan, readyChan)
	server := api.NewServer(envVars[0], inputChan, outputChan, errorChan, readyChan)
	server.Start()
}
