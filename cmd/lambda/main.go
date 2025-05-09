package main

import (
	"theary_test/internal/lambda"

	awslambda "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	awslambda.Start(lambda.Handler)
}
