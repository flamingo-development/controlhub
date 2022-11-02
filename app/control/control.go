package main

import "flamingo.dev/control-hub/internal/controller"

func main() {
	c := controller.NewPluginController()
	err := c.Init()
	if err != nil {
		panic(err)
	}

	err = c.Start()
	if err != nil {
		panic(err)
	}
}
