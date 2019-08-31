package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"sync"
	"time"
)

// Tune load test and concurrency
var (
	total           = 1000
	sendConcurrency = 25
	recvConcurrency = 25
	qName           = "local-pedro"
)

func main() {

	// Setup a new session to AWS
	// In our test, we contact a local GoAWS docker container on port 4575
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("us-west-1"),
		Credentials: credentials.NewStaticCredentials("localstack", "secret", ""),
		Endpoint:    aws.String("http://localhost:4575"),
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
		go Send(svc, qURL, body, total/sendConcurrency, &wg)
	}

	// We are done sending. Print some basic stats
	wg.Wait()
	PrintRate(start, total, "Send")

	// Start receiving. We have N goroutines consuming messages
	// from SQS and reporting via the 'received' channel
	// to the Count goroutine which helps synchronize and
	// wait until all messages have finally been consumed
	start = time.Now()

	received := make(chan int, 10)
	wg.Add(1)

	// The Count goroutines keeps the count of how
	// many items are left
	go Count(total, received, &wg)

	// Start consuming messages with the given receive
	// concurrency
	for i := 0; i < recvConcurrency; i++ {
		go Recv(svc, qURL, received)
	}

	// We are done receiving. Print some basic stats
	wg.Wait()
	PrintRate(start, total, "Recv")
}

// PrintRate prints basic average rate and latency for the given context
func PrintRate(start time.Time, total int, context string) {
	duration := time.Since(start)
	rate := int(float64(total) / duration.Seconds())
	avgLatency := int(float64(1) / float64(rate) * 1000)
	fmt.Println(fmt.Sprintf("%v: %v (%v per sec ~ %v ms/req)", context, duration, rate, avgLatency))
}

// Send
func Send(svc *sqs.SQS, qURL string, body string, count int, wg *sync.WaitGroup) {

	for i := 0; i < count; i++ {
		_, err := svc.SendMessage(&sqs.SendMessageInput{
			DelaySeconds: aws.Int64(0),
			MessageAttributes: map[string]*sqs.MessageAttributeValue{
				"Title": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("The Whistler"),
				},
				"Author": &sqs.MessageAttributeValue{
					DataType:    aws.String("String"),
					StringValue: aws.String("John Grisham"),
				},
				"WeeksOn": &sqs.MessageAttributeValue{
					DataType:    aws.String("Number"),
					StringValue: aws.String(fmt.Sprintf("%v", i)),
				},
			},
			MessageBody: aws.String(body),
			QueueUrl:    &qURL,
		})
		if err != nil {
			fmt.Println("Error", err)
			return
		}
	}

	wg.Done()
}

func Recv(svc *sqs.SQS, qURL string, received chan int) {
	for {
		result2, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
			QueueUrl:            &qURL,
			MaxNumberOfMessages: aws.Int64(10),
			VisibilityTimeout:   aws.Int64(20), // 20 seconds
			WaitTimeSeconds:     aws.Int64(20),
		})

		if err != nil {
			fmt.Println("Error", err)
			return
		}

		items := len(result2.Messages)

		if items == 0 {
			fmt.Println("Received no messages")
		} else {

			for _, m := range result2.Messages {

				_, err = svc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      &qURL,
					ReceiptHandle: m.ReceiptHandle,
				})

				if err != nil {
					fmt.Println("Delete Error", err)
					return
				}
			}

			received <- items
		}
	}
}

func Count(expected int, received chan int, wg *sync.WaitGroup) {
	remaining := expected
	for {
		select {
		case items := <-received:
			remaining = remaining - items
			if remaining <= 0 {
				wg.Done()
				return
			}
		}
	}
}

