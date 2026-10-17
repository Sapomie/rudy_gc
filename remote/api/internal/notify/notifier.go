package notify

import "context"

type Message struct {
	Subject string
	Body    string
}

type Notifier interface {
	Notify(ctx context.Context, message Message) error
}
