build-lambda::
	GOOS=linux GOARCH=amd64 go build -o ./internal/aws/handler ./internal/aws/handler/handler.go
	zip -j handler.zip ./internal/aws/handler/handler
