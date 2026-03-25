package broker

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/moon-eye/velune/shared/contracts"
	constx "github.com/moon-eye/velune/shared/constx"
	errs "github.com/moon-eye/velune/shared/errors"
)

type RabbitMQ struct {
	conn       *amqp.Connection
	ch         *amqp.Channel
	exchange   string
	dlx        string
	queue      string
	dlq        string
	routingKey string
	dlqKey     string
}

func New(url, exchange, queue, routingKey, dlx, dlq, dlqRoutingKey string) (*RabbitMQ, error) {
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
	args := amqp.Table{}
	if dlx != "" {
		args["x-dead-letter-exchange"] = dlx
		if dlqRoutingKey != "" {
			args["x-dead-letter-routing-key"] = dlqRoutingKey
		}
	}
	q, err := ch.QueueDeclare(queue, true, false, false, false, args)
	if err != nil {
		return nil, err
	}
	if err := ch.QueueBind(q.Name, routingKey, exchange, false, nil); err != nil {
		return nil, err
	}
	if dlq != "" && dlx != "" {
		dlqQ, err := ch.QueueDeclare(dlq, true, false, false, false, nil)
		if err != nil {
			return nil, err
		}
		if err := ch.QueueBind(dlqQ.Name, dlqRoutingKey, dlx, false, nil); err != nil {
			return nil, err
		}
	}
	return &RabbitMQ{
		conn:       conn,
		ch:         ch,
		exchange:   exchange,
		dlx:        dlx,
		queue:      q.Name,
		dlq:        dlq,
		routingKey: routingKey,
		dlqKey:     dlqRoutingKey,
	}, nil
}

func (r *RabbitMQ) Close() error {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

func (r *RabbitMQ) Publish(ctx context.Context, env contracts.EventEnvelope) error {
	body, err := json.Marshal(env)
	if err != nil {
		return err
	}
	return r.ch.PublishWithContext(ctx, r.exchange, r.routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   env.EventID.String(),
		Timestamp:   time.Now().UTC(),
		Body:        body,
	})
}

func (r *RabbitMQ) Consume(ctx context.Context, handler func(context.Context, contracts.EventEnvelope) error) error {
	msgs, err := r.ch.Consume(r.queue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return nil
			}
			var env contracts.EventEnvelope
			if err := json.Unmarshal(msg.Body, &env); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			if err := handler(ctx, env); err != nil {
				_ = msg.Nack(false, false)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}
