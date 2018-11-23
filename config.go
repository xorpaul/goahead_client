package main

import (
	"io/ioutil"
	"net/url"
	"strings"
	"time"

	h "github.com/xorpaul/gohelper"
	yaml "gopkg.in/yaml.v2"
)

// configSettings contains the key value pairs from the config file
type configSettings struct {
	Timeout                                 time.Duration `yaml:"timeout"`
	ServiceUrl                              string        `yaml:"service_url"`
	ServiceUrlCaFile                        string        `yaml:"service_url_ca_file"`
	Fqdn                                    string        `yaml:"requesting_fqdn"`
	PrivateKey                              string        `yaml:"ssl_private_key,omitempty"`
	CertificateFile                         string        `yaml:"ssl_certificate_file,omitempty"`
	RequireAndVerifyClientCert              bool          `yaml:"ssl_require_and_verify_client_cert"`
	RestartConditionScript                  string        `yaml:"restart_condition_script"`
	RestartConditionScriptExitCodeForReboot int           `yaml:"restart_condition_script_exit_code_for_reboot"`
	OsRestartHooksDir                       string        `yaml:"os_restart_hooks_dir"`
	OsRestartHooksAllowFail                 bool          `yaml:"os_restart_hooks_allow_fail"`
}

// readConfigfile creates the configSettings struct from the config file
func readConfigfile(configFile string) configSettings {
	h.Debugf("Trying to read config file: " + configFile)
	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		h.Fatalf("readConfigfile(): There was an error parsing the config file " + configFile + ": " + err.Error())
	}

	var config configSettings
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		h.Fatalf("In config file " + configFile + ": YAML unmarshal error: " + err.Error())
	}

	//fmt.Print("config: ")
	//fmt.Printf("%+v\n", config)

	// set default timeout to 5 seconds if no timeout setting found
	if config.Timeout == 0 {
		config.Timeout = 5
	}

	if len(config.ServiceUrl) < 1 {
		h.Fatalf("Missing service_url setting in config file: " + configFile)
	}
	_, err = url.ParseRequestURI(config.ServiceUrl)
	if err != nil {
		h.Fatalf("Failed to parse/validate service_url setting " + config.ServiceUrl + " in config file: " + configFile)
	}
	if !strings.HasSuffix(config.ServiceUrl, "/") {
		config.ServiceUrl = config.ServiceUrl + "/"
	}

	if len(config.ServiceUrlCaFile) > 0 && !h.FileExists(config.ServiceUrlCaFile) {
		h.Fatalf("Failed to find configured service_url_ca_file " + config.ServiceUrlCaFile)
	}

	if len(config.PrivateKey) > 0 && !h.FileExists(config.PrivateKey) {
		h.Fatalf("Failed to find configured ssl_private_key " + config.PrivateKey)
	}

	if len(config.CertificateFile) > 0 && !h.FileExists(config.CertificateFile) {
		h.Fatalf("Failed to find configured ssl_certificate_file " + config.CertificateFile)
	}

	if len(config.RestartConditionScript) < 1 {
		h.Fatalf("Missing restart_condition_script setting in config file: " + configFile)
	} else if !h.FileExists(config.RestartConditionScript) {
		h.Fatalf("Failed to find configured restart_condition_script " + config.RestartConditionScript)
	}

	if len(config.OsRestartHooksDir) < 1 {
		h.Fatalf("Missing os_restart_hooks_dir setting in config file: " + configFile)
	} else if !h.FileExists(config.OsRestartHooksDir) {
		h.Fatalf("Failed to find configured os_restart_hooks_dir " + config.OsRestartHooksDir)
	}

	return config
}
