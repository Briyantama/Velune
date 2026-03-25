package otelx

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.opentelemetry.io/otel"
)

// AMQPHeaderCarrier adapts amqp.Table for W3C trace context inject/extract.
type AMQPHeaderCarrier amqp.Table

func (c AMQPHeaderCarrier) Get(key string) string {
	if c == nil {
		return ""
	}
	v, ok := c[key]
	if !ok {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case []byte:
		return string(t)
	default:
		return fmt.Sprint(t)
	}
}

func (c AMQPHeaderCarrier) Set(key, value string) {
	if c == nil {
		return
	}
	c[key] = value
}

func (c AMQPHeaderCarrier) Keys() []string {
	if c == nil {
		return nil
	}
	out := make([]string, 0, len(c))
	for k := range c {
		out = append(out, k)
	}
	return out
}

// InjectAMQP merges trace context from ctx into headers (mutates headers map).
func InjectAMQP(ctx context.Context, headers amqp.Table) {
	if headers == nil {
		return
	}
	otel.GetTextMapPropagator().Inject(ctx, AMQPHeaderCarrier(headers))
}

// ExtractAMQP returns a context that carries remote span from message headers, if any.
func ExtractAMQP(ctx context.Context, headers amqp.Table) context.Context {
	if len(headers) == 0 {
		return ctx
	}
	return otel.GetTextMapPropagator().Extract(ctx, AMQPHeaderCarrier(headers))
}
