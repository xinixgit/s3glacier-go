package domain

import "time"

type JobNotificationHandler interface {
	GetNotification(queueName *string, waitInterval time.Duration) (*string, error)
}
