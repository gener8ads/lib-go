package pubsub

import (
	"context"
	"log"
	"os"
	"time"

	"cloud.google.com/go/pubsub"
)

var timeout = time.Second * 5

// Client creates a new pubsub client
func Client() *pubsub.Client {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	projectID := os.Getenv("GCP_PROJECT")

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

// Topic creates an instantiated pubsub topic
func Topic(topicID string) (*pubsub.Topic, error) {
	client := Client()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	topic := client.Topic(topicID)
	topicExists, err := topic.Exists(ctx)

	if err != nil {
		return topic, err
	}
	if !topicExists {
		if _, createErr := client.CreateTopic(ctx, topicID); createErr != nil {
			return topic, createErr
		}
	}

	return topic, nil
}

// Subscription creates a instantiated pubsub subscription
func Subscription(subID string, topic *pubsub.Topic) (*pubsub.Subscription, error) {
	client := Client()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	subscription := client.Subscription(subID)
	subExists, err := subscription.Exists(ctx)

	if err != nil {
		return subscription, err
	}
	if !subExists {
		if _, createErr := client.CreateSubscription(ctx, subID, pubsub.SubscriptionConfig{Topic: topic}); createErr != nil {
			return subscription, createErr
		}
	}

	return subscription, nil
}
