package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

// TODO: Maybe receiving configuration from a file or main args is better.
const (
	_threads     = 2
	_command     = "cypress"
	_configFile  = "docker.cypress.config.ts"
	_specBaseDir = "cypress/e2e"
)

var commonArgs = []string{
	"run",
	"--browser",
	"chrome",
	"--headless",
	"--config-file",
	_configFile,
	"--spec",
}

func main() {
	files := getSpecFiles(_specBaseDir)
	args := getSpecArgs(_threads, files)

	for idx, arg := range args {
		log.Println("THREAD ", idx, ": ", arg)
	}

	runCypressParallel(args)
}

// runCypressParallel executes n number (length of args) of cypress test instances in go routine.
func runCypressParallel(args []string) {
	var wg sync.WaitGroup

	for _, arg := range args {
		wg.Add(1)
		go func(arg string) {
			defer wg.Done()
			trail := append(commonArgs, "'"+arg+"'")
			execute(_command, trail...)
		}(arg)
	}
	wg.Wait()
	// Delaying this message is necessary to display all output buffer from the go routines.
	time.AfterFunc(5*time.Second, func() {
		execute("echo", "Cypress Test Completed")
	})
	log.Println("Cypress Test Completed")
}

// getSpecFiles retrieves all cypress spec files recursively from baseDir. The spec file must contain *.spec.* pattern
func getSpecFiles(baseDir string) []string {
	var result = make([]string, 0)
	files, err := os.ReadDir(baseDir)
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		if file.IsDir() {
			result = append(result, getSpecFiles(baseDir+"/"+file.Name())...)
		} else if strings.Contains(file.Name(), ".spec.") {
			result = append(result, baseDir+"/"+file.Name())
		}
	}
	return result
}

// getSpecArgs spreads specFiles into number of nThreads for trailing arguments for cypress instance execution.
// Spreading algorithm is linear purposely for load balancing reason, since certain folders like integration test
// contains tests that are more complex and need to separate them into each bin.
func getSpecArgs(nThreads int, specFiles []string) []string {
	var args = make([][]string, nThreads)
	for i, file := range specFiles {
		idx := i % nThreads
		args[idx] = append(args[idx], file)
	}

	var results = make([]string, nThreads)
	for i, arg := range args {
		results[i] = strings.Join(arg, ",")
	}
	return results
}

// execute executes command and displays its output on shell.
func execute(cmd string, args ...string) {
	out, err := exec.Command(cmd, args...).Output()

	if err != nil {
		fmt.Printf("%s", err)
	}

	output := string(out[:])
	fmt.Println(output)
}
