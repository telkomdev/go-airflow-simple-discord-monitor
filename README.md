## Airflow Discord uptime monitor
Simple Airflow monitor with Discord written in Go

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev/)
[![Go Reference](https://pkg.go.dev/badge/github.com/telkomdev/go-airflow-simple-discord-monitor.svg)](https://pkg.go.dev/github.com/telkomdev/go-airflow-simple-discord-monitor)
[![Go Report Card](https://goreportcard.com/badge/github.com/telkomdev/go-airflow-simple-discord-monitor)](https://goreportcard.com/badge/github.com/telkomdev/go-airflow-simple-discord-monitor)
[![GitHub license](https://img.shields.io/github/license/telkomdev/go-airflow-simple-discord-monitor.svg)](https://github.com/telkomdev/go-airflow-simple-discord-monitor/blob/master/LICENSE)

### Build
```shell
$ git clone https://github.com/telkomdev/go-airflow-simple-discord-monitor \
    && cd https://github.com/telkomdev/go-airflow-simple-discord-monitor \
    && make build \
    && sudo make install
```

### Example config files
*config.json*
```json
{
    "airflow_url": "https://my-airflow-server.local",
    "interval": 60,
    "discord_thread_url": "https://my-discord-webhook",
    "discord_name": "My Bot",
    "discord_avatar_url": "https://my-domain/my-avatar.jpg"
}
```

### Run Airflow monitor
```shell
$ ./go-airflow-simple-discord-monitor

info: Scheduling task to run every 60 seconds
info: Executing health check
...
```

### Example generated alert
![Alt text](/screenshots/example-discord-alert.png?raw=true "Discord Alert")

### Todo
- Add unit test
- Improve documentation
- Improve code quality
