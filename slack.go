// Package slack is a Slack channel for togo notifications (Incoming Webhook).
// Set SLACK_WEBHOOK_URL; PushNotification title/body become a Slack message.
package slack

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/togo-framework/notifications"
	"github.com/togo-framework/togo"
)

func init() {
	notifications.RegisterChannel("slack", func(k *togo.Kernel) notifications.Channel {
		return &channel{webhook: os.Getenv("SLACK_WEBHOOK_URL"), client: &http.Client{Timeout: 15 * time.Second}}
	})
}

type channel struct {
	webhook string
	client  *http.Client
}

func (c *channel) Send(ctx context.Context, to notifications.Notifiable, n notifications.Notification) error {
	pn, ok := n.(notifications.PushNotification)
	if !ok || c.webhook == "" {
		return nil
	}
	m := pn.ToPush(to)
	body, _ := json.Marshal(map[string]any{
		"text": "*" + m.Title + "*\n" + m.Body,
		"blocks": []map[string]any{
			{"type": "header", "text": map[string]any{"type": "plain_text", "text": m.Title}},
			{"type": "section", "text": map[string]any{"type": "mrkdwn", "text": m.Body}},
		},
	})
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, c.webhook, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("slack: %s", resp.Status)
	}
	return nil
}
