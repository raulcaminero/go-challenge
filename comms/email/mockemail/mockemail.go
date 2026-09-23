package mockemail

import (
	"encoding/json"

	"github.com/gloveboxhq/glovebox-go-code-challenge/comms/email"
)

type Client struct {
	sendLogs SendLogs
}

func NewClient() *Client {
	return &Client{
		sendLogs: SendLogs{},
	}
}

// Compile-time guarantee that the mock satisfies email.MailProvider.
var _ email.MailProvider = (*Client)(nil)

// Send records one log entry per To recipient. Each entry carries the full CC
// list so tests can assert who was copied on the message that recipient saw.
func (c *Client) Send(to, cc []string, message json.RawMessage, tplID email.TplID) error {

	for _, v := range to {
		c.sendLogs = append(c.sendLogs, SendLog{
			to:      v,
			cc:      append([]string(nil), cc...),
			message: message,
			tplID:   tplID,
		})
	}

	return nil
}

func (c *Client) SendLogs() SendLogs {
	return c.sendLogs
}

func (c *Client) FlushSendLogs() {
	c.sendLogs = SendLogs{}
}
