# Control Hub

Control Hub is a small application written to handle webhooks and format them into other webhooks not supported by other services.
One example of this is for terraform which does not support discord webhooks.
This also supports sending a single webhook to multiple endpoints at once.
Aswell as cronjobs triggering webhooks.

This project can also support different plugins to widen the scope of triggers, and formatters.
For example you could make a webhook to trigger a plugin to restart a service on a server.
Or have a cronjob sending you a good morning email each morning.

## Installation

In order to setup this application you will need 2 files, `config.json` and `install.sh`.
The `config.json` defines your webhook endpoints, cronjobs and outputs.
The `install.sh` runs on each startup of the server to install any plugins you might want to use.

If not given, the application will use the default thats already provided in the project.

As for the locations that the files need to be placed in:
- `config.json` needs to be placed in `/app/config.json`
- `install.sh` needs to be placed in `/app/plugins/install.sh`

### config.json

The config.json file is used to define your webhook endpoints, cronjobs and outputs.
You can use `$env.VARIABLE` to use environment variables in the config file.
This is useful for things like passwords, tokens and urls.

Inputs can also use formatters to change the data sent to the webhook.
This can be done using a javascript file that has to be included in the docker container.
Within this javascript file, the input data is defined as a global variable `input` and the output is defined as a global variable `output`.

Both of these global variables are strings, so you will need to parse them into a json object if you want to use them as such using `JSON.parse(input)` and `JSON.stringify(output)`.

```json
{
    "inputs": {
        "input-name": {
            "token": "$env.WEBHOOK_TOKEN",
            "to": [
                "output-name"
            ],
            "formatter": "./formatter/location/in/dockerfile.js"
        }
    },

    "cron": {
        "cronjob-name": {
            "time": "0 0 * * *",
            "data": {
                "data": "fields",
                "that": "the webhook will send"
            },
            "to": [
                "output-name"
            ]
        }
    },

    "outputs": {
        "output-name": {
            "type": "plugin-name",
            "config": {
                "whatever": "config",
                "the": "plugin",
                "wants": "!"
            }
        }
    }
}
```

### install.sh

The install.sh file is used to install any plugins you might want to use.
This file is run on each startup of the server.
This file is not required, but if you want to use plugins, you will need to create this file.
If not included, the default will be used which installs the default plugins included in the project.
They do need to be built within the container to be built with the correct GOARCH and GOOS.

```bash
#!/bin/bash

# Install plugins here
# Example:
go build -o ./plugins/generic.so -buildmode=plugin ./default_plugins/generic/generic.go
```

## Plugins

Plugins are used to extend the functionality of the application.
There are default plugins which show the basic examples of how to create a plugin.
The default plugins are located in the `default_plugins` folder.

All plugins must be built as a shared object file with the buildmode set to plugin.
This can be done using the command `go build -o ./plugins/plugin-name.so -buildmode=plugin ./plugin-location/plugin.go`.

The interfaces and structs that are used to create a plugin are located in `/pkg/control/plugin.go`.

# Generic Plugin

The generic plugin is a plugin that can be used to send a webhook to any endpoint.
This is useful for sending a webhook to a service that does not have a plugin for it.

The config for this plugin is as follows:

```json
"output-name": {
    "type": "generic",
    "config": {
        "url": "$env.WEBHOOK_URL",
        "headers": {
            "Authorization": "Bearer $env.WEBHOOK_TOKEN",
            "X-Other-Header": "value"
        }
    }
}
```