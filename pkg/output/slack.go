package output

import (
	"github.com/devopsext/imperva-events/pkg/common"
	"github.com/devopsext/tools/vendors"
	"time"
)

type Slack struct {
	client  *vendors.Slack
	channel string
}

func (s *Slack) Send(e *common.Event) ([]byte, error) {
	return s.client.SendCustomMessage(vendors.SlackMessage{
		Title:   "[" + e.Time.Format(time.RFC822) + "]" + e.Title,
		Message: e.Body,
		Channel: s.channel,
	})
}

func NewSlack(token string, channel string) (*Slack, error) {
	slack, err := vendors.NewSlack(vendors.SlackOptions{
		Timeout:  30,
		Insecure: false,
		Token:    token,
		Channel:  channel,
	})
	if err != nil {
		return nil, err
	}
	return &Slack{
		client:  slack,
		channel: channel,
	}, nil
}
