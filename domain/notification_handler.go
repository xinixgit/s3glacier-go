package domain

import "time"

const DefaultWaitInterval = 10 * time.Minute // default wait 10 minutes before querying again

type NotificationHandler interface {
	// poll notification messages to detect a completed job
	//
	// it returns
	// - the message body if it receives one
	// - a boolean flag whether it receives non-empty message
	// - any error encountered
	Poll(queueName *string) (*string, bool, error)
	PollWithInterval(queueName *string, interval time.Duration) (*string, error)
}
