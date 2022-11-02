package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"

	"flamingo.dev/control-hub/pkg/control"
)

type GenericHook struct{}

func (d GenericHook) Info() control.PluginInfo {
	return control.PluginInfo{
		Name:        "generic",
		Version:     "0.0.1",
		Description: "Generic webhook",
		Website:     "",
		License:     "MIT",
	}
}

func (d GenericHook) CreateHook(instance control.HookInstance) (control.Hook, error) {
	url := ""
	if value, ok := instance.Config["url"]; ok {
		if value, ok := value.(string); ok {
			url = value
		}
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}
	if value, ok := instance.Config["headers"]; ok {
		if value, ok := value.(map[string]interface{}); ok {
			for k, v := range value {
				if v, ok := v.(string); ok {
					headers[k] = v
				}
			}
		}
	}

	return GenericHookMethod{
		URL:     url,
		Headers: headers,
	}, nil
}

type GenericHookMethod struct {
	URL     string
	Headers map[string]string
}

func (d GenericHookMethod) Name() string {
	return "Generic"
}

func (d GenericHookMethod) Call(payload []byte) error {
	if d.URL == "" {
		return errors.New("URL is empty")
	}

	req, err := http.NewRequest("POST", d.URL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	for k, v := range d.Headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		fmt.Printf("Error '%d': %s/n", resp.StatusCode, string(data))
	}

	return nil
}

func (d GenericHookMethod) End() error {
	return nil
}

var Plugin GenericHook
