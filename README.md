# otel-tui

🚧 This project is under construction 🚧

A terminal OpenTelemetry viewer inspired by [otel-desktop-viewer](https://github.com/CtrlSpice/otel-desktop-viewer/tree/main)

Traces
![Traces](./docs/traces.png)
![Spans](./docs/spans.png)

Logs
![Logs](./docs/logs.png)

## Getting Started
### Homebrew

```sh
$ brew install ymtdzzz/tap/otel-tui
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
  - [ ] Display metrics
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
