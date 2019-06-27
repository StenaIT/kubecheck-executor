package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	program      string = "kubecheck-executor"
	statusPassed string = "passed"
)

var (
	kubecheckURL              string
	failedWebhookURL          string
	failedWebhookBodyTemplate string
	exitWebhookURL            string
	exitWebhookBodyTemplate   string
)

type apiResponse map[string]apiCheckResponse

type apiCheckResponse struct {
	Description string      `json:"description"`
	Status      string      `json:"status"`
	Reason      string      `json:"reason,omitempty"`
	Input       interface{} `json:"input,omitempty"`
	Output      interface{} `json:"output,omitempty"`
}

type templateData struct {
	Name        string
	Description string
	Timestamp   string
	Check       string
	Status      string
	Reason      string
	Input       string
	Output      string
}

func main() {
	if err := ensureConfigure(); err != nil {
		log.Fatal(err)
	}

	result, err := executeKubecheck()
	if err != nil {
		log.Fatal(err)
	}

	if err := handleResult(result); err != nil {
		log.Fatal(err)
	}

	log.Printf("invoking \"exit\" webhook (%s)\n", exitWebhookURL)
	invokeWebhook(exitWebhookURL, exitWebhookBodyTemplate, templateData{
		Name:        program,
		Description: "Program ran to completion",
		Timestamp:   time.Now().UTC().String(),
	})
}

func ensureConfigure() error {
	kubecheckURL = os.Getenv("KUBECHECK_URL")
	if kubecheckURL == "" {
		return fmt.Errorf("environment variable KUBECHECK_URL must be set")
	}

	failedWebhookURL = os.Getenv("FAILED_WEBHOOK_URL")
	if failedWebhookURL == "" {
		return fmt.Errorf("environment variable FAILED_WEBHOOK_URL must be set")
	}

	exitWebhookURL = os.Getenv("EXIT_WEBHOOK_URL")
	if exitWebhookURL == "" {
		return fmt.Errorf("environment variable EXIT_WEBHOOK_URL must be set")
	}

	failedWebhookBodyTemplate = os.Getenv("FAILED_WEBHOOK_BODY_TEMPLATE")
	exitWebhookBodyTemplate = os.Getenv("EXIT_WEBHOOK_BODY_TEMPLATE")

	return nil
}

func executeKubecheck() (apiResponse, error) {
	req, err := http.NewRequest("GET", kubecheckURL, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create a request: %v", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("unable to read the response body: %v", err)
	}

	result := apiResponse{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("unable to parse json response: %v", err)
	}

	return result, nil
}

func handleResult(result apiResponse) error {
	for name, check := range result {
		log.Printf("check=%s status=%s\n", name, check.Status)
		if check.Status != statusPassed {
			data := templateData{
				Name:        program,
				Check:       name,
				Status:      check.Status,
				Timestamp:   time.Now().UTC().String(),
				Description: check.Description,
				Reason:      check.Reason,
				Input:       toJSON(check.Input),
				Output:      toJSON(check.Output),
			}

			log.Printf("invoking \"failed\" webhook (%s)\n", failedWebhookURL)
			if err := invokeWebhook(failedWebhookURL, failedWebhookBodyTemplate, data); err != nil {
				return err
			}
		}
	}

	return nil
}

func toJSON(data interface{}) string {
	if data != nil {
		bytes, _ := json.Marshal(data)
		return string(bytes)
	}
	return ""
}

func invokeWebhook(url, body string, data templateData) error {
	var bodyReader io.Reader

	if body != "" {
		t := template.Must(template.New("body").Parse(body))

		var buf bytes.Buffer
		if err := t.Execute(&buf, data); err != nil {
			return fmt.Errorf("executing template: %v", err)
		}

		bodyReader = strings.NewReader(buf.String())
	}

	req, err := http.NewRequest("POST", url, bodyReader)
	if err != nil {
		return fmt.Errorf("unable to create a request: %v", err)
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("invoking webhook resulted in an error: %v", err)
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invoking webhook failed with HTTP status code: %d", res.StatusCode)
	}

	return nil
}
