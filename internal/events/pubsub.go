package events

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"cloud.google.com/go/pubsub"
)

// Publisher publishes document events.
type Publisher interface {
	PublishDocumentUploaded(ctx context.Context, event DocumentUploadedEvent) error
	Close() error
}

// PubSubPublisher publishes document events to Google Pub/Sub.
type PubSubPublisher struct {
	client *pubsub.Client
	topic  *pubsub.Topic
}

// NewPubSubPublisher creates a new Pub/Sub publisher.
func NewPubSubPublisher(ctx context.Context, projectID, topicName string) (*PubSubPublisher, error) {
	if projectID == "" {
		return nil, fmt.Errorf("pubsub project ID is required")
	}
	if topicName == "" {
		return nil, fmt.Errorf("pubsub topic name is required")
	}

	topicID := NormalizeResourceID(topicName)
	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("create pubsub client: %w", err)
	}

	return &PubSubPublisher{
		client: client,
		topic:  client.Topic(topicID),
	}, nil
}

// PublishDocumentUploaded publishes a document uploaded event.
func (p *PubSubPublisher) PublishDocumentUploaded(ctx context.Context, event DocumentUploadedEvent) error {
	if p == nil || p.topic == nil {
		return fmt.Errorf("pubsub publisher not configured")
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	res := p.topic.Publish(ctx, &pubsub.Message{
		Data: data,
		Attributes: map[string]string{
			"type": event.Type,
		},
	})

	if _, err := res.Get(ctx); err != nil {
		return fmt.Errorf("publish event: %w", err)
	}

	return nil
}

// Close closes the publisher client.
func (p *PubSubPublisher) Close() error {
	if p == nil {
		return nil
	}
	if p.topic != nil {
		p.topic.Stop()
	}
	if p.client != nil {
		return p.client.Close()
	}
	return nil
}

// NormalizeResourceID extracts the resource ID from a full Pub/Sub resource name.
func NormalizeResourceID(name string) string {
	if name == "" {
		return ""
	}
	if !strings.Contains(name, "/") {
		return name
	}
	parts := strings.Split(name, "/")
	return parts[len(parts)-1]
}
