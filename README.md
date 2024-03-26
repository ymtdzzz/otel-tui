# otel-tui

🚧 This project is under construction 🚧

A terminal OpenTelemetry viewer inspired by [otel-desktop-viewer](https://github.com/CtrlSpice/otel-desktop-viewer/tree/main)

![Traces](./docs/traces.png)
![Spans](./docs/spans.png)


## Getting Started

```sh
# install the CLI tool
go install github.com/ymtdzzz/otel-tui@main

# run the CLI tool (running on localhost:4317 by default)
otel-tui
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
- UI
  - [ ] Improve UI
  - [ ] Add more keybindings
  - [ ] ...
- Performance
  - [x] Timer based refresh
  - [x] Data rotation (current buffer size: 1000 service root spans)
  - [ ] ...
- Configurations
  - [ ] Endpoint
  - [ ] Refresh interval
  - [ ] Buffer size
  - [ ] ...
