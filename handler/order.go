package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"

	"github.com/aleddaheig/orders-api/model"
	redis "github.com/aleddaheig/orders-api/repository/order"
	"github.com/google/uuid"
)

const (
	decimal = 10
	bitSize = 64
)

// Order handler manages HTTP requests for order resources
type Order struct {
	Repo *redis.RedisRepo
}

func (h *Order) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		CustomerID uuid.UUID        `json:"customer_id"`
		LineItems  []model.LineItem `json:"line_items"`
	}

	// Decode the request body into the body struct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	now := time.Now().UTC()
	order := model.Order{
		OrderID:    rand.Uint64(),
		CustomerID: body.CustomerID,
		LineItems:  body.LineItems,
		CreatedAt:  &now,
	}

	// Insert the order into the repository
	err := h.Repo.Insert(r.Context(), order)
	if err != nil {
		h.handleError(w, err, "insert")
		return
	}

	// Write the order as JSON
	if err := writeJSON(w, http.StatusCreated, order); err != nil {
		h.handleError(w, err, "encode")
		return
	}
}

func (h *Order) List(w http.ResponseWriter, r *http.Request) {
	cursorStr := r.URL.Query().Get("cursor")
	if cursorStr == "" {
		cursorStr = "0"
	}

	// Parse the cursor string into a uint64
	cursor, err := strconv.ParseUint(cursorStr, decimal, bitSize)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Set the size of the page to 50
	const size = 50
	res, err := h.Repo.FindAll(r.Context(), redis.FindAllPage{
		Offset: cursor,
		Size:   size,
	})
	if err != nil {
		h.handleError(w, err, "get orders")
		return
	}

	var response struct {
		Items []model.Order `json:"items"`
		Next  uint64        `json:"next,omitempty"`
	}

	response.Items = res.Orders
	response.Next = res.Cursor

	// Write the response as JSON
	if err := writeJSON(w, http.StatusOK, response); err != nil {
		h.handleError(w, err, "encode")
		return
	}
}

func (h *Order) GetByID(w http.ResponseWriter, r *http.Request) {
	orderID, err := h.parseID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find the order by ID
	o, err := h.Repo.FindByID(r.Context(), orderID)
	if err != nil {
		h.handleError(w, err, "get")
		return
	}

	// Write the order as JSON
	if err := writeJSON(w, http.StatusOK, o); err != nil {
		h.handleError(w, err, "encode")
		return
	}
}

func (h *Order) UpdateByID(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Status string `json:"status"`
	}

	// Decode the request body into the body struct
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	orderID, err := h.parseID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Find the order by ID
	theOrder, err := h.Repo.FindByID(r.Context(), orderID)
	if err != nil {
		h.handleError(w, err, "get")
		return
	}

	const completedStatus = "completed"
	const shippedStatus = "shipped"
	now := time.Now().UTC()

	// Update the order status
	switch body.Status {

	case shippedStatus:
		if theOrder.ShippedAt != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.ShippedAt = &now

	case completedStatus:
		if theOrder.CompletedAt != nil || theOrder.ShippedAt == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		theOrder.CompletedAt = &now

	default:
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Update the order in the repository
	err = h.Repo.Update(r.Context(), theOrder)
	if err != nil {
		h.handleError(w, err, "update")
		return
	}

	// Write the updated order as JSON
	if err := writeJSON(w, http.StatusOK, theOrder); err != nil {
		h.handleError(w, err, "encode")
		return
	}
}

func (h *Order) DeleteByID(w http.ResponseWriter, r *http.Request) {
	orderID, err := h.parseID(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Delete the order from the repository
	err = h.Repo.DeleteByID(r.Context(), orderID)
	if err != nil {
		h.handleError(w, err, "delete")
		return
	}
}

// handleError writes the appropriate HTTP error response based on the error type
func (h *Order) handleError(w http.ResponseWriter, err error, op string) {
	if errors.Is(err, redis.ErrNotExist) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Printf("failed to %s order: %v\n", op, err)
	w.WriteHeader(http.StatusInternalServerError)
}

// parseID extracts and parses the order ID from the request URL path
func (h *Order) parseID(r *http.Request) (uint64, error) {
	idParam := r.PathValue("id")

	return strconv.ParseUint(idParam, decimal, bitSize)
}

// writeJSON marshals the response and sends it with the given status code
func writeJSON(w http.ResponseWriter, status int, v interface{}) error {
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}
