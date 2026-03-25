package events

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/otelx"
	"github.com/moon-eye/velune/shared/sim"
)

type RabbitPublisher struct {
	conn              *amqp.Connection
	ch                *amqp.Channel
	exchange          string
	dlx               string
	dlqRoutingKey     string
	defaultRoutingKey string
	sim               *sim.Config
}

// NewRabbitPublisher connects to RabbitMQ. sim may be nil or sim.LoadFromEnv(); Publish fail injection never applies to PublishDLQ.
func NewRabbitPublisher(url, exchange, routingKey string, dlx string, dlqRoutingKey string, chaos *sim.Config) (*RabbitPublisher, error) {
	if chaos != nil && chaos.BrokerDown {
		return nil, errs.New("UPSTREAM_UNAVAILABLE", "rabbitmq unavailable (simulated broker down)", constx.StatusBadGateway)
	}
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, errs.New("UPSTREAM_UNAVAILABLE", "rabbitmq unavailable", constx.StatusBadGateway)
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
		sim:               chaos,
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
	headers := amqp.Table{}
	otelx.InjectAMQP(ctx, headers)
	return p.ch.PublishWithContext(ctx, p.dlx, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   env.EventID.String(),
		Timestamp:   time.Now().UTC(),
		Headers:     headers,
		Body:        body,
	})
}

func (p *RabbitPublisher) Publish(ctx context.Context, env contracts.EventEnvelope) error {
	if p.sim != nil && p.sim.SimulatePublishFailure() {
		return sim.ErrSimulatedPublishFailure
	}
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	key := env.EventType
	if key == "" {
		key = p.defaultRoutingKey
	}
	headers := amqp.Table{}
	otelx.InjectAMQP(ctx, headers)
	return p.ch.PublishWithContext(ctx, p.exchange, key, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   env.EventID.String(),
		Timestamp:   time.Now().UTC(),
		Headers:     headers,
		Body:        body,
	})
}
