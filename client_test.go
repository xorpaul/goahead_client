package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"

	H "github.com/xorpaul/gohelper"
)

var (
	defaultUrl = "https://127.0.0.1:8443/"
	ts         *httptest.Server
)

func spinUpFakeGoahead() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var request request
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&request); err != nil {
			H.Fatalf("Error while reading response body: " + err.Error())
			return
		}
		defer r.Body.Close()

		var responseFile string
		uptime, err := time.ParseDuration(request.Uptime)
		if err != nil {
			log.Fatal(err)
		}
		if uptime.Seconds() < 1800 {
			responseFile = "tests/uptime-too-low.json"
		} else if r.URL.Path == "/v1/inquire/restart/" {
			responseFile = "tests/inquireRestart-false.json"
		} else if r.URL.Path == "/v1/request/restart/os" {
			if request.RequestID == "sqEALyco" {
				responseFile = "tests/goahead-true.json"
			} else {
				responseFile = "tests/requestRestart-true.json"
			}
		} else {
			log.Fatal("Unexpected request URL: " + r.URL.Path)
		}
		body, err := ioutil.ReadFile(responseFile)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Fprint(w, string(body))
	}))
	return ts

}

func TestMain(m *testing.M) {
	H.WarnExit = true
	config = readConfigfile("./config.yml")
	ts = spinUpFakeGoahead()
	defer ts.Close()
	// default config overwrites for test cases
	config.ServiceUrl = ts.URL + "/"
	config.RestartConditionScript = "./tests/always-true.sh"
	config.OsRestartHooksDir = "./tests/TestRestartHooks/"
	client = setupHttpClient()
	exitCode := m.Run()
	os.Exit(exitCode)
}

func TestRestartConditionScriptFalse(t *testing.T) {
	config.RestartConditionScript = "./tests/always-false.sh"
	config.OsRestartHooksDir = "./tests/TestRestartHooks/"

	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		H.Debug = true
		doMain()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+H.FuncName()+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+H.FuncName()+"=1")
	out, err := cmd.CombinedOutput()

	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok { // there is error code
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if 0 != exitCode {
		t.Errorf("terminated with %v, but we expected exit status %v Output: %s", exitCode, 0, string(out))
	}

	expectedLines := [7]string{
		"Debug ExecuteCommand(): Executing ./tests/always-false.sh",
		"Did not find local reason to restart. Asking if I should restart, because of other reasons.",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(string(out), expectedLine) {
			t.Error("Could not find expected line '" + expectedLine + "' in output.")
		}
	}

	//fmt.Println(string(out))
}

func TestRestartConditionScriptTrue(t *testing.T) {
	preRestartHooksFile := "/var/tmp/goahead_client/restart_was_triggered"
	H.PurgeDir(preRestartHooksFile, H.FuncName())

	config.RestartConditionScript = "./tests/always-true.sh"

	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		H.Debug = true
		doMain()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+H.FuncName()+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+H.FuncName()+"=1")
	out, err := cmd.CombinedOutput()

	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok { // there is error code
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if 0 != exitCode {
		t.Errorf("terminated with %v, but we expected exit status %v Output: %s", exitCode, 0, string(out))
	}

	expectedLines := []string{
		"Debug ExecuteCommand(): Executing ./tests/always-true.sh",
		"Sleeping for 1s",
		"Debug ExecuteCommand(): Executing tests/TestRestartHooks/001_pre_restart_trigger01.sh",
	}

	for _, expectedLine := range expectedLines {
		if !strings.Contains(string(out), expectedLine) {
			t.Error("Could not find expected line '" + expectedLine + "' in output.")
		}
	}

	if !H.FileExists(preRestartHooksFile) {
		t.Errorf("Resulting file from pre restart trigger missing: %s", preRestartHooksFile)
	}

	//fmt.Println(string(out))
}

func TestRestartConditionScriptTrueFailing(t *testing.T) {
	preRestartHooksFile := "/var/tmp/goahead_client/restart_was_triggered"
	H.PurgeDir(preRestartHooksFile, H.FuncName())

	config.RestartConditionScript = "./tests/always-true.sh"
	config.OsRestartHooksDir = "./tests/TestRestartHooksFailing/"

	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		H.Debug = true
		doMain()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+H.FuncName()+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+H.FuncName()+"=1")
	out, err := cmd.CombinedOutput()

	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok { // there is error code
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if 1 != exitCode {
		t.Errorf("terminated with %v, but we expected exit status %v Output: %s", exitCode, 1, string(out))
	}

	expectedLines := []string{
		"Debug ExecuteCommand(): Executing ./tests/always-true.sh",
		"Debug ExecuteCommand(): Executing tests/TestRestartHooksFailing/001_pre_restart_trigger01.sh",
	}
	for _, expectedLine := range expectedLines {
		if !strings.Contains(string(out), expectedLine) {
			t.Error("Could not find expected line '" + expectedLine + "' in output.")
		}
	}

	if H.FileExists(preRestartHooksFile) {
		t.Errorf("Resulting file from pre restart trigger should be missing, but exists: %s", preRestartHooksFile)
	}

	//fmt.Println(string(out))
}

