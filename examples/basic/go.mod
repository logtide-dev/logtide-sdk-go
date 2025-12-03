module example-basic

go 1.25.4

require github.com/logward-dev/logward-sdk-go v0.1.0

require (
	go.opentelemetry.io/otel v1.38.0 // indirect
	go.opentelemetry.io/otel/trace v1.38.0 // indirect
)

replace github.com/logward-dev/logward-sdk-go => ../..
