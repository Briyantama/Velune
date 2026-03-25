package events

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/moon-eye/velune/shared/contracts"
	"github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
)

type RabbitPublisher struct {
	conn              *amqp.Connection
	ch                *amqp.Channel
	exchange          string
	dlx               string
	dlqRoutingKey     string
	defaultRoutingKey string
}

func NewRabbitPublisher(url, exchange, routingKey string, dlx string, dlqRoutingKey string) (*RabbitPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, errs.New("UPSTREAM_UNAVAILABLE", "rabbitmq unavailable",constx.StatusBadGateway)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(exchange, "topic", true, false, false, false, nil); err != nil {
		return nil, err
	}
	if dlx != "" {
		if err := ch.ExchangeDeclare(dlx, "topic", true, false, false, false, nil); err != nil {
			return nil, err
		}
	}
	return &RabbitPublisher{
		conn:              conn,
		ch:                ch,
		exchange:          exchange,
		dlx:               dlx,
		dlqRoutingKey:     dlqRoutingKey,
		defaultRoutingKey: routingKey,
	}, nil
}

func (p *RabbitPublisher) Close() error {
	if p.ch != nil {
		_ = p.ch.Close()
	}
	if p.conn != nil {
		return p.conn.Close()
	}
	return nil
}

func (p *RabbitPublisher) PublishDLQ(ctx context.Context, env contracts.EventEnvelope) error {
	if p.dlx == "" {
		return p.Publish(ctx, env)
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	key := p.dlqRoutingKey
	if key == "" {
		key = env.EventType
	}
	return p.ch.PublishWithContext(ctx, p.dlx, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   env.EventID.String(),
		Timestamp:   time.Now().UTC(),
		Body:        body,
	})
}

func (p *RabbitPublisher) Publish(ctx context.Context, env contracts.EventEnvelope) error {
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	key := env.EventType
	if key == "" {
		key = p.defaultRoutingKey
	}
	return p.ch.PublishWithContext(ctx, p.exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   env.EventID.String(),
		Timestamp:   time.Now().UTC(),
		Body:        body,
	})
}
