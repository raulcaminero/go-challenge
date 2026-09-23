package main

import (
	"context"
	"fmt"

	"github.com/gloveboxhq/glovebox-go-code-challenge/notifier/channels/email"
)

type Producer struct {
	email         email.MailProvider
	topicBuilders map[string]TopicRequestBuilder
}

type ProducerProvider interface {
	NotifyTopic(ctx context.Context, topic string, input any) error
	Notify(ctx context.Context, req Request) error
}

type TopicRequestBuilder interface {
	Topic() string
	BuildRequest(ctx context.Context, input any) (Request, error)
}

type Request struct {
	Topic      string
	Recipients []string
	Template   email.TplID
	Vars       map[string]any
}

func NewProducer(emailProvider email.MailProvider) *Producer {
	p := &Producer{
		email:         emailProvider,
		topicBuilders: map[string]TopicRequestBuilder{},
	}
	p.register(
		docUploadTopicBuilder{},
		otpLoginTopicBuilder{},
		policyRenewalTopicBuilder{},
	)
	return p
}

// register indexes builders by their own Topic() so a topic can never be
// registered under a mismatched key.
func (p *Producer) register(builders ...TopicRequestBuilder) {
	for _, builder := range builders {
		p.topicBuilders[builder.Topic()] = builder
	}
}

func (p *Producer) NotifyTopic(ctx context.Context, topic string, input any) error {
	builder, ok := p.topicBuilders[topic]
	if !ok {
		return fmt.Errorf("topic builder not registered: %s", topic)
	}
	req, err := builder.BuildRequest(ctx, input)
	if err != nil {
		return err
	}
	return p.Notify(ctx, req)
}

func (p *Producer) Notify(_ context.Context, req Request) error {
	if err := email.RequireRecipient(req.Recipients); err != nil {
		return fmt.Errorf("notification %q: %w", req.Topic, err)
	}
	return p.email.Send(req.Recipients, req.Template, req.Vars)
}
