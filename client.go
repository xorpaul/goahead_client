package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	h "github.com/xorpaul/gohelper"
)

var (
	debug     bool
	verbose   bool
	info      bool
	quiet     bool
	buildtime string
	config    configSettings
	client    *http.Client
)

type request struct {
	Fqdn      string `json:"fqdn"`
	Uptime    string `json:"uptime"`
	RequestID string `json:"request_id,omitempty"`
}

type response struct {
	Error          string    `json:"error"`
	Timestamp      time.Time `json:"timestamp"`
	Goahead        bool      `json:"go_ahead"`
	UnknownHost    bool      `json:"unknown_host"`
	AskagainIn     string    `json:"ask_again_in"`
	RequestID      string    `json:"request_id"`
	FoundCluster   string    `json:"found_cluster"`
	RequestingFqdn string    `json:"requesting_fqdn"`
	Message        string    `json:"message"`
}

func inquireRestart() {
	url := config.ServiceUrl + "v1/inquire/restart/"
	body := doRequest(url, "")
	var response response
	err := json.Unmarshal(body, &response)
	if err != nil {
		h.Warnf("Could not parse JSON response: " + string(body) + " Error: " + err.Error())
	}
	if len(response.Error) > 1 {
		h.Fatalf("Recieved error: " + response.Error)
		h.Infof("Received valid response from " + url)
	}

	if strings.HasPrefix(response.Message, "YesInquireToRestart") {
		h.Infof("Recieved reason from middle-ware to restart: " + response.Message)
		doRestart()
	}

}

func askForOSRestart(rid string) response {
	url := config.ServiceUrl + "v1/request/restart/os"
	body := doRequest(url, rid)
	var response response
	err := json.Unmarshal(body, &response)
	if err != nil {
		h.Warnf("Could not parse JSON response: " + string(body) + " Error: " + err.Error())
	}
	if len(response.Error) > 1 {
		h.Fatalf("Recieved error: " + response.Error)
	}
	h.Infof("Received valid response from " + url)
	return response
}

func getPayload(rid string) *bytes.Buffer {
	var req request

	if len(rid) > 0 {
		req.RequestID = rid
	}
	if flag.Lookup("test.v") == nil {
		req.Fqdn = getPayloadFqdn()
		req.Uptime = getPayloadUptime()
	} else {
		req.Fqdn = "foobar-server-aa02.domain.tld"
		if os.Getenv("TEST_FOR_CRASH_TestUptimeLow") == "1" {
			req.Uptime = (time.Duration(2) * time.Second).String()
		} else {
			req.Uptime = (time.Duration(83836) * time.Second).String()
		}
	}

	reqBytes, err := json.Marshal(req)
	if err != nil {
		h.Fatalf("Error while json.Marshal request. Error: " + err.Error())
	}

	h.Debugf("Trying to send payload: " + string(reqBytes))

	return bytes.NewBuffer(reqBytes)
}

func doRequest(url string, rid string) []byte {
	h.Debugf("sending HTTP request " + url)
	payload := getPayload(rid)
	resp, err := client.Post(url, "application/json", payload)
	if err != nil {
		h.Fatalf("Error while issuing request to " + url + " Error: " + err.Error())
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		h.Fatalf("Error while reading response body: " + err.Error())
	}
	h.Debugf("Received response: " + string(body))

	return body

}

func main() {
	log.SetOutput(os.Stdout)

	var (
		configFileFlag   = flag.String("config", "/etc/goahead/client.yml", "which config file to use")
		disabledFileFlag = flag.String("disabled", "/etc/goahead/disabled", "file to check if goahead run should be skipped")
		versionFlag      = flag.Bool("version", false, "show build time and version number")
	)
	flag.BoolVar(&debug, "debug", false, "log debug output, defaults to false")
	flag.Parse()

	configFile := *configFileFlag
	disabledFile := *disabledFileFlag
	version := *versionFlag

	if version {
		fmt.Println("goahead client version 0.0.2 Build time:", buildtime, "UTC")
		os.Exit(0)
	}

	h.Info = true
	h.Debug = debug
	h.InfoTimestamp = true
	h.WarnExit = true

	h.Debugf("Using as config file: " + configFile)
	config = readConfigfile(configFile)
	client = setupHttpClient()
	if h.FileExists(disabledFile) {
		data, err := ioutil.ReadFile(disabledFile)
		if err != nil {
			h.Fatalf("There was an error parsing the file to disabled goahead" + disabledFile + ": " + err.Error())
		}
		reason := "reason not specified"
		if len(data) > 0 {
			reason = string(data)
			reason = strings.ReplaceAll(reason, "\n", "")
		}
		fmt.Printf("Notice: Skipping run of goahead client; administratively disabled (Reason: '%s')\n", reason)
	} else {
		doMain()
	}

}

func setupHttpClient() *http.Client {
	// Get the SystemCertPool, continue with an empty pool on error
	rootCAs, _ := x509.SystemCertPool()
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	if len(config.ServiceUrlCaFile) > 0 {
		// Read in the cert file
		certs, err := ioutil.ReadFile(config.ServiceUrlCaFile)
		if err != nil {
			h.Fatalf("Failed to append " + config.ServiceUrlCaFile + " to RootCAs Error: " + err.Error())
		}

		// Append our cert to the system pool
		h.Debugf("Appending certificate " + config.ServiceUrlCaFile + " to trusted CAs")
		if ok := rootCAs.AppendCertsFromPEM(certs); !ok {
			h.Debugf("No certs appended, using system certs only")
		}
	}

	// Trust the augmented cert pool in our client
	tlsConfig := &tls.Config{
		RootCAs: rootCAs,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: tr}
}

func doMain() {
	er := h.ExecuteCommand(config.RestartConditionScript, 5, true)
	if er.ReturnCode == config.RestartConditionScriptExitCodeForReboot {
		doRestart()
	} else {
		h.Infof("Did not find local reason to restart. Asking if I should restart, because of other reasons.")
		inquireRestart()
	}
}

func doRestart() {
	response := askForOSRestart("")
	if len(response.FoundCluster) < 1 || len(response.AskagainIn) == 0 {
		h.Warnf(response.Message + " Exiting...")
	}
	h.Infof("Sleeping for " + response.AskagainIn)
	sleep, err := time.ParseDuration(response.AskagainIn)
	if err != nil {
		h.Fatalf("Error while trying to parse response.AskagainIn to Duration. Error: " + err.Error())
	}
	time.Sleep(sleep)
	response = askForOSRestart(response.RequestID)

	if response.Goahead {
		// execute hooks and check their exit code
		executeRestartHooks()
	} else {
		h.Infof("Did not recieve go ahead to restart. Reason: " + response.Message)
	}

}

func executeRestartHooks() {
	if len(config.OsRestartHooksDir) > 0 {
		if h.IsDir(config.OsRestartHooksDir) {
			globPath := filepath.Join(config.OsRestartHooksDir, "*")
			h.Debugf("Glob'ing with path " + globPath)
			matches, err := filepath.Glob(globPath)
			if len(matches) == 0 {
				h.Fatalf("Could not find any restart hook scripts matching " + globPath)
			}
			h.Debugf("found pre restart hook script: " + strings.Join(matches, " "))
			if err != nil {
				h.Fatalf("Failed to glob pre restart hook script directory with glob path " + globPath + " Error: " + err.Error())
			}
			sort.Strings(matches)
			for _, file := range matches {
				_ = h.ExecuteCommand(file, 10, config.OsRestartHooksAllowFail)
			}
		}
	}
}
