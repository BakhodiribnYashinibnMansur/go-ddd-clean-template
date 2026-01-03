package telegram

type MessageType string

const (
	Error MessageType = "error"
	Info  MessageType = "info"
)
