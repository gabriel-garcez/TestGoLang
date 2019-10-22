package suite

import "strings"

// All system constants
const (
	INFO        string     = "information"
	NOTICE      string     = "notice"
	ERROR       string     = "error"
	WARNING     string     = "warning"
	PANIC       string     = "panic"
	FATAL       string     = "fatal"
	TESTSUCCESS testResult = 0
	TESTWARNING testResult = 2
	TESTFAIL    testResult = 1
)

// Logger constants
const (
	InfoColor    = "[\033[1;34m%s\033[0m]%s"
	NoticeColor  = "[\033[1;36m%s\033[0m]%s"
	WarningColor = "[\033[1;33m%s\033[0m]%s"
	ErrorColor   = "[\033[1;31m%s\033[0m]%s"
	DebugColor   = "[\033[0;36m%s\033[0m]%s"
)

var (
	logger *SystemLogs
)

type testResult int8

// SuiteInformation has all test structure
type SuiteInformation struct {
	Control           *suiteChannels         `json:"-"`
	Entities          []*Entities            `json:"tests"`
	Logs              *SystemLogs            `json:"logs"`
	ServerInformation *ServerInformation     `json:"server_info"`
	SuiteVariables    map[string]interface{} `json:"variables_saved"`
	UserInformation   *UserInformation       `json:"user"`
	VariableKeys      []string               `json:"variables"`
	VerboseMode       bool                   `json:"-"`
}

type suiteChannels struct {
	TestResult    chan testResult
	TestsFinished chan bool
}

// ServerInformation informations about target server
type ServerInformation struct {
	AccessPort         string `json:"access_port"`
	APIVersion         string `json:"api_version"`
	AuthorizationMode  string `json:"authorization_mode"`
	CheckTLS           bool   `json:"check_tls"`
	DefaultContentType string `json:"default_content_type"`
	DNSAddress         string `json:"dns_address"`
	FullURL            string `json:"-"`
	Name               string `json:"name"`
	Scheme             string `json:"scheme"`
	TestPath           string `json:"test_path"`
}

// UserInformation informations about test user
type UserInformation struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Entities test entities with all test cases
type Entities struct {
	Cases  []*Routine `json:"cases"`
	Loop   int        `json:"loop_times"`
	Parent bool       `json:"parent"`
	Path   string     `json:"path"`
	Title  string     `json:"title"`
}

// Routine test routine with informations about the task
type Routine struct {
	endpointUpdated bool
	Creation        bool                   `json:"creation"`
	CodeExpected    int                    `json:"code_expected"`
	Description     string                 `json:"description"`
	Endpoint        string                 `json:"endpoint"`
	Headers         map[string]interface{} `json:"headers"`
	Method          string                 `json:"method"`
	RequestBody     map[string]interface{} `json:"body"`
	ReturnExpected  []string               `json:"return_expected"`
	TestResult      map[string]interface{} `json:"test_result"`
	Turbo           bool                   `json:"turbo_test"`
	VariablesToSave []string               `json:"variables_to_save"`
}

type routineResult struct {
	Code     int                    `json:"code"`
	Result   map[string]interface{} `json:"result"`
	Error    error                  `json:"error"`
	Expected struct {
		Code   int      `json:"code"`
		Return []string `json:"return"`
	} `json:"expected"`
	Suite              *SuiteInformation      `json:"suite"`
	CaseVariables      *[]string              `json:"case_variables"`
	EntityTitle        string                 `json:"entity"`
	RoutineDescription string                 `json:"routine"`
	Headers            map[string]interface{} `json:"headers"`
}

type SystemLogs struct {
	LogHistory  []*LogDetail `json:"history"`
	logHub      chan *LogDetail
	verboseMode bool
}

type LogDetail struct {
	Date         int64                  `json:"date"`
	Details      map[string]interface{} `json:"details"`
	EntityName   string                 `json:"entity"`
	Level        string                 `json:"level"`
	Message      string                 `json:"message"`
	RoutineTitle string                 `json:"routine"`
}

func translateVariableTransfer(rawVariable string) (string, string) {
	if len(rawVariable) > 2 {
		varList := strings.Split(rawVariable, "->")
		if len(varList) > 1 {
			return varList[0], varList[1]
		}
		return rawVariable, ""
	}
	return "", ""
}

// NewSuiteControl return a new suite channels
func NewSuiteControl() *suiteChannels {
	return &suiteChannels{
		TestsFinished: make(chan bool),
		TestResult:    make(chan testResult),
	}
}
