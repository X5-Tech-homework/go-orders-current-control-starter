package order

import "time"

type Status string

const (
	StatusNew        Status = "new"
	StatusProcessing Status = "processing"
	StatusCompleted  Status = "completed"
	StatusCancelled  Status = "cancelled"
)

type Item struct {
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
	Price    int64  `json:"price"`
}

type StatusChange struct {
	Status    Status    `json:"status"`
	ChangedAt time.Time `json:"changed_at"`
}

type Order struct {
	ID         int64          `json:"id"`
	Items      []Item         `json:"items"`
	TotalPrice int64          `json:"total_price"`
	Status     Status         `json:"status"`
	CreatedAt  time.Time      `json:"created_at"`
	History    []StatusChange `json:"-"`
}

type CreateOrderInput struct {
	Items []Item `json:"items"`
}

type UpdateStatusInput struct {
	Status Status `json:"status"`
}
