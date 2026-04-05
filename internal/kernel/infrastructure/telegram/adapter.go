package telegram

type TaskAdapter struct {
	client *Client
}

func NewTaskAdapter(client *Client) *TaskAdapter {
	return &TaskAdapter{client: client}
}

func (a *TaskAdapter) Send(msgType, text string) error {
	return a.client.SendMessage(MessageType(msgType), text)
}
