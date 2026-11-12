package order

import "context"

type Repository interface {
	Create(ctx context.Context, value Order) (Order, error)
	GetByID(ctx context.Context, id int64) (Order, error)
	List(ctx context.Context) ([]Order, error)
	Update(ctx context.Context, value Order) (Order, error)
}
