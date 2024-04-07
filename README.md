# otel-tui

🚧 This project is under construction 🚧

A terminal OpenTelemetry viewer inspired by [otel-desktop-viewer](https://github.com/CtrlSpice/otel-desktop-viewer/tree/main)

Traces
![Traces](./docs/traces.png)
![Spans](./docs/spans.png)

Logs
![Logs](./docs/logs.png)

## Getting Started

This project is currently in the early stages of development, so you can only run it using the `go run` command:

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
  - [ ] Display metrics
  - [ ] ...
- Logs
  - [x] Display logs
  - [x] Filter logs
  - [x] Show log information
  - [ ] Show logs related to a specific trace or span
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
