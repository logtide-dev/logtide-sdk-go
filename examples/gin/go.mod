module example-gin

go 1.21

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/logward-dev/logward-sdk-go v0.1.0
)

replace github.com/logward-dev/logward-sdk-go => ../..
