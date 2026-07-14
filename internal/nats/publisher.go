package nats

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
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

func (p *Publisher) Publish(event, action string, data any) error {
	subject := fmt.Sprintf(subjectPattern, p.workspace, event, action)
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	if err := p.nc.Publish(subject, payload); err != nil {
		return err
	}
	log.Printf("published %s (%d bytes)", subject, len(payload))
	return nil
}

func (p *Publisher) Close() {
	p.nc.Drain()
}
