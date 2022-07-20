package adapter

import (
	"fmt"
	"s3glacier-go/domain"
	"s3glacier-go/util"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type AWSSQS struct {
	svc *sqs.SQS
}

func NewJobNotificationHandler(svc *sqs.SQS) domain.JobNotificationHandler {
	return &AWSSQS{
		svc: svc,
	}
}

func (h *AWSSQS) GetNotification(queueName *string, waitInterval time.Duration) (*string, error) {
	urlResult, err := h.svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: queueName,
	})

	if err != nil {
		fmt.Println("Failed to construct the url from queueName: ", *queueName)
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
		res, err := h.svc.ReceiveMessage(input)
		if err != nil {
			return nil, err
		}

		fmt.Printf("[%s] %d messages received.\n", util.GetDBNowStr(), len(res.Messages))

		if len(res.Messages) > 0 {
			msg := res.Messages[0]
			fmt.Println("Message published at: ", *msg.Attributes[sqs.MessageSystemAttributeNameSentTimestamp])
			return msg.Body, nil
		}

		time.Sleep(waitInterval)
	}
}
