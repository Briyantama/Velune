package dlq

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/moon-eye/velune/shared/contracts"
)

// DLQMessage is a DLQ listing row (broker does not store outbox retry_count on the message).
type DLQMessage struct {
	EventID      string    `json:"event_id"`
	EventType    string    `json:"event_type"`
	Idempotency  string    `json:"idempotency,omitempty"`
	PayloadTrunc string    `json:"payload_truncated"`
	Timestamp    time.Time `json:"timestamp,omitempty"`
}

const maxPreview = 512

// Peek lists up to limit messages from the DLQ without removing them (Nack requeue).
func Peek(ctx context.Context, ch *amqp.Channel, queue string, limit int) ([]DLQMessage, error) {
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	var out []DLQMessage
	for i := 0; i < limit; i++ {
		if err := ctx.Err(); err != nil {
			return out, err
		}
		msg, ok, err := ch.Get(queue, false)
		if err != nil {
			return out, err
		}
		if !ok {
			break
		}
		var env contracts.EventEnvelope
		if err := json.Unmarshal(msg.Body, &env); err != nil {
			_ = msg.Nack(false, true)
			continue
		}
		prev := string(msg.Body)
		if len(prev) > maxPreview {
			prev = prev[:maxPreview] + "..."
		}
		m := DLQMessage{
			EventID:      env.EventID.String(),
			EventType:    env.EventType,
			Idempotency:  env.Idempotency,
			PayloadTrunc: prev,
			Timestamp:    msg.Timestamp,
		}
		out = append(out, m)
		if err := msg.Nack(false, true); err != nil {
			return out, err
		}
	}
	return out, nil
}

// Replay finds a message by event ID (message body or Rabbit message id), republishes via publish, then acks to remove from DLQ.
func Replay(ctx context.Context, ch *amqp.Channel, dlqQueue, eventID string, publish func(context.Context, contracts.EventEnvelope) error) error {
	const maxScan = 10000
	for i := 0; i < maxScan; i++ {
		if err := ctx.Err(); err != nil {
			return err
		}
		msg, ok, err := ch.Get(dlqQueue, false)
		if err != nil {
			return err
		}
		if !ok {
			return fmt.Errorf("event_id not found in DLQ: %s", eventID)
		}
		var env contracts.EventEnvelope
		if err := json.Unmarshal(msg.Body, &env); err != nil {
			_ = msg.Nack(false, true)
			continue
		}
		match := env.EventID.String() == eventID || msg.MessageId == eventID
		if !match {
			_ = msg.Nack(false, true)
			continue
		}
		if err := publish(ctx, env); err != nil {
			_ = msg.Nack(false, true)
			return err
		}
		if err := msg.Ack(false); err != nil {
			return err
		}
		return nil
	}
	return fmt.Errorf("dlq replay scan exceeded limit")
}

// Depth returns approximate message count in the DLQ (RabbitMQ QueueInspect).
func Depth(ch *amqp.Channel, queue string) (int, error) {
	q, err := ch.QueueInspect(queue)
	if err != nil {
		return 0, err
	}
	return q.Messages, nil
}
