package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"example.com/order-service/internal/order"
	"example.com/order-service/internal/service"
	"github.com/stretchr/testify/require"
)

// repositoryStub позволяет тестировать сервис без настоящего хранилища
type repositoryStub struct {
	createFunc  func(context.Context, order.Order) (order.Order, error)
	getByIDFunc func(context.Context, int64) (order.Order, error)
	listFunc    func(context.Context) ([]order.Order, error)
	updateFunc  func(context.Context, order.Order) (order.Order, error)
}

func (r *repositoryStub) Create(ctx context.Context, value order.Order) (order.Order, error) {
	return r.createFunc(ctx, value)
}

func (r *repositoryStub) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return r.getByIDFunc(ctx, id)
}

func (r *repositoryStub) List(ctx context.Context) ([]order.Order, error) {
	return r.listFunc(ctx)
}

func (r *repositoryStub) Update(ctx context.Context, value order.Order) (order.Order, error) {
	return r.updateFunc(ctx, value)
}

func TestCreate(t *testing.T) {
	t.Parallel()

	t.Run("успешное создание и расчет стоимости", func(t *testing.T) {
		items := []order.Item{
			{Name: "Keyboard", Quantity: 2, Price: 500},
			{Name: "Mouse", Quantity: 3, Price: 200},
		}
		var saved order.Order
		called := false
		repo := &repositoryStub{
			createFunc: func(_ context.Context, value order.Order) (order.Order, error) {
				called = true
				saved = value
				value.ID = 1
				return value, nil
			},
		}
		svc := service.NewOrderService(repo)

		got, err := svc.Create(context.Background(), order.CreateOrderInput{Items: items})

		require.NoError(t, err)
		require.True(t, called)
		require.Equal(t, items, saved.Items)
		require.Equal(t, int64(1600), saved.TotalPrice)
		require.Equal(t, order.StatusNew, saved.Status)
		require.False(t, saved.CreatedAt.IsZero())
		require.Len(t, saved.History, 1)
		require.Equal(t, order.StatusNew, saved.History[0].Status)
		require.Equal(t, saved.CreatedAt, saved.History[0].ChangedAt)
		require.Equal(t, int64(1), got.ID)
		require.Equal(t, saved.TotalPrice, got.TotalPrice)
	})

	t.Run("ошибка валидации", func(t *testing.T) {
		tests := []struct {
			name  string
			items []order.Item
		}{
			{name: "без позиций"},
			{name: "пустое название", items: []order.Item{{Quantity: 1, Price: 100}}},
			{name: "нулевое количество", items: []order.Item{{Name: "Item", Price: 100}}},
			{name: "отрицательное количество", items: []order.Item{{Name: "Item", Quantity: -1, Price: 100}}},
			{name: "нулевая цена", items: []order.Item{{Name: "Item", Quantity: 1}}},
			{name: "отрицательная цена", items: []order.Item{{Name: "Item", Quantity: 1, Price: -1}}},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				called := false
				repo := &repositoryStub{
					createFunc: func(_ context.Context, value order.Order) (order.Order, error) {
						called = true
						return value, nil
					},
				}
				svc := service.NewOrderService(repo)

				_, err := svc.Create(context.Background(), order.CreateOrderInput{Items: tt.items})

				require.ErrorIs(t, err, order.ErrInvalidOrder)
				require.False(t, called, "репозиторий не должен вызываться при ошибке валидации")
			})
		}
	})

	t.Run("ошибка репозитория", func(t *testing.T) {
		storageErr := errors.New("storage failure")
		repo := &repositoryStub{
			createFunc: func(_ context.Context, _ order.Order) (order.Order, error) {
				return order.Order{}, storageErr
			},
		}
		svc := service.NewOrderService(repo)

		_, err := svc.Create(context.Background(), order.CreateOrderInput{
			Items: []order.Item{{Name: "Item", Quantity: 1, Price: 100}},
		})

		require.ErrorIs(t, err, storageErr)
	})
}

