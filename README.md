# Go SQS Load Test

This is a simple SQS load test tool written in Go.

Performs concurrent sends over an SQS queue, and computes average throughput and latency.

This project was forked from https://github.com/pedro-gutierrez/go-sqs

## Goal 

After forked, it was changed to fit a specific case:

> The application under load test owns a SQS queue subscribed to an SNS topic.
> But the SNS topic is owned by another application.

The publish-to-SQS action is not important, thus it can be executed from local machines.

On the other hand, the SQS consumption must happen in a real environment, to simulate real application execution.

## How to use this tool

```shell
go mod vendor
```

### Change the code to fit your needs

Configure the execution:

```go
var (
	total           = 1000 // amount of publishes to make
	sendConcurrency = 25 // amount of go routines
	qName           = "your-queue-name" // the SQS queue name
)
```

Customise the `getRandomBody()` function to fit your needs. The following example simulates a message coming from an SNS topic:
```go
func getRandomBody() (body string) {
	body = `
	{
	  "Type": "Notification",
	  "MessageId": "%s",
	  "TopicArn": "arn:aws:sns:eu-west-1:000000000000:this-is-the-topic-origin",
	  "Message": "{\"meta\":{},\"data\":{\"partner_urn\":\"urn:site:something\",\"id\":\"%s\",\"buyer_uuid\":\"...\",\"seller_uuid\":\"...\",\"advert_id\":\"...\",\"channel\":\"...\",\"status\":\"WAITING_FOR_ANSWER\",\"conversation_occurred_at\":\"2022-05-07 04:43:28.266215\",\"advert_category\":\"...\"}}",
	  "Timestamp": "2022-06-27T09:26:18.423Z",
	  "SignatureVersion": "1",
	  "Signature": "EXAMPLEpH+..",
	  "SigningCertURL": "https://sns.us-east-1.amazonaws.com/SimpleNotificationService-0000000000000000000000.pem",
	  "Subject": null
	}`

body = fmt.Sprintf(body, uuid.NewRandom().String(), uuid.NewRandom().String())
return
}
```

Get your AWS credentials information and run the program:

```shell
env AWS_ACCESS_KEY_ID="localstack" \
  AWS_SECRET_ACCESS_KEY="secret" \
  AWS_SESSION_TOKEN="" \
  AWS_REGION="eu-west-1" \
  AWS_ENDPOINT="http://localhost:44569" \
  go run main.go
```
