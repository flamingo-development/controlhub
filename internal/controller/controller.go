package controller

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"plugin"
	"strings"
	"time"

	"flamingo.dev/control-hub/pkg/control"
	"github.com/dop251/goja"
	"github.com/gorhill/cronexpr"
	"golang.org/x/exp/slices"
)

const ENV_VAR_PREFIX = "$env."

type PluginController struct {
	plugins map[string]control.Plugin

	cron  map[string]*control.Cronjob
	hooks map[string]control.Hook

	config control.Config
}

func NewPluginController() *PluginController {
	return &PluginController{
		plugins: make(map[string]control.Plugin),

		cron:  make(map[string]*control.Cronjob),
		hooks: make(map[string]control.Hook),

		config: control.Config{},
	}
}

// Init loads all the plugins from the ./plugins directory using the golang plugin package and loads the config from ./config.json
func (c *PluginController) Init() error {
	files, err := os.ReadDir("./plugins")
	if err != nil {
		return err
	}

	for _, f := range files {
		if f.IsDir() {
			continue
		}

		plugin, err := plugin.Open("./plugins/" + f.Name())
		if err != nil {
			continue
		}

		symPlugin, err := plugin.Lookup("Plugin")
		if err != nil {
			return err
		}

		plug, ok := symPlugin.(control.Plugin)
		if !ok {
			return errors.New("plugin does not implement control.Plugin")
		}

		info := plug.Info()

		fmt.Printf("Loaded plugin '%s:%s'\n", info.Name, info.Version)
		c.plugins[info.Name] = plug
	}

	configBytes, err := os.ReadFile("./config.json")
	if err != nil {
		return err
	}

	for _, v := range os.Environ() {
		parts := strings.Split(v, "=")
		configBytes = []byte(strings.ReplaceAll(string(configBytes), ENV_VAR_PREFIX+parts[0], parts[1]))
	}

	if err := json.Unmarshal(configBytes, &c.config); err != nil {
		return err
	}

	now := time.Now()

	for name, instance := range c.config.Cron {
		var expr *cronexpr.Expression
		var err error
		var next time.Time

		if instance.Time == "@reboot" {
			goto skipExpr
		}

		expr, err = cronexpr.Parse(instance.Time)
		if err != nil {
			return fmt.Errorf("invalid cron expression: '%v'", instance.Time)
		}
		next = expr.Next(now)

	skipExpr:

		c.cron[name] = &control.Cronjob{
			Time:   expr,
			Next:   next,
			OnBoot: instance.Time == "@reboot",
			Data:   instance.Data,
			To:     instance.To,
		}
	}

	return nil
}

func (c *PluginController) Start() error {
	defer func() {
		for _, hook := range c.hooks {
			hook.End()
		}
	}()

	for name, instance := range c.config.Outputs {
		plugin, ok := c.plugins[instance.Type]
		if !ok {
			return fmt.Errorf("plugin '%s' not found", instance.Type)
		}

		hook, err := plugin.CreateHook(instance)
		if err != nil {
			return err
		}

		c.hooks[name] = hook
	}

	c.startCronjobs()
	fmt.Printf("Started %v cronjobs\n", len(c.cron))
	fmt.Println("Listening on :8080")
	return http.ListenAndServe(":8080", c.Handler())
}

func (c *PluginController) startCronjobs() {
	filtered := map[string]*control.Cronjob{}
	for name, cronjob := range c.cron {
		if cronjob.OnBoot {
			fmt.Printf("Running cronjob '%v' on boot\n", name)

			payload, err := json.Marshal(cronjob.Data)
			if err != nil {
				fmt.Printf("cronjob '%s': %s\n", name, err)
				continue
			}

			err = c.CallHooks(cronjob.To, payload)
			if err != nil {
				fmt.Printf("cronjob '%s': %s\n", name, err)
				continue
			}
		} else {
			filtered[name] = cronjob
		}
	}

	c.cron = filtered

	go func() {
		ticker := time.NewTicker(time.Minute)

		for {
			now := <-ticker.C

			for name, cron := range c.cron {
				if cron.Next.Before(now) {
					cron.Next = cron.Time.Next(now)

					fmt.Printf("Running cronjob '%v'\n", name)
					payload, err := json.Marshal(cron.Data)
					if err != nil {
						fmt.Printf("cronjob '%s': %s\n", name, err)
						continue
					}

					if err := c.CallHooks(cron.To, payload); err != nil {
						fmt.Printf("cronjob '%s': %s\n", name, err)
					}
				}
			}
		}
	}()
}

func (c *PluginController) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, err := io.ReadAll(r.Body)
		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, input := range c.config.Inputs {
			if input.Token == r.URL.Query().Get("token") {
				err := c.PerformInput(input, data)

				if err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
			}
		}

		w.WriteHeader(http.StatusOK)
	})
}

func (c *PluginController) PerformInput(input control.InputInstance, data []byte) error {

	if input.Formatter != "" {
		vm := goja.New()
		vm.Set("input", string(data))
		vm.Set("output", "")

		content, err := os.ReadFile(input.Formatter)
		if err != nil {
			return err
		}

		_, err = vm.RunScript("formatter.js", string(content))
		if err != nil {
			return err
		}

		data = []byte(vm.Get("output").String())
	}

	if slices.Contains(input.To, "*") {
		return c.CallAllHooks(data)
	}

	for _, hook := range input.To {
		if err := c.CallHook(hook, data); err != nil {
			return err
		}
	}

	return nil
}

func (c *PluginController) CallAllHooks(payload []byte) error {
	for _, h := range c.hooks {
		if err := h.Call(payload); err != nil {
			return err
		}
	}

	return nil
}

func (c *PluginController) CallHooks(hooks []string, payload []byte) error {
	for _, hook := range hooks {
		if err := c.CallHook(hook, payload); err != nil {
			return err
		}
	}

	return nil
}

func (c *PluginController) CallHook(hook string, payload []byte) error {
	if h, ok := c.hooks[hook]; ok {
		return h.Call(payload)
	}

	return nil
}
