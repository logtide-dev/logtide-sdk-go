module example-gin

go 1.21

require (
	github.com/gin-gonic/gin v1.10.0
	github.com/logtide-dev/logtide-sdk-go v0.1.0
)

replace github.com/logtide-dev/logtide-sdk-go => ../..
