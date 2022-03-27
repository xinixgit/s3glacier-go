package util

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type JobNotificationHandler struct {
	QueueName     *string
	Svc           *sqs.SQS
	SleepInterval time.Duration
}

func (h JobNotificationHandler) GetNotification() (*string, error) {
	urlResult, err := h.Svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: h.QueueName,
	})

	if err != nil {
		fmt.Println("Failed to construct the url from queueName: ", *h.QueueName)
		return nil, err
	}

	input := &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		QueueUrl:            urlResult.QueueUrl,
		MaxNumberOfMessages: aws.Int64(1),
	}

	for {
		res, err := h.Svc.ReceiveMessage(input)
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] %d messages received.\n", GetDBNowStr(), len(res.Messages))

		if len(res.Messages) > 0 {
			msg := res.Messages[0]
			fmt.Println("Message published at: ", *msg.Attributes[sqs.MessageSystemAttributeNameSentTimestamp])
			return msg.Body, nil
		}

		time.Sleep(h.SleepInterval)
	}
}
