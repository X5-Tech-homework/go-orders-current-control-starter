package memory

import (
	"context"
	"testing"

	"example.com/order-service/internal/order"
	"github.com/stretchr/testify/require"
)

func TestCreateOrderRepository(t *testing.T) {
	t.Parallel()

	repo := NewOrderRepository()

	first, err := repo.Create(context.Background(), order.Order{})
	require.NoError(t, err)
	require.Equal(t, int64(1), first.ID)

	second, err := repo.Create(context.Background(), order.Order{})
	require.NoError(t, err)
	require.Equal(t, int64(2), second.ID)
}

func TestGetByID(t *testing.T) {
	t.Parallel()
	repo := NewOrderRepository()
	created, err := repo.Create(context.Background(), order.Order{})
	require.NoError(t, err)
	tests := []struct {
		name    string
		id      int64
		want    order.Order
		wantErr error
	}{
		{
			name: "заказ найден",
			id:   created.ID,
			want: created,
		},
		{
			name:    "заказ не найден",
			id:      2,
			want:    order.Order{},
			wantErr: order.ErrOrderNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(context.Background(), tt.id)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.want, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		orders []order.Order
		want   []order.Order
	}{
		{
			name: "пустой репозиторий",
			want: []order.Order{},
		},
		{
			name: "заказы возвращаются в порядке создания",
			orders: []order.Order{
				{
					Items: []order.Item{
						{Name: "Keyboard", Quantity: 1, Price: 500000},
					},
					History: []order.StatusChange{
						{Status: order.StatusNew},
					},
				},
				{
					Items: []order.Item{
						{Name: "Mouse", Quantity: 2, Price: 150000},
					},
					History: []order.StatusChange{
						{Status: order.StatusNew},
					},
				},
			},
			want: []order.Order{
				{
					ID: 1,
					Items: []order.Item{
						{Name: "Keyboard", Quantity: 1, Price: 500000},
					},
					History: []order.StatusChange{
						{Status: order.StatusNew},
					},
				},
				{
					ID: 2,
					Items: []order.Item{
						{Name: "Mouse", Quantity: 2, Price: 150000},
					},
					History: []order.StatusChange{
						{Status: order.StatusNew},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewOrderRepository()

			for _, value := range tt.orders {
				_, err := repo.Create(context.Background(), value)
				require.NoError(t, err)
			}

			got, err := repo.List(context.Background())
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestUpdate(t *testing.T) {
	t.Parallel()
	order_start := order.Order{
		ID: 1,
		Items: []order.Item{
			{Name: "Keyboard", Quantity: 1, Price: 500000},
		},
		History: []order.StatusChange{
			{Status: order.StatusNew},
		},
	}
	tests := []struct {
		name    string
		update  order.Order
		want    order.Order
		wantErr error
	}{
		{
			name: "для обновления найден заказ по ID",
			update: order.Order{
				ID: order_start.ID,
				Items: []order.Item{
					{Name: "Mouse", Quantity: 2, Price: 150000},
				},
				History: []order.StatusChange{
					{Status: order.StatusNew},
				},
			},
			want: order.Order{
				ID: order_start.ID,
				Items: []order.Item{
					{Name: "Mouse", Quantity: 2, Price: 150000},
				},
				History: []order.StatusChange{
					{Status: order.StatusNew},
				},
			},
		},
		{
			name: "для обновления не найден заказ по ID",
			update: order.Order{
				ID: order_start.ID + 1,
				Items: []order.Item{
					{Name: "Mouse", Quantity: 2, Price: 150000},
				},
				History: []order.StatusChange{
					{Status: order.StatusNew},
				},
			},
			want:    order.Order{},
			wantErr: order.ErrOrderNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewOrderRepository()
			_, err := repo.Create(
				context.Background(),
				order_start,
			)
			require.NoError(t, err)

			got, err := repo.Update(context.Background(), tt.update)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Equal(t, tt.want, got)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
