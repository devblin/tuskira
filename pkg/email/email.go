package email

// Message is the data needed to send an email.
type Message struct {
	From    string
	To      string
	Subject string
	Body    string
}

// Sender is the interface that all email provider implementations must satisfy.
type Sender interface {
	Send(msg Message) error
}
