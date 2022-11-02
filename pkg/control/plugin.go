package control

import (
	"time"

	"github.com/gorhill/cronexpr"
)

type PluginInfo struct {
	Name        string
	Version     string
	Description string
	Website     string
	Author      string
	License     string
}

type InputInstance struct {
	Token     string   `json:"token"`
	To        []string `json:"to"`
	Formatter string   `json:"formatter"`
}

type Cron struct {
	Time string   `json:"time"`
	Data any      `json:"data"`
	To   []string `json:"to"`
}

type HookInstance struct {
	Type   string         `json:"type"`
	Config map[string]any `json:"config"`
}

type Config struct {
	Inputs  map[string]InputInstance `json:"inputs"`
	Cron    map[string]Cron          `json:"cron"`
	Outputs map[string]HookInstance  `json:"outputs"`
}

type Cronjob struct {
	Time   *cronexpr.Expression
	Next   time.Time
	Data   any
	OnBoot bool
	To     []string
}

type Plugin interface {
	Info() PluginInfo
	CreateHook(instance HookInstance) (Hook, error)
}

type Hook interface {
	Name() string
	Call(payload []byte) error
	End() error
}
