package main

import (
	"bookish-memory/suite"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	persistTest    bool
	verboseMode    bool
	logger         *suite.SystemLogs
	exitCode       int8   = 0
	outputFilePath string = "."
)

func main() {
	defer func() {
		os.Exit(int(exitCode))
	}()

	var err error
	var testSuite *suite.SuiteInformation
	var testFilePath string
	var testFileRemote io.ReadCloser

	for index, arg := range os.Args {
		switch arg {
		case "-f", "--file", "file":
			testFilePath = os.Args[index+1]
		case "-o", "--output", "output":
			outputFilePath = os.Args[index+1]
		case "-h", "--help", "help":
			help()
		case "-v", "--version", "version":
			version()
		case "-p", "--persist", "persist":
			persistTest = true
		case "-vb", "--verbose", "verbose":
			verboseMode = true
		case "-r", "--remote", "remote":
			response, err := http.Get(os.Args[index+1])
			if err != nil {
				log.Fatalf("[FATAL] ERROR GETTING FILE FROM URL: %s", os.Args[index+1])
			}
			testFileRemote = response.Body
		}
	}

	if testFilePath != "" {
		testSuite, err = loadFile(testFilePath, nil)
	}

	if testFileRemote != nil {
		testSuite, err = loadFile("", testFileRemote)
	}

	if err != nil {
		log.Fatalf("[FATAL] ERROR READING FILE!!! [%v]", err)
	}

	if testSuite == nil {
		exitCode = 1
		return
	}

	testSuite.VerboseMode = verboseMode

	logger = suite.NewSystemLogs(testSuite.VerboseMode)

	testSuite.Logs = logger

	startTests(testSuite)
}

func startTests(testSuite *suite.SuiteInformation) {
	logMain("Start CCAP Test Suite", suite.INFO, nil)

	err := pingServer(testSuite)
	if err != nil {
		logMain("ERROR COMUNICATING WITH SERVER!!!", suite.FATAL, map[string]interface{}{"error": err})
	}

	testSuite.VerboseMode = verboseMode

	startTest := time.Now()
	go testSuite.TestSuite()
	func() {
		for {
			select {
			case res := <-testSuite.Control.TestResult:
				if int8(res) > exitCode {
					exitCode = int8(res)
				}
			case <-testSuite.Control.TestsFinished:
				return
			}
		}
	}()
	timeSinceStart := time.Since(startTest)

	logMain(fmt.Sprintf("All tests finished in %v", timeSinceStart), suite.INFO, map[string]interface{}{"duration_time": timeSinceStart})

	if persistTest {
		logJSON, err := json.Marshal(testSuite)
		if err != nil {
			exitCode = 2
			logMain("Parsing test suite", suite.ERROR, map[string]interface{}{"error_message": err})
		} else {
			if err := ioutil.WriteFile(fmt.Sprintf("%s/bookish-memory_%v.json", outputFilePath, time.Now().Unix()), logJSON, 0644); err != nil {
				dir, _ := os.Getwd()
				logMain(fmt.Sprintf("Saving Test State in: %s", dir), suite.ERROR, map[string]interface{}{"error_message": err})
			}
		}
	}
}

func loadFile(path string, loadedJSON io.ReadCloser) (*suite.SuiteInformation, error) {
	var file *os.File
	var decoder *json.Decoder
	var err error

	newTestSuite := &suite.SuiteInformation{
		SuiteVariables: make(map[string]interface{}),
		Control:        suite.NewSuiteControl(),
	}

	if path != "" {
		file, err = os.Open(path)
		if err != nil {
			return nil, err
		}
		decoder = json.NewDecoder(file)
	}

	if loadedJSON != nil {
		decoder = json.NewDecoder(loadedJSON)
	}

	err = decoder.Decode(&newTestSuite)
	if err != nil {
		return nil, err
	}

	if newTestSuite.ServerInformation == nil || newTestSuite.Entities == nil {
		return nil, fmt.Errorf("Server information and test entities are required")
	}

	return newTestSuite, nil
}

func pingServer(server *suite.SuiteInformation) error {
	reqURL := server.ServerInformation.FullURL
	if reqURL == "" {
		reqURL = fmt.Sprintf(
			"%s%s%s%s",
			server.ServerInformation.Scheme,
			server.ServerInformation.DNSAddress,
			server.ServerInformation.AccessPort,
			server.ServerInformation.APIVersion,
		)
		server.ServerInformation.FullURL = reqURL
	}
	logMain("Starting Ping", suite.INFO, map[string]interface{}{"server_address": reqURL})

	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	if _, err := http.Get(fmt.Sprintf("%s%s", reqURL, server.ServerInformation.TestPath)); err != nil {
		return err
	}

	logMain("Successful Ping", suite.INFO, map[string]interface{}{"server_address": reqURL})

	return nil
}

func help() {
	fmt.Println(`
	 _                 _    _     _                                                     
	| |__   ___   ___ | | _(_)___| |__        _ __ ___   ___ _ __ ___   ___  _ __ _   _ 
	| '_ \ / _ \ / _ \| |/ / / __| '_ \ _____| '_ ' _ \ / _ \ '_ ' _ \ / _ \| '__| | | |
	| |_) | (_) | (_) |   <| \__ \ | | |_____| | | | | |  __/ | | | | | (_) | |  | |_| |
	|_.__/ \___/ \___/|_|\_\_|___/_| |_|     |_| |_| |_|\___|_| |_| |_|\___/|_|   \__, |
	..............................................................................|___/ 

	Command: bookish-memory <OPTION> [...]

	Options:

		-f, --file, file			-> Path to test file
		-o, --output, output			-> Path to save result file
		-h, --help, help			-> Show this help message
		-v, --version, version			-> Show tool version
		-p, --persist, persist			-> Save result output in a .json file
		-vb, --verbose, verbose			-> Verbose mode, output in terminal all possible logs (decreases performance)
		-r, --remote, remote			-> Get test file from url
	
	Author: Rafael Gomides <rafael.gomides.trd@c6bank.com>

	Version: 0.11.1`)
}

func logMain(msg, level string, details map[string]interface{}) {
	logger.Log("", "", msg, level, details, true)
}

func version() {
	filePath, _ := os.Executable()
	fmt.Printf("%s ~> bookish-memory v0.11.0 17/10/2019", filePath)
}
