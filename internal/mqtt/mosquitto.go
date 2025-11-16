package mqtt

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand"
	"strconv"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MosquittoClient struct {
	c         mqtt.Client
	l         *slog.Logger
	separator string
}

func NewMosquittoClient(deps Dependencies) (*MosquittoClient, error) {
	opts := mqtt.NewClientOptions().
		AddBroker(deps.Config.URL).
		SetClientID(deps.Config.ClientID + "-" + strconv.Itoa(rand.Intn(1000))).
		SetUsername(deps.Config.Username).
		SetPassword(deps.Config.Password).
		SetKeepAlive(30 * time.Second).
		SetPingTimeout(10 * time.Second)

	c := mqtt.NewClient(opts)
	token := c.Connect()
	if token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	return &MosquittoClient{
		c:         c,
		l:         deps.Logger,
		separator: deps.Config.PayloadSeparator,
	}, nil
}

func (c *MosquittoClient) Publish(topic string, payload []MessagePayload) error {
	token := c.c.Publish(topic, 0, false, decode(payload, c.separator))
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt publish: %w", token.Error())
	}
	return nil
}

func (c *MosquittoClient) Subscribe(ctx context.Context, topic string, handler MessageHandler) error {
	token := c.c.Subscribe(topic, 0, func(ic mqtt.Client, msg mqtt.Message) {
		handler(ctx, c, Message{
			Topic:   msg.Topic(),
			Payload: encode(string(msg.Payload()), c.separator),
		})
	})
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt subscribe: %w", token.Error())
	}

	return nil
}

func (c *MosquittoClient) Unsubscribe(topic string) error {
	token := c.c.Unsubscribe(topic)
	if token.Wait() && token.Error() != nil {
		return fmt.Errorf("mqtt unsubscribe: %w", token.Error())
	}
	return nil
}

func (c *MosquittoClient) Close() {
	c.c.Disconnect(250)
}