func TestGetByID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		id      int64
		want    order.Order
		wantErr error
	}{
		{name: "существующий заказ", id: 1, want: order.Order{ID: 1}},
		{name: "заказ отсутствует", id: 2, wantErr: order.ErrOrderNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{getByIDFunc: func(_ context.Context, id int64) (order.Order, error) {
				require.Equal(t, tt.id, id)
				return tt.want, tt.wantErr
			}}
			got, err := service.NewOrderService(repo).GetByID(context.Background(), tt.id)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()
	storageErr := errors.New("storage failure")
	tests := []struct {
		name    string
		want    []order.Order
		wantErr error
	}{
		{name: "пустой список", want: []order.Order{}},
		{name: "список заказов", want: []order.Order{{ID: 1}, {ID: 2}}},
		{name: "ошибка репозитория", wantErr: storageErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{listFunc: func(context.Context) ([]order.Order, error) {
				return tt.want, tt.wantErr
			}}
			got, err := service.NewOrderService(repo).List(context.Background())
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		from    order.Status
		to      order.Status
		wantErr error
	}{
		{name: "new в processing", from: order.StatusNew, to: order.StatusProcessing},
		{name: "new в cancelled", from: order.StatusNew, to: order.StatusCancelled},
		{name: "processing в completed", from: order.StatusProcessing, to: order.StatusCompleted},
		{name: "processing в cancelled", from: order.StatusProcessing, to: order.StatusCancelled},
		{name: "new в completed запрещен", from: order.StatusNew, to: order.StatusCompleted, wantErr: order.ErrInvalidStatusTransition},
		{name: "processing в new запрещен", from: order.StatusProcessing, to: order.StatusNew, wantErr: order.ErrInvalidStatusTransition},
		{name: "повтор статуса запрещен", from: order.StatusNew, to: order.StatusNew, wantErr: order.ErrInvalidStatusTransition},
		{name: "completed конечный", from: order.StatusCompleted, to: order.StatusProcessing, wantErr: order.ErrInvalidStatusTransition},
		{name: "cancelled конечный", from: order.StatusCancelled, to: order.StatusProcessing, wantErr: order.ErrInvalidStatusTransition},
		{name: "неизвестный статус", from: order.StatusNew, to: order.Status("unknown"), wantErr: order.ErrInvalidStatus},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initialTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
			initial := order.Order{ID: 1, Status: tt.from, History: []order.StatusChange{{Status: tt.from, ChangedAt: initialTime}}}
			called := false
			repo := &repositoryStub{
				getByIDFunc: func(_ context.Context, id int64) (order.Order, error) {
					require.Equal(t, int64(1), id)
					return initial, nil
				},
				updateFunc: func(_ context.Context, value order.Order) (order.Order, error) {
					called = true
					return value, nil
				},
			}
			got, err := service.NewOrderService(repo).UpdateStatus(context.Background(), 1, tt.to)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.False(t, called)
				return
			}
			require.NoError(t, err)
			require.True(t, called)
			require.Equal(t, tt.to, got.Status)
			require.Len(t, got.History, 2)
			require.Equal(t, initial.History[0], got.History[0])
			require.Equal(t, tt.to, got.History[1].Status)
			require.False(t, got.History[1].ChangedAt.IsZero())
			require.False(t, got.History[1].ChangedAt.Before(initialTime))
		})
	}
}

func TestUpdateStatusRepositoryErrors(t *testing.T) {
	t.Parallel()
	storageErr := errors.New("storage failure")
	tests := []struct {
		name      string
		getErr    error
		updateErr error
	}{
		{name: "заказ не найден", getErr: order.ErrOrderNotFound},
		{name: "ошибка сохранения", updateErr: storageErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &repositoryStub{
				getByIDFunc: func(context.Context, int64) (order.Order, error) {
					return order.Order{ID: 1, Status: order.StatusNew}, tt.getErr
				},
				updateFunc: func(context.Context, order.Order) (order.Order, error) {
					return order.Order{}, tt.updateErr
				},
			}
			_, err := service.NewOrderService(repo).UpdateStatus(context.Background(), 1, order.StatusProcessing)
			if tt.getErr != nil {
				require.ErrorIs(t, err, tt.getErr)
			} else {
				require.ErrorIs(t, err, tt.updateErr)
			}
		})
	}
}

func TestHistory(t *testing.T) {
	t.Parallel()
	early := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	late := early.Add(time.Hour)
	t.Run("история по времени и без изменения исходных данных", func(t *testing.T) {
		original := []order.StatusChange{
			{Status: order.StatusProcessing, ChangedAt: late},
			{Status: order.StatusNew, ChangedAt: early},
		}
		repo := &repositoryStub{getByIDFunc: func(_ context.Context, id int64) (order.Order, error) {
			require.Equal(t, int64(1), id)
			return order.Order{History: original}, nil
		}}
		got, err := service.NewOrderService(repo).History(context.Background(), 1)
		require.NoError(t, err)
		require.Equal(t, []order.StatusChange{original[1], original[0]}, got)
		require.Equal(t, order.StatusProcessing, original[0].Status)
	})
	t.Run("заказ не найден", func(t *testing.T) {
		repo := &repositoryStub{getByIDFunc: func(context.Context, int64) (order.Order, error) {
			return order.Order{}, order.ErrOrderNotFound
		}}
		_, err := service.NewOrderService(repo).History(context.Background(), 1)
		require.ErrorIs(t, err, order.ErrOrderNotFound)
	})
}
