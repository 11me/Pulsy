# Pulsy

Pulsy is an open-source monitoring tool that keeps an eye on your critical
systems so you don't have to. With Pulsy, you can easily configure custom
monitors for your applications and servers, set up notification channels for
alerts, and view historical data on uptime and performance metrics.


## Features

- Monitors multiple URLs simultaneously
- Configurable interval between requests
- Retry mechanism to handle temporary failures
- Logs HTTP response codes and response times to stdout/stderr (for now)
- Sends notifications to Telegram and Discord when the website is down
- Supports YAML configuration file for easy setup

## Getting started

To get started you can follow the instructions in the installation
guide to install the application on your system. Once installed, you can
configure custom monitors using the YAML configuration file, and set up
notification channels for alerts.

## Installation

Pulsy can be installed using Docker or by building the application from source.

```sh
$ cd Pulsy

$ go build .

$ ./pulsy -c config.yaml
```

## Configuration

To configure Pulsy, you can create a YAML configuration file with the following fields:
* monitors: a list of objects, each representing a monitor with the following fields:
  * name: a unique name for the monitor
  * url: the URL to monitor
  * interval: the time interval between requests to the URL
  * timeout: the maximum time to wait for a response from the URL
  * retry: the number of times to retry if a request fails

* notifiers: a list of objects, each representing a notification channel with the following fields:
  * name: a unique name for the notifier
  * type: the type of notifier (currently only "telegram" is supported)
  * options: a set of options specific to the notifier type (e.g. "token" for Telegram)

Here is an example:
```yaml
monitors:
  - name: "My Website"
    url: "https://www.example.com"
    interval: 5s
    timeout: 10s
    retry: 3
  - name: "My API"
    url: "https://api.example.com"
    interval: 10s
    timeout: 5s
    retry: 5

notifiers:
  - name: "Telegram"
    type: "telegram"
    options:
      token: "YOUR_TELEGRAM_BOT_TOKEN"
      chat_id: "YOUR_TELEGRAM_CHAT_ID"
```


## Contribution

Please feel free to submit bug reports, feature requests, or pull requests. We
welcome contributions in the form of code, documentation, or even just ideas
for improvement. 
By contributing to Pulsy, you can help make it a more robust and feature-rich
monitoring tool for everyone to use.

