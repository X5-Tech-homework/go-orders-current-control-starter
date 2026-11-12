package memory

import (
	"context"
	"sync"

	"example.com/order-service/internal/order"
)

type OrderRepository struct {
	mu     sync.RWMutex
	items  []order.Order
	nextID int64
}

func NewOrderRepository() *OrderRepository {
	return &OrderRepository{
		items:  make([]order.Order, 0),
		nextID: 1,
	}
}

// cloneOrder отделяет срезы заказа от данных вызывающего кода.
func cloneOrder(value order.Order) order.Order {
	if value.Items != nil {
		items := make([]order.Item, len(value.Items))
		copy(items, value.Items)
		value.Items = items
	}
	if value.History != nil {
		history := make([]order.StatusChange, len(value.History))
		copy(history, value.History)
		value.History = history
	}
	return value
}

func (r *OrderRepository) Create(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	value.ID = r.nextID
	r.nextID++
	r.items = append(r.items, cloneOrder(value))
	return cloneOrder(value), nil
}

func (r *OrderRepository) GetByID(_ context.Context, id int64) (order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, value := range r.items {
		if value.ID == id {
			return cloneOrder(value), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}

func (r *OrderRepository) List(_ context.Context) ([]order.Order, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]order.Order, len(r.items))
	for i, value := range r.items {
		result[i] = cloneOrder(value)
	}
	return result, nil
}

func (r *OrderRepository) Update(_ context.Context, value order.Order) (order.Order, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, stored := range r.items {
		if value.ID == stored.ID {
			r.items[i] = cloneOrder(value)
			return cloneOrder(value), nil
		}
	}
	return order.Order{}, order.ErrOrderNotFound
}
