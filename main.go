package main

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/pborman/uuid"
)

var (
	total           = 1000
	sendConcurrency = 25
	qName           = "your-queue-name"
)

func main() {

	// Setup a new session to AWS
	// In our test, we contact a local GoAWS docker container on port 44569
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
		Credentials: credentials.NewStaticCredentials(
			os.Getenv("AWS_ACCESS_KEY_ID"),
			os.Getenv("AWS_SECRET_ACCESS_KEY"),
			os.Getenv("AWS_SESSION_TOKEN")),
		Endpoint: aws.String(os.Getenv("AWS_ENDPOINT")),
	})

	// Create a SQS client
	svc := sqs.New(sess)

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	fmt.Println("Created AWS session")

	// Get the url for the test queue. The url depends on the internal
	// GoAWS settings. Ensure the host/port are accessible from this test
	r, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: aws.String(qName),
	})

	if err != nil {
		fmt.Println("Error", err)
		return
	}

	qURL := *r.QueueUrl

	// Start sending items to SQS. We divide the total number of
	// messages to be sent across the different goroutine configured
	// by the sendConcurrency variable
	fmt.Println(fmt.Sprintf("Sending %v items to %v", total, qURL))

	start := time.Now()

	var wg sync.WaitGroup
	for i := 0; i < sendConcurrency; i++ {
		wg.Add(1)
		go Send(svc, qURL, total/sendConcurrency, &wg)
	}

	// We are done sending. Print some basic stats
	wg.Wait()
	PrintRate(start, total, "Send")
}

// PrintRate prints basic average rate and latency for the given context
func PrintRate(start time.Time, total int, context string) {
	duration := time.Since(start)
	rate := int(float64(total) / duration.Seconds())
	avgLatency := int(float64(1) / float64(rate) * 1000)
	fmt.Println(fmt.Sprintf("%v: %v (%v per sec ~ %v ms/req)", context, duration, rate, avgLatency))
}

// Send .
func Send(svc *sqs.SQS, qURL string, count int, wg *sync.WaitGroup) {

	for i := 0; i < count; i++ {
		_, err := svc.SendMessage(&sqs.SendMessageInput{
			DelaySeconds: aws.Int64(0),
			MessageBody:  aws.String(getRandomBody()),
			QueueUrl:     &qURL,
		})
		if err != nil {
			fmt.Println("Error", err)
			return
		}
	}

	wg.Done()
}

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
