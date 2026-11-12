package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"example.com/order-service/internal/handler"
	"example.com/order-service/internal/order"
	"github.com/stretchr/testify/require"
)

type serviceStub struct {
	create  func(context.Context, order.CreateOrderInput) (order.Order, error)
	get     func(context.Context, int64) (order.Order, error)
	list    func(context.Context) ([]order.Order, error)
	update  func(context.Context, int64, order.Status) (order.Order, error)
	history func(context.Context, int64) ([]order.StatusChange, error)
}

func (s *serviceStub) Create(ctx context.Context, input order.CreateOrderInput) (order.Order, error) {
	return s.create(ctx, input)
}
func (s *serviceStub) GetByID(ctx context.Context, id int64) (order.Order, error) {
	return s.get(ctx, id)
}
func (s *serviceStub) List(ctx context.Context) ([]order.Order, error) {
	return s.list(ctx)
}
func (s *serviceStub) UpdateStatus(ctx context.Context, id int64, status order.Status) (order.Order, error) {
	return s.update(ctx, id, status)
}
func (s *serviceStub) History(ctx context.Context, id int64) ([]order.StatusChange, error) {
	return s.history(ctx, id)
}

func TestRouter(t *testing.T) {
	t.Parallel()
	service := &serviceStub{
		create: func(_ context.Context, input order.CreateOrderInput) (order.Order, error) {
			if len(input.Items) == 0 {
				return order.Order{}, order.ErrInvalidOrder
			}
			return order.Order{ID: 1, Items: input.Items}, nil
		},
		get: func(_ context.Context, id int64) (order.Order, error) {
			if id == 99 {
				return order.Order{}, order.ErrOrderNotFound
			}
			if id == 98 {
				return order.Order{}, errors.New("private database details")
			}
			return order.Order{ID: id}, nil
		},
		list: func(context.Context) ([]order.Order, error) {
			return []order.Order{{ID: 1}}, nil
		},
		update: func(_ context.Context, id int64, status order.Status) (order.Order, error) {
			if status == order.StatusCompleted {
				return order.Order{}, order.ErrInvalidStatusTransition
			}
			if status != order.StatusProcessing {
				return order.Order{}, order.ErrInvalidStatus
			}
			return order.Order{ID: id, Status: status}, nil
		},
		history: func(context.Context, int64) ([]order.StatusChange, error) {
			return []order.StatusChange{{Status: order.StatusNew}}, nil
		},
	}
	router := handler.NewRouter(service, slog.New(slog.NewTextHandler(io.Discard, nil)))
	tests := []struct {
		name   string
		method string
		path   string
		body   string
		status int
		code   string
	}{
		{name: "health", method: http.MethodGet, path: "/health", status: 200},
		{name: "создание", method: http.MethodPost, path: "/api/v1/orders", body: `{"items":[{"name":"Keyboard","quantity":1,"price":100}]}`, status: 201},
		{name: "список", method: http.MethodGet, path: "/api/v1/orders", status: 200},
		{name: "получение", method: http.MethodGet, path: "/api/v1/orders/1", status: 200},
		{name: "изменение статуса", method: http.MethodPatch, path: "/api/v1/orders/1/status", body: `{"status":"processing"}`, status: 200},
		{name: "история", method: http.MethodGet, path: "/api/v1/orders/1/history", status: 200},
		{name: "невалидный JSON", method: http.MethodPost, path: "/api/v1/orders", body: `{`, status: 400, code: "INVALID_JSON"},
		{name: "два JSON значения", method: http.MethodPost, path: "/api/v1/orders", body: `{} {}`, status: 400, code: "INVALID_JSON"},
		{name: "невалидный заказ", method: http.MethodPost, path: "/api/v1/orders", body: `{}`, status: 400, code: "INVALID_ORDER"},
		{name: "невалидный ID", method: http.MethodGet, path: "/api/v1/orders/abc", status: 400, code: "INVALID_ID"},
		{name: "нет заказа", method: http.MethodGet, path: "/api/v1/orders/99", status: 404, code: "ORDER_NOT_FOUND"},
		{name: "конфликт статуса", method: http.MethodPatch, path: "/api/v1/orders/1/status", body: `{"status":"completed"}`, status: 409, code: "INVALID_STATUS_TRANSITION"},
		{name: "неизвестный статус", method: http.MethodPatch, path: "/api/v1/orders/1/status", body: `{"status":"unknown"}`, status: 400, code: "INVALID_STATUS"},
		{name: "внутренняя ошибка", method: http.MethodGet, path: "/api/v1/orders/98", status: 500, code: "INTERNAL_ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)
			require.Equal(t, tt.status, response.Code)
			require.Contains(t, response.Header().Get("Content-Type"), "application/json")
			if tt.code != "" {
				var body map[string]string
				require.NoError(t, json.Unmarshal(response.Body.Bytes(), &body))
				require.Equal(t, tt.code, body["code"])
				require.NotContains(t, response.Body.String(), "private database details")
			}
		})
	}
}
