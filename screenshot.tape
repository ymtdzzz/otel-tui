Set Shell bash
Set Width 2500
Set Height 1500
Set FontSize 17

Type "./otel-tui --from-json-file ./test_data.json"
Enter
Sleep 3s

Type "/recommend"
Enter
Sleep 10ms
Down 3
Sleep 100ms
Screenshot docs/traces.png

Enter
Sleep 100ms
Type "L"
Ctrl+L
Sleep 10ms
Ctrl+L
Sleep 10ms
Ctrl+L
Sleep 10ms
Ctrl+L
Sleep 10ms
Ctrl+L
Sleep 10ms
Ctrl+L
Sleep 100ms
Screenshot docs/spans.png
Sleep 1s

Escape
Tab
Type "/redis"
Enter
Sleep 10ms
Down 8
Sleep 150ms
Screenshot docs/metrics.png
Sleep 1s

Tab
Type "/recommend"
Enter
Sleep 100ms
Down
Sleep 100ms
Screenshot docs/logs.png
Sleep 1s

Tab
Sleep 100ms
Screenshot docs/topology.png
Sleep 1s

Ctrl+C
Ctrl+C
