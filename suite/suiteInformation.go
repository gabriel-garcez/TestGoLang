package suite

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
)

// TestSuite run all test cases
func (tsi *SuiteInformation) TestSuite() {
	logger = tsi.Logs
	logger.Log("", "", "Test Suite Started", INFO, nil, true)
	wg := &sync.WaitGroup{}
	wg.Add(len(tsi.Entities))
	for _, tests := range tsi.Entities {
		if tests.Parent {
			tsi.testEntity(tests, wg)
		} else {
			go tsi.testEntity(tests, wg)
		}
	}
	wg.Wait()
	tsi.Control.TestsFinished <- true
}

func (tsi *SuiteInformation) saveVariable(variable string, value interface{}) {
	if variable != "" && value != nil {
		tsi.SuiteVariables[variable] = value
	}
}

func (tsi *SuiteInformation) loadVariable(isURL bool, data string) string {
	for _, key := range tsi.VariableKeys {
		if value, thisExists := tsi.SuiteVariables[key]; thisExists {
			data = strings.ReplaceAll(data, fmt.Sprintf("{{%s}}", key), value.(string))
		}
	}

	return data
}

func (tsi *SuiteInformation) request(testCase *Routine, entityTitle string) *routineResult {

	reqURL := tsi.loadVariable(true, fmt.Sprintf(
		"%s%s",
		tsi.ServerInformation.FullURL,
		testCase.Endpoint,
	))

	var bodyJSON []byte
	var err error

	if testCase.RequestBody != nil {
		bodyJSON, err = json.Marshal(testCase.RequestBody)
		if err != nil {
			logger.Log(entityTitle, testCase.Description, "Error parsing body", FATAL,
				map[string]interface{}{
					"error_message": err,
				},
				true,
			)
		}
	}

	bodyString := tsi.loadVariable(false, string(bodyJSON))

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !tsi.ServerInformation.CheckTLS},
	}

	client := &http.Client{Transport: transport}

	request, err := http.NewRequest(testCase.Method, reqURL, bytes.NewBuffer([]byte(bodyString)))
	if err != nil {
		logger.Log(
			entityTitle, testCase.Description, "Request error", WARNING,
			map[string]interface{}{
				"error_message": err,
				"url":           reqURL,
				"Method":        testCase.Method,
			},
			false,
		)
	}

	var requestHeader = make(http.Header)

	if tsi.ServerInformation.AuthorizationMode != "" && tsi.loadVariable(false, "{{token}}") != "{{token}}" {
		requestHeader["Authorization"] = []string{fmt.Sprintf("%s {{token}}", tsi.ServerInformation.AuthorizationMode)}
	}

	for header, value := range testCase.Headers {
		requestHeader[header] = strings.Split(value.(string), ", ")
	}

	if tsi.ServerInformation.DefaultContentType != "" {
		requestHeader["Content-Type"] = []string{tsi.ServerInformation.DefaultContentType}
	}

	for headerKey, headerValue := range requestHeader {
		for index, value := range headerValue {
			requestHeader[headerKey][index] = tsi.loadVariable(false, value)
		}
	}

	request.Header = requestHeader

	res, err := client.Do(request)

	return tsi.treatHTTPResponse(err, res, request, testCase, entityTitle)
}

func (tsi *SuiteInformation) treatHTTPResponse(resErr error, response *http.Response, request *http.Request, testCase *Routine, entityTitle string) *routineResult {
	var result = make(map[string]interface{})
	statusCode := 500

	if resErr != nil {
		logger.Log(entityTitle, testCase.Description, "Response with error", ERROR, map[string]interface{}{"error": resErr}, false)
	} else {
		statusCode = response.StatusCode
		if testCase.Creation {
			buffer := new(bytes.Buffer)
			buffer.ReadFrom(response.Body)
			result["data"] = strings.Split(buffer.String(), "\"")[1]
		} else {
			json.NewDecoder(response.Body).Decode(&result)
		}
		response.Body.Close()
	}

	testCase.TestResult = result

	return &routineResult{
		Error:  resErr,
		Code:   statusCode,
		Result: result,
		Expected: struct {
			Code   int      `json:"code"`
			Return []string `json:"return"`
		}{
			Code:   testCase.CodeExpected,
			Return: testCase.ReturnExpected,
		},
		Suite:              tsi,
		CaseVariables:      &testCase.VariablesToSave,
		EntityTitle:        entityTitle,
		RoutineDescription: testCase.Description,
		Headers: map[string]interface{}{
			"request":  request.Header,
			"response": response.Header,
		},
	}
}

func (tsi *SuiteInformation) testEntity(tests *Entities, wg *sync.WaitGroup) {
	defer func() {
		wg.Done()
	}()

	var cover int
	var err []error

	logger.Log(tests.Title, "", "Start entity tests", INFO, nil, true)

	if tests.Loop == 0 {
		tests.Loop = 1
	}

	for run := 0; run < tests.Loop; run++ {
		wgEntity := &sync.WaitGroup{}

		for caseIndex, testCase := range tests.Cases {
			wgEntity.Add(1)
			if testCase.Turbo {
				go tsi.testRoutine(&cover, caseIndex, &err, tests, testCase, wgEntity)
			} else {
				tsi.testRoutine(&cover, caseIndex, &err, tests, testCase, wgEntity)
			}
		}

		wgEntity.Wait()
	}

	coverage := ((cover * 100) / (len(tests.Cases) * tests.Loop))

	if coverage < 100 {
		tsi.Control.TestResult <- TESTWARNING
	} else if coverage == 0 {
		tsi.Control.TestResult <- TESTFAIL
	}

	logger.Log(
		tests.Title,
		"",
		"Finish entity tests",
		INFO,
		map[string]interface{}{
			"cover": map[string]interface{}{
				"passed_tests": cover,
				"test_amount":  len(tests.Cases) * tests.Loop,
				"percent":      coverage,
			},
		},
		true,
	)
}

func (tsi *SuiteInformation) testRoutine(cover *int, caseIndex int, err *[]error, tests *Entities, testCase *Routine, wg *sync.WaitGroup) {
	defer wg.Done()

	if !testCase.endpointUpdated {
		testCase.Endpoint = fmt.Sprintf("%s%s", tests.Path, testCase.Endpoint)
		testCase.endpointUpdated = true
	}
	var passed bool
	var errList []error
	var result routineResult

	result, passed, errList = tsi.request(
		testCase,
		tests.Title,
	).validate()

	if len(errList) > 0 {
		*err = append(*err, errList...)
	}
	if passed {
		*cover++
	}

	logger.Log(
		tests.Title,
		testCase.Description,
		"Test Routine Success",
		NOTICE,
		map[string]interface{}{
			"case_index": caseIndex,
			"pass":       passed,
			"result":     result,
		}, false,
	)
	for _, err := range errList {
		logger.Log(tests.Title, testCase.Description, "Test Routine Fail", ERROR, map[string]interface{}{"case_index": caseIndex, "pass": passed, "error_message": err}, false)
	}
}
