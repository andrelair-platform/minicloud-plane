package nats

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Subject pattern: plane.<workspace>.<event>.<action>
// e.g. plane.andrelair.issue.created
const subjectPattern = "plane.%s.%s.%s"

type Publisher struct {
	nc        *nats.Conn
	workspace string
}

func NewPublisher(url, workspace string) (*Publisher, error) {
	nc, err := nats.Connect(url,
		nats.Name("minicloud-plane"),
		nats.MaxReconnects(-1),
	)
	if err != nil {
		return nil, err
	}
	log.Printf("NATS connected: %s", url)
	return &Publisher{nc: nc, workspace: workspace}, nil
}

func (p *Publisher) Publish(ctx context.Context, event, action string, data any) error {
	subject := fmt.Sprintf(subjectPattern, p.workspace, event, action)

	tracer := otel.Tracer("minicloud-plane/nats")
	ctx, span := tracer.Start(ctx, "nats.publish",
		trace.WithSpanKind(trace.SpanKindProducer),
		trace.WithAttributes(
			semconv.MessagingSystemKey.String("nats"),
			semconv.MessagingDestinationName(subject),
			attribute.String("messaging.operation", "publish"),
		),
	)
	defer span.End()

	payload, err := json.Marshal(data)
	if err != nil {
		span.RecordError(err)
		return err
	}
	if err := p.nc.Publish(subject, payload); err != nil {
		span.RecordError(err)
		return err
	}
	span.SetAttributes(attribute.Int("messaging.message.body.size", len(payload)))
	log.Printf("published %s (%d bytes)", subject, len(payload))
	return nil
}

func (p *Publisher) Close() {
	p.nc.Drain()
}
