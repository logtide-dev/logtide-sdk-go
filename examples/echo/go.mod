module example-echo

go 1.21

require (
	github.com/labstack/echo/v4 v4.12.0
	github.com/logtide-dev/logtide-sdk-go v0.1.0
)

replace github.com/logtide-dev/logtide-sdk-go => ../..
