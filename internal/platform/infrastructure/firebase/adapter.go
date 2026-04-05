package firebase

import "context"

type TaskAdapter struct {
	fb *Firebase
}

func NewTaskAdapter(fb *Firebase) *TaskAdapter {
	return &TaskAdapter{fb: fb}
}

func (a *TaskAdapter) Send(ctx context.Context, token, fcmType, title, body string, data map[string]string) error {
	return a.fb.SendNotification(ctx, token, fcmType, Content{Title: title, Body: body}, data)
}

func (a *TaskAdapter) SendMulti(ctx context.Context, tokens []string, fcmType, title, body string, data map[string]string) error {
	return a.fb.SendMultiNotification(ctx, tokens, fcmType, Content{Title: title, Body: body}, data)
}
