# otel-tui

A terminal OpenTelemetry viewer inspired by [otel-desktop-viewer](https://github.com/CtrlSpice/otel-desktop-viewer/tree/main)

Traces
![Traces](./docs/traces.png)
![Spans](./docs/spans.png)

Metrics
![Metrics](./docs/metrics.png)

Logs
![Logs](./docs/logs.png)

## Getting Started
Currently, this tool exposes port 4317 to receive OpenTelemetry signals.

### Homebrew

```sh
$ brew install ymtdzzz/tap/otel-tui
```

### Docker

Run in the container simply:

```sh
$ docker run --rm -it --name otel-tui ymtdzzz/otel-tui:latest
```

Or, run as a background process and attach it:

```sh
# Run otel-tui as a background process
$ docker run --rm -dit --name otel-tui ymtdzzz/otel-tui:latest

# Show TUI in your current terminal session
$ docker attach otel-tui
```

### Docker Compose

First, add service to your manifest (`docker-compose.yaml`) for the instrumanted app

```yml
  oteltui:
    image: ymtdzzz/otel-tui:latest
    container_name: otel-tui
    stdin_open: true
    tty: true
```

Modify configuration for otelcol

```yml
exporters:
  otlp:
    endpoint: oteltui:4317
service:
  pipelines:
    traces:
      exporters: [otlp]
    logs:
      exporters: [otlp]
```

Run as a background process and attach it:

```sh
# Run services as usual
$ docker compose up -d

# Show TUI in your current terminal session
$ docker compose attach oteltui
```


### Executable Binary from Github Release page

https://github.com/ymtdzzz/otel-tui/releases

### From Source

```sh
$ git clone https://github.com/ymtdzzz/otel-tui.git
$ cd otel-tui
$ go run ./...
```

## TODOs

There're a lot of things to do. Here are some of them:

- Traces
  - [x] Display traces
  - [x] Filter traces
  - [x] Show trace information
  - [ ] ...
- Metrics
  - [x] Metric stream
    - [x] Display metric stream
    - [x] Filter metrics
    - [x] Show metric information
    - [x] Display basic chart of the selected metric
  - [ ] Metric list
    - [ ] Display metric stream
    - [ ] Flexible chart (query, selectable dimensions, etc.)
  - [ ] Auto refresh chart
  - [ ] Asynchronous chart rendering
  - [ ] ...
- Logs
  - [x] Display logs
  - [x] Filter logs
  - [x] Show log information
  - [x] Show logs related to a specific trace or span
  - [ ] ...
- UI
  - [ ] Improve UI
  - [ ] Add more keybindings
  - [ ] ...
- Performance
  - [x] Timer based refresh
  - [x] Data rotation (current buffer size: 1000 service root spans and logs)
  - [ ] ...
- Configurations
  - [ ] Port
  - [ ] Refresh interval
  - [ ] Buffer size
  - [ ] ...
