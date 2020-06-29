// Copyright 2018. Akamai Technologies, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package edgegrid

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"os"
	"strings"

	logstd "log"

	log "github.com/sirupsen/logrus"
)

var logBuffer *bufio.Writer
var LogFile *os.File
var EdgegridLog *log.Logger
var ctx context.Context

type SomeContextKey string

// LogCorrelationID ID for header and footer of log file outputs
//var LogCorrelationID *string

func SetupLogging() {
	keyA := SomeContextKey("LogCorrelationID")
	ctx = context.Background()
	ctx = context.WithValue(ctx, keyA, "foo")

	if EdgegridLog != nil {
		return // already configured
	}

	EdgegridLog = log.New()
	EdgegridLog.SetFormatter(&log.TextFormatter{
		DisableLevelTruncation:    true,
		EnvironmentOverrideColors: true,
	})
	// Log file destination specified? If not, use default stdout
	if logFileName := os.Getenv("AKAMAI_LOG_FILE"); logFileName != "" {
		// If the file doesn't exist, create it, or append to the file
		LogFile, err := os.OpenFile(logFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}
		EdgegridLog.SetOutput(LogFile)
	}

	EdgegridLog.SetLevel(log.PanicLevel)
	if logLevel := os.Getenv("AKAMAI_LOG"); logLevel != "" {
		level, err := log.ParseLevel(logLevel)
		if err == nil {
			EdgegridLog.SetLevel(level)
		} else {
			log.Warningln("[WARN] Unknown AKAMAI_LOG value. Allowed values: panic, fatal, error, warn, info, debug, trace")

		}
	}

	defer LogFile.Close()
}

func LogMultiline(f func(args ...interface{}), args ...string) {
	for _, str := range args {
		for _, str := range strings.Split(strings.Trim(str, "\n"), "\n") {
			f(str)
		}
	}
}

func LogMultilineln(f func(args ...interface{}), args ...string) {
	LogMultiline(f, args...)
}

func LogMultilinef(f func(formatter string, args ...interface{}), formatter string, args ...interface{}) {
	str := fmt.Sprintf(formatter, args...)
	for _, str := range strings.Split(strings.Trim(str, "\n"), "\n") {
		f(str)
	}
}

/*func PrintLogHeader() {
	if LogCorrelationID != nil {
		strLogCorrelationIDValue := *LogCorrelationID
		LogMultiline(EdgegridLog.Traceln, "START CORRELATION ID "+strLogCorrelationIDValue)

	}
}

func PrintLogFooter() {
	if LogCorrelationID != nil {
		strLogCorrelationIDValue := *LogCorrelationID
		LogMultiline(EdgegridLog.Traceln, "END CORRELATION ID "+strLogCorrelationIDValue)
		LogCorrelationID = nil
	}
}
*/

// Utility func to print http req
func PrintHttpRequest(req *http.Request, body bool) {

	if req == nil {
		return
	}
	b, err := httputil.DumpRequestOut(req, body)
	if err == nil {
		LogMultiline(EdgegridLog.Traceln, string(b))
		logstd.Printf("[DEBUG]  REQUEST  %s\n", string(b))
	}
}

func PrintHttpRequestCorrelation(req *http.Request, body bool, correlationid string) {

	if req == nil {
		return
	}
	b, err := httputil.DumpRequestOut(req, body)
	if err == nil {
		LogMultiline(EdgegridLog.Traceln, string(b))
		logstd.Printf("[DEBUG]%v REQUEST  %s\n", correlationid, string(b))
	}
}

// Utility func to print http response
func PrintHttpResponse(res *http.Response, body bool) {

	if res == nil {
		return
	}
	b, err := httputil.DumpResponse(res, body)
	if err == nil {
		//PrintLogHeader()
		LogMultiline(EdgegridLog.Traceln, string(b))
		logstd.Printf("[DEBUG] RESPONSE %s\n", string(b))
	}
}

func PrintHttpResponseCorrelation(res *http.Response, body bool, correlationid string) {

	if res == nil {
		return
	}
	b, err := httputil.DumpResponse(res, body)
	if err == nil {
		LogMultiline(EdgegridLog.Traceln, string(b))
		logstd.Printf("[DEBUG]%v RESPONSE %s\n", correlationid, string(b))
	}
}

func PrintfCorrelation(level string, correlationid string, msg string) {
	logstd.Printf("%s%v  %s\n", level, correlationid, msg)
}
