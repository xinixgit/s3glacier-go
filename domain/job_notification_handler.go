package domain

import "time"

const DefaultWaitInterval = 10 * time.Minute // default wait 10 minutes before querying again

type JobNotificationHandler interface {
	// Get a notification message when a job is completed, and return that message
	GetNotification(queueName *string, waitInterval time.Duration) (*string, error)
}