func TestUptimeLow(t *testing.T) {
	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		H.Debug = true
		doMain()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+H.FuncName()+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+H.FuncName()+"=1")
	out, err := cmd.CombinedOutput()

	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok { // there is error code
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}

	if 1 != exitCode {
		t.Errorf("terminated with %v, but we expected exit status %v Output: %s", exitCode, 1, string(out))
	}

	expectedLines := []string{
		"Debug getPayload(): Trying to send payload: {\"fqdn\":\"foobar-server-aa02.domain.tld\",\"uptime\":\"2s\"",
		"WARN doRestart(): Configured minimum uptime for cluster: 30m0s was not reached by client's uptime: 2s Exiting...",
	}
	for _, expectedLine := range expectedLines {
		if !strings.Contains(string(out), expectedLine) {
			t.Error("Could not find expected line '" + expectedLine + "' in output.")
		}
	}

	preRestartHooksFile := "/var/tmp/goahead_client/restart_was_triggered"
	if H.FileExists(preRestartHooksFile) {
		t.Errorf("Resulting file from pre restart trigger should be missing, but exists: %s", preRestartHooksFile)
	}

	//fmt.Println(string(out))

}

// a restart condition script that exits with a code other than 0 or 1 must be compared with restart_condition_script_exit_code_for_reboot as it is
func TestRestartConditionScriptExitCode2(t *testing.T) {
	preRestartHooksFile := "/var/tmp/goahead_client/restart_was_triggered"
	H.PurgeDir(preRestartHooksFile, H.FuncName())

	config.RestartConditionScript = "./tests/exit-2.sh"
	config.RestartConditionScriptExitCodeForReboot = 2
	config.OsRestartHooksDir = "./tests/TestRestartHooks/"

	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		H.Debug = true
		doMain()
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run="+H.FuncName()+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+H.FuncName()+"=1")
	out, err := cmd.CombinedOutput()

	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok {
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}
	if 0 != exitCode {
		t.Errorf("terminated with %v, but we expected exit status %v Output: %s", exitCode, 0, string(out))
	}

	for _, expectedLine := range []string{
		"Debug ExecuteCommand(): Executing ./tests/exit-2.sh",
		"Sleeping for 1s",
	} {
		if !strings.Contains(string(out), expectedLine) {
			t.Error("Could not find expected line '" + expectedLine + "' in output: " + string(out))
		}
	}
}

func runSlowHook(t *testing.T, name string, timeout int) (string, int) {
	if os.Getenv("TEST_FOR_CRASH_"+name) == "1" {
		H.Debug = true
		config.RestartConditionScript = "./tests/always-true.sh"
		config.OsRestartHooksDir = "./tests/TestRestartHooksSlow/"
		config.OsRestartHooksTimeout = timeout
		executeRestartHooks()
		return "", 0
	}
	cmd := exec.Command(os.Args[0], "-test.run="+name+"$")
	cmd.Env = append(os.Environ(), "TEST_FOR_CRASH_"+name+"=1")
	out, err := cmd.CombinedOutput()
	exitCode := 0
	if msg, ok := err.(*exec.ExitError); ok {
		exitCode = msg.Sys().(syscall.WaitStatus).ExitStatus()
	}
	return string(out), exitCode
}

func TestRestartHookNoTimeout(t *testing.T) {
	finished := "/var/tmp/goahead_client/slow_hook_finished"
	H.PurgeDir(finished, H.FuncName())
	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		runSlowHook(t, H.FuncName(), 0)
		return
	}
	out, exitCode := runSlowHook(t, H.FuncName(), 0)
	if exitCode != 0 {
		t.Errorf("terminated with %v, but we expected 0 Output: %s", exitCode, out)
	}
	if !H.FileExists(finished) {
		t.Errorf("slow hook was killed although no timeout is set. Output: %s", out)
	}
}

func TestRestartHookTimeout(t *testing.T) {
	finished := "/var/tmp/goahead_client/slow_hook_finished"
	H.PurgeDir(finished, H.FuncName())
	if os.Getenv("TEST_FOR_CRASH_"+H.FuncName()) == "1" {
		runSlowHook(t, H.FuncName(), 1)
		return
	}
	out, _ := runSlowHook(t, H.FuncName(), 1)
	if H.FileExists(finished) {
		t.Errorf("slow hook should have been killed by the 1s timeout. Output: %s", out)
	}
	if !strings.Contains(out, "signal: killed") {
		t.Errorf("expected killed message in output: %s", out)
	}
}
