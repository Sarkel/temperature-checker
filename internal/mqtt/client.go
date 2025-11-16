package mqtt

import (
	"context"
	"log/slog"
	"temperature-checker/internal/config"
)

type Dependencies struct {
	Logger *slog.Logger
	Config *config.MQTTBrokerConfig
}

type Client interface {
	Publish(topic string, payload []string) error
	Subscribe(ctx context.Context, topic string, handler MessageHandler) error
	Unsubscribe(topic string) error
	Close()
}

type Message struct {
	Topic   string
	Payload []string
}

type MessageHandler func(context.Context, Client, Message)
