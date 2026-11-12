package service

import (
	"context"
	"sort"
	"time"

	"example.com/order-service/internal/order"
)

type OrderService struct {
	repository order.Repository
}

func NewOrderService(repository order.Repository) *OrderService {
	return &OrderService{repository: repository}
}

func (s *OrderService) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	if len(input.Items) == 0 {
		return order.Order{}, order.ErrInvalidOrder
	}
	totalPrice := int64(0)
	for _, item := range input.Items {
		if item.Name == "" || item.Quantity <= 0 || item.Price <= 0 {
			return order.Order{}, order.ErrInvalidOrder
		}
		totalPrice += int64(item.Quantity) * item.Price
	}
	now := time.Now()
	value := order.Order{
		Items:      input.Items,
		TotalPrice: totalPrice,
		Status:     order.StatusNew,
		CreatedAt:  now,
		History: []order.StatusChange{
			{Status: order.StatusNew, ChangedAt: now},
		},
	}

	return s.repository.Create(ctx, value)
}

func (s *OrderService) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.repository.GetByID(ctx, id)
}

func (s *OrderService) List(ctx context.Context) ([]order.Order, error) {
	return s.repository.List(ctx)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	switch status {
	case order.StatusNew, order.StatusProcessing, order.StatusCompleted, order.StatusCancelled:
	default:
		return order.Order{}, order.ErrInvalidStatus
	}

	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return order.Order{}, err
	}

	allowed := (value.Status == order.StatusNew && (status == order.StatusProcessing || status == order.StatusCancelled)) ||
		(value.Status == order.StatusProcessing && (status == order.StatusCompleted || status == order.StatusCancelled))
	if !allowed {
		return order.Order{}, order.ErrInvalidStatusTransition
	}

	value.Status = status
	// Копируем историю, чтобы не менять данные хранилища до успешного Update.
	value.History = append(append([]order.StatusChange(nil), value.History...), order.StatusChange{
		Status:    status,
		ChangedAt: time.Now(),
	})
	return s.repository.Update(ctx, value)
}

func (s *OrderService) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	value, err := s.repository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Сортируем копию, не изменяя историю, полученную из репозитория.
	history := append([]order.StatusChange{}, value.History...)
	sort.SliceStable(history, func(i, j int) bool {
		return history[i].ChangedAt.Before(history[j].ChangedAt)
	})
	return history, nil
}
