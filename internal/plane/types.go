package plane

type Project struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
	Network     int    `json:"network"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Issue struct {
	ID          string   `json:"id"`
	Sequence    int      `json:"sequence_id"`
	Name        string   `json:"name"`
	Description string   `json:"description_stripped"`
	State       string   `json:"state"`
	Priority    string   `json:"priority"`
	Assignees   []string `json:"assignees"`
	Labels      []string `json:"label_ids"`
	ProjectID   string   `json:"project"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

type ListResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

// WebhookEvent is the payload Plane POSTs to the webhook endpoint.
type WebhookEvent struct {
	Event     string         `json:"event"`
	Action    string         `json:"action"` // created | updated | deleted
	Actor     string         `json:"actor"`
	Workspace string         `json:"workspace"`
	Data      map[string]any `json:"data"`
}
