package application

import (
	"net/http"

	"github.com/aleddaheig/orders-api/handler"
	"github.com/aleddaheig/orders-api/repository/order"
)

func (a *App) loadRoutes() {
	router := http.NewServeMux()

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	a.loadOrderRoutes(router)

	a.router = router
}

func (a *App) loadOrderRoutes(router *http.ServeMux) {
	orderHandler := &handler.Order{
		Repo: &order.RedisRepo{
			Client: a.rdb,
		},
	}

	router.HandleFunc("POST /orders", orderHandler.Create)
	router.HandleFunc("GET /orders", orderHandler.List)
	router.HandleFunc("GET /orders/{id}", orderHandler.GetByID)
	router.HandleFunc("PUT /orders/{id}", orderHandler.UpdateByID)
	router.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteByID)
}
