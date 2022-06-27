# Go SQS Load Test

A simple SQS load test tool written in Go.

Performs concurrent sends over an SQS queue, and computes average throughput and latency.

This project was forked and changed aiming to only SEND to SQS.
The SQS consumption must happen in a real environment, to simulate real application execution.

```shell
go mod vendor
env AWS_ACCESS_KEY_ID="localstack" \
  AWS_SECRET_ACCESS_KEY="secret" \
  AWS_SESSION_TOKEN="" \
  AWS_REGION="eu-west-1" \
  AWS_ENDPOINT="http://localhost:44569" \
  go run main.go
```