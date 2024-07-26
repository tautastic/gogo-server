package cmd

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os/exec"
	"strings"
)

func StartProcess(kataGoPath string, args []string, inputChan <-chan string, outputChan chan<- string, errorChan chan<- error, readyChan chan<- bool) {
	defer close(errorChan)
	cmd := exec.Command(kataGoPath, args...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		errorChan <- fmt.Errorf("failed to get stdin pipe: %w", err)
		return
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		errorChan <- fmt.Errorf("failed to get stdout pipe: %w", err)
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		errorChan <- fmt.Errorf("failed to get stderr pipe: %w", err)
		return
	}

	if err := cmd.Start(); err != nil {
		errorChan <- fmt.Errorf("failed to start process: %w", err)
		return
	}

	handleOutput := func(reader io.Reader) {
		scanner := bufio.NewScanner(reader)
		for scanner.Scan() {
			if scannedText := strings.TrimSpace(scanner.Text()); len(scannedText) > 0 {
				if strings.Contains(scannedText, "GTP ready") {
					readyChan <- true
					close(readyChan)
					log.Println("sent to readyChan:", scannedText)
					continue
				}
				if strings.HasPrefix(scannedText, "= ") {
					if trimmedText := strings.ReplaceAll(scannedText, "= ", ""); len(trimmedText) > 0 {
						outputChan <- trimmedText
						log.Println("sent to outputChan:", trimmedText)
						continue
					}
				}
				if strings.HasPrefix(scannedText, "?") {
					errorChan <- fmt.Errorf("output error: %s", scannedText)
					continue
				}
			}
		}
		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading output: %w", err)
		}
		close(outputChan)
	}

	go handleOutput(stdout)
	go handleOutput(stderr)

	go func() {
		defer func(stdin io.WriteCloser) {
			err := stdin.Close()
			if err != nil {
				log.Fatal("Could not close stdin", err)
			}
		}(stdin)
		for input := range inputChan {
			log.Println("stdin:", input)
			if _, err := stdin.Write([]byte(input + "\n")); err != nil {
				errorChan <- fmt.Errorf("error writing to stdin: %w", err)
			}
		}
	}()

	if err := cmd.Wait(); err != nil {
		errorChan <- fmt.Errorf("process exited with error: %w", err)
	}
}
