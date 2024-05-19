package adapter

import (
	"fmt"
	"s3glacier-go/domain"
	"s3glacier-go/util"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type sqsNotificationHandler struct {
	sqsSvc *sqs.SQS
}

func NewNotificationHandler(sqsSvc *sqs.SQS) domain.NotificationHandler {
	return &sqsNotificationHandler{
		sqsSvc: sqsSvc,
	}
}

func (h *sqsNotificationHandler) PollWithInterval(queueName *string, interval time.Duration) (*string, error) {
	var res *string
	var done bool

	for !done {
		var err error
		res, done, err = h.Poll(queueName)
		if err != nil {
			return nil, err
		}

		if done {
			return res, nil
		}

		time.Sleep(interval)
	}

	return nil, fmt.Errorf("unexpected ending of poll loop")
}

func (h *sqsNotificationHandler) Poll(queueName *string) (*string, bool, error) {
	getQueueUrlInput := sqs.GetQueueUrlInput{
		QueueName: queueName,
	}

	urlResult, err := h.sqsSvc.GetQueueUrl(&getQueueUrlInput)
	if err != nil {
		fmt.Println("Failed to construct the url from queueName: ", *queueName)
		return nil, false, err
	}

	receiveMessageInput := &sqs.ReceiveMessageInput{
		AttributeNames: []*string{
			aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
		},
		QueueUrl:            urlResult.QueueUrl,
		MaxNumberOfMessages: aws.Int64(1),
	}

	res, err := h.sqsSvc.ReceiveMessage(receiveMessageInput)
	if err != nil {
		return nil, false, err
	}

	fmt.Printf("[%s] %d messages received.\n", util.GetDBNowStr(), len(res.Messages))

	if len(res.Messages) > 0 {
		msg := res.Messages[0]
		fmt.Println("Message published at: ", *msg.Attributes[sqs.MessageSystemAttributeNameSentTimestamp])
		return msg.Body, true, nil
	}

	return nil, false, nil
}
