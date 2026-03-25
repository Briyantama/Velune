package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"

	constx "github.com/moon-eye/velune/shared/constx"
	"github.com/moon-eye/velune/shared/contracts"
	errs "github.com/moon-eye/velune/shared/errors"
	"github.com/moon-eye/velune/shared/otelx"
	"github.com/moon-eye/velune/shared/sim"
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
	sim        *sim.Config
	log        *zap.Logger
}

func New(url, exchange, queue, routingKey, dlx, dlq, dlqRoutingKey string, chaos *sim.Config, log *zap.Logger) (*RabbitMQ, error) {
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
		sim:        chaos,
		log:        log,
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
	headers := amqp.Table{}
	otelx.InjectAMQP(ctx, headers)
	return r.ch.PublishWithContext(ctx, r.exchange, r.routingKey, false, false, amqp.Publishing{
		ContentType: "application/json",
		MessageId:   env.EventID.String(),
		Timestamp:   time.Now().UTC(),
		Headers:     headers,
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
			msgCtx := otelx.ExtractAMQP(ctx, msg.Headers)
			herr := func() (err error) {
				defer func() {
					if rec := recover(); rec != nil {
						if r.log != nil {
							r.log.Error("consumer_panic_simulated",
								zap.Any("recover", rec),
								zap.String("event_id", env.EventID.String()),
								zap.String("event_type", env.EventType),
							)
						}
						err = fmt.Errorf("consumer panic: %v", rec)
					}
				}()
				if r.sim != nil && r.sim.ConsumerPanic {
					panic("SIMULATE_CONSUMER_PANIC")
				}
				return handler(msgCtx, env)
			}()
			if herr != nil {
				_ = msg.Nack(false, false)
				continue
			}
			_ = msg.Ack(false)
		}
	}
}

// RunDLQSnoop starts a background consumer on the DLQ queue that logs envelopes (SIMULATE_DLQ_SNOOP only).
func (r *RabbitMQ) RunDLQSnoop(ctx context.Context, log *zap.Logger) {
	if r.dlq == "" || log == nil {
		return
	}
	ch, err := r.conn.Channel()
	if err != nil {
		log.Error("dlq_snoop_channel_open", zap.Error(err))
		return
	}
	go func() {
		defer ch.Close()
		if _, err := ch.QueueDeclare(r.dlq, true, false, false, false, nil); err != nil {
			log.Error("dlq_snoop_queue_declare", zap.Error(err))
			return
		}
		msgs, err := ch.Consume(r.dlq, "", true, false, false, false, nil)
		if err != nil {
			log.Error("dlq_snoop_consume", zap.Error(err))
			return
		}
		log.Info("dlq_snoop_consumer_started", zap.String("queue", r.dlq))
		for {
			select {
			case <-ctx.Done():
				log.Info("dlq_snoop_consumer_stopped")
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				var env contracts.EventEnvelope
				if err := json.Unmarshal(msg.Body, &env); err != nil {
					log.Warn("dlq_snoop_unmarshal", zap.Error(err))
					continue
				}
				log.Info("dlq_snoop_message",
					zap.String("event_id", env.EventID.String()),
					zap.String("event_type", env.EventType),
					zap.String("rabbit_message_id", msg.MessageId),
				)
			}
		}
	}()
}
