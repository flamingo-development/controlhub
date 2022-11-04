package main

import (
	"encoding/json"
	"errors"

	"flamingo.dev/control-hub/pkg/control"
	"github.com/jason0x43/go-toggl"
)

type TogglHook struct{}

func (d TogglHook) Info() control.PluginInfo {
	return control.PluginInfo{
		Name:        "toggl",
		Version:     "0.0.1",
		Description: "Toggl integration to start or stop your toggl tracker",
		Website:     "",
		License:     "MIT",
	}
}

func (d TogglHook) CreateHook(instance control.HookInstance) (control.Hook, error) {

	api_token := ""
	if value, ok := instance.Config["api_token"]; ok {
		if value, ok := value.(string); ok {
			api_token = value
		}
	}
	if api_token == "" {
		return nil, errors.New("toggl integration requires `api_token` in its config")
	}

	action := ""
	if value, ok := instance.Config["action"]; ok {
		if value, ok := value.(string); ok {
			action = value
		}
	}

	description := ""
	if value, ok := instance.Config["description"]; ok {
		if value, ok := value.(string); ok {
			description = value
		}
	}

	return &TogglIntegration{
		API_token:   api_token,
		Action:      action,
		Description: description,
	}, nil
}

type TogglIntegration struct {
	API_token   string
	Action      string
	Description string

	Session *toggl.Session
}

func (d TogglIntegration) Name() string {
	return "Toggl"
}

func (d *TogglIntegration) Call(payload []byte) error {
	if d.Session == nil {
		session := toggl.OpenSession(d.API_token)
		d.Session = &session
	}

	action := d.Action
	description := d.Description

	type TogglPayload struct {
		Action      string `json:"action"`
		Description string `json:"description"`
	}

	var p TogglPayload
	err := json.Unmarshal(payload, &p)
	if err == nil {
		if p.Action != "" {
			action = p.Action
			description = p.Description
		}
	}

	switch action {
	case "start":
		_, err := d.Session.StartTimeEntry(description)
		return err

	case "stop":
		currentTimer, err := d.Session.GetCurrentTimeEntry()
		if err != nil {
			if err.Error() == "No time entry is running" {
				return nil
			}
			return err
		}

		_, err = d.Session.StopTimeEntry(currentTimer)
		return err

	default:
		return nil
	}
}

func (d TogglIntegration) End() error {
	return nil
}

var Plugin TogglHook
