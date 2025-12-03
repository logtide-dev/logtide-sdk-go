module example-echo

go 1.21

require (
	github.com/labstack/echo/v4 v4.12.0
	github.com/logward-dev/logward-sdk-go v0.1.0
)

replace github.com/logward-dev/logward-sdk-go => ../..