var body = `
[
  {
    "_id": "5d3b34a29b930f4dc593ed48",
    "index": 0,
    "guid": "85abef1e-06a0-40c5-959e-583769e4ca8c",
    "isActive": true,
    "balance": "$2,451.19",
    "picture": "http://placehold.it/32x32",
    "age": 21,
    "eyeColor": "blue",
    "name": "Cruz Jackson",
    "gender": "male",
    "company": "ZBOO",
    "email": "cruzjackson@zboo.com",
    "phone": "+1 (863) 582-2701",
    "address": "105 Clinton Avenue, Darrtown, Missouri, 1347",
    "about": "Nulla magna veniam pariatur commodo aliquip dolor veniam incididunt labore consectetur consectetur occaecat. Sunt consectetur mollit eu excepteur est amet nulla laborum culpa est laborum exercitation exercitation. Incididunt consectetur nisi adipisicing cillum velit exercitation elit nulla aliquip occaecat tempor est. Aute tempor officia eu mollit minim eiusmod consequat fugiat.\r\n",
    "registered": "2014-03-02T06:49:12 -01:00",
    "latitude": 80.093713,
    "longitude": 89.119776,
    "tags": [
      "exercitation",
      "duis",
      "velit",
      "eu",
      "voluptate",
      "sit",
      "quis"
    ],
    "friends": [
      {
        "id": 0,
        "name": "Franklin Weaver"
      },
      {
        "id": 1,
        "name": "Joyce Hopkins"
      },
      {
        "id": 2,
        "name": "Cardenas West"
      }
    ],
    "greeting": "Hello, Cruz Jackson! You have 8 unread messages.",
    "favoriteFruit": "strawberry"
  },
  {
    "_id": "5d3b34a2fb49143695b7e214",
    "index": 1,
    "guid": "64b423dc-317d-417f-b238-74611999445e",
    "isActive": true,
    "balance": "$1,415.92",
    "picture": "http://placehold.it/32x32",
    "age": 40,
    "eyeColor": "green",
    "name": "Justice Reilly",
    "gender": "male",
    "company": "COMTRAIL",
    "email": "justicereilly@comtrail.com",
    "phone": "+1 (853) 514-3011",
    "address": "720 Ferris Street, Sidman, Pennsylvania, 6510",
    "about": "Ea irure deserunt ad cupidatat fugiat anim nostrud id magna magna ex laboris Lorem. Cillum aute irure aute Lorem dolor aliqua. Consequat excepteur occaecat ad sit laborum elit enim. Lorem officia cillum nisi non minim elit id laboris proident velit ullamco.\r\n",
    "registered": "2017-07-17T01:35:54 -02:00",
    "latitude": -35.808757,
    "longitude": -125.854665,
    "tags": [
      "ad",
      "mollit",
      "aute",
      "est",
      "mollit",
      "nisi",
      "exercitation"
    ],
    "friends": [
      {
        "id": 0,
        "name": "Rodriguez Henderson"
      },
      {
        "id": 1,
        "name": "Ruiz Berry"
      },
      {
        "id": 2,
        "name": "Mccoy Witt"
      }
    ],
    "greeting": "Hello, Justice Reilly! You have 5 unread messages.",
    "favoriteFruit": "apple"
  },
  {
    "_id": "5d3b34a260c0fd71c70030b9",
    "index": 2,
    "guid": "536d6a77-c9cf-4c41-bb36-4ca237a605b8",
    "isActive": false,
    "balance": "$1,201.09",
    "picture": "http://placehold.it/32x32",
    "age": 37,
    "eyeColor": "blue",
    "name": "Dale Conway",
    "gender": "female",
    "company": "UNIA",
    "email": "daleconway@unia.com",
    "phone": "+1 (902) 501-3348",
    "address": "423 Bijou Avenue, Maplewood, Connecticut, 7679",
    "about": "Mollit nostrud tempor non elit sit consectetur irure sint nulla non sint sit nostrud. Non veniam duis sint laborum magna duis adipisicing aliqua anim. Proident aute est dolor officia velit laboris minim adipisicing aliquip ipsum. Aute dolor est adipisicing sit reprehenderit mollit laborum eu excepteur. Minim est ipsum elit elit quis mollit id eu sunt est consequat fugiat dolore laborum.\r\n",
    "registered": "2017-01-08T03:04:53 -01:00",
    "latitude": -84.562239,
    "longitude": 91.917599,
    "tags": [
      "exercitation",
      "excepteur",
      "ea",
      "dolor",
      "Lorem",
      "deserunt",
      "consectetur"
    ],
    "friends": [
      {
        "id": 0,
        "name": "Huff Rodriguez"
      },
      {
        "id": 1,
        "name": "Santana Guthrie"
      },
      {
        "id": 2,
        "name": "Crosby Snider"
      }
    ],
    "greeting": "Hello, Dale Conway! You have 10 unread messages.",
    "favoriteFruit": "banana"
  },
  {
    "_id": "5d3b34a24fa3d2511b0a1266",
    "index": 3,
    "guid": "a24a8927-7118-443b-b2ea-e0c69e21c71a",
    "isActive": true,
    "balance": "$3,845.24",
    "picture": "http://placehold.it/32x32",
    "age": 20,
    "eyeColor": "blue",
    "name": "Janna Yates",
    "gender": "female",
    "company": "BESTO",
    "email": "jannayates@besto.com",
    "phone": "+1 (890) 409-3943",
    "address": "859 Front Street, Walker, Arkansas, 9797",
    "about": "Exercitation tempor culpa voluptate Lorem proident occaecat amet eu aliqua labore nulla. Nulla magna cillum enim incididunt veniam eiusmod qui nulla laborum. Id elit aliquip sit aute magna et qui nisi in voluptate. Elit Lorem veniam amet elit qui. Laborum fugiat Lorem incididunt ad officia pariatur enim. Culpa laboris voluptate nisi occaecat eu eiusmod excepteur esse anim sint. Occaecat adipisicing et consequat id laboris culpa dolor ea tempor sint sint occaecat commodo aute.\r\n",
    "registered": "2017-04-22T04:33:18 -02:00",
    "latitude": -15.672343,
    "longitude": 176.336019,
    "tags": [
      "laborum",
      "Lorem",
      "enim",
      "eiusmod",
      "ad",
      "et",
      "esse"
    ],
    "friends": [
      {
        "id": 0,
        "name": "Darla Kaufman"
      },
      {
        "id": 1,
        "name": "Glover Boyle"
      },
      {
        "id": 2,
        "name": "Lakeisha Thomas"
      }
    ],
    "greeting": "Hello, Janna Yates! You have 3 unread messages.",
    "favoriteFruit": "strawberry"
  },
  {
    "_id": "5d3b34a2adbb3349b6658c6b",
    "index": 4,
    "guid": "732434e6-2d0d-402f-814f-a3f22dcd5c20",
    "isActive": true,
    "balance": "$1,527.40",
    "picture": "http://placehold.it/32x32",
    "age": 40,
    "eyeColor": "brown",
    "name": "Laurie Wilkerson",
    "gender": "female",
    "company": "QABOOS",
    "email": "lauriewilkerson@qaboos.com",
    "phone": "+1 (860) 560-3743",
    "address": "369 Montana Place, Whitehaven, Colorado, 6304",
    "about": "Qui officia dolore fugiat laborum qui minim sint magna minim sint eu. Esse sint non ullamco aliquip fugiat in eiusmod consequat nulla deserunt. Exercitation nisi anim eu nostrud. Quis consectetur fugiat non labore sit et ipsum irure ullamco magna ipsum eu. Adipisicing cupidatat laboris voluptate aliqua et ipsum minim nulla nisi consequat minim consequat laboris quis. Magna dolor cupidatat cupidatat aute.\r\n",
    "registered": "2017-07-25T01:10:15 -02:00",
    "latitude": -35.09454,
    "longitude": 116.666404,
    "tags": [
      "est",
      "esse",
      "sunt",
      "officia",
      "magna",
      "eiusmod",
      "sit"
    ],
    "friends": [
      {
        "id": 0,
        "name": "Rachel Guzman"
      },
      {
        "id": 1,
        "name": "Tommie Watson"
      },
      {
        "id": 2,
        "name": "Montoya Rocha"
      }
    ],
    "greeting": "Hello, Laurie Wilkerson! You have 9 unread messages.",
    "favoriteFruit": "banana"
  },
  {
    "_id": "5d3b34a222c7b1a695644da0",
    "index": 5,
    "guid": "9feacef3-3f77-48ea-8a25-ddf3b657c167",
    "isActive": true,
    "balance": "$2,847.81",
    "picture": "http://placehold.it/32x32",
    "age": 34,
    "eyeColor": "brown",
    "name": "Helena Gamble",
    "gender": "female",
    "company": "VANTAGE",
    "email": "helenagamble@vantage.com",
    "phone": "+1 (939) 545-2057",
    "address": "711 Clove Road, Boling, Federated States Of Micronesia, 4709",
    "about": "Aliquip do occaecat cillum sint velit voluptate. Et sit consequat consectetur aute in labore dolore nostrud dolor. Proident deserunt nulla veniam veniam. Laborum laborum dolor sint in in nostrud. Dolore proident cillum sit eu duis pariatur ut deserunt. Elit nisi commodo proident cupidatat. Nulla qui aliqua esse deserunt occaecat sit.\r\n",
    "registered": "2019-07-03T05:18:09 -02:00",
    "latitude": -54.924922,
    "longitude": -75.710242,
    "tags": [
      "dolore",
      "ut",
      "tempor",
      "dolor",
      "exercitation",
      "dolore",
      "proident"
    ],
    "friends": [
      {
        "id": 0,
        "name": "Deleon Murphy"
      },
      {
        "id": 1,
        "name": "Madeleine Banks"
      },
      {
        "id": 2,
        "name": "Lindsay Hunt"
      }
    ],
    "greeting": "Hello, Helena Gamble! You have 9 unread messages.",
    "favoriteFruit": "banana"
  }
]

	`
