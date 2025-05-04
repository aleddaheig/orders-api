package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/aleddaheig/orders-api/model"
)

// ErrNotExist is returned when the requested order does not exist
var ErrNotExist = errors.New("order does not exist")

// FindAllPage represents pagination parameters for order retrieval
type FindAllPage struct {
	Size   uint64
	Offset uint64
}

// FindResult represents the result of a paginated order retrieval
type FindResult struct {
	Orders []model.Order
	Cursor uint64
}

// RedisRepo handles order persistence operations with Redis
type RedisRepo struct {
	Client *redis.Client
}

// Insert stores a new order in Redis
func (r *RedisRepo) Insert(ctx context.Context, order model.Order) error {
	data, err := serializeOrder(order)
	if err != nil {
		return err
	}

	key := orderIDKey(order.OrderID)

	// Use a transaction to ensure atomicity
	txn := r.Client.TxPipeline()

	// Set the order in Redis
	res := txn.SetNX(ctx, key, data, 0)
	if err := res.Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to set: %w", err)
	}

	// Add the order to the orders set
	if err := txn.SAdd(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to add to orders set: %w", err)
	}

	// Execute the transaction
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

// FindByID retrieves an order by its ID
func (r *RedisRepo) FindByID(ctx context.Context, id uint64) (model.Order, error) {
	key := orderIDKey(id)

	// Retrieve the order from Redis
	value, err := r.Client.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return model.Order{}, ErrNotExist
	} else if err != nil {
		return model.Order{}, fmt.Errorf("get order: %w", err)
	}

	return parseOrder(value)
}

// DeleteByID removes an order by its ID
func (r *RedisRepo) DeleteByID(ctx context.Context, id uint64) error {
	key := orderIDKey(id)

	// Use a transaction to ensure atomicity
	txn := r.Client.TxPipeline()

	// Delete the order from Redis
	err := txn.Del(ctx, key).Err()
	if errors.Is(err, redis.Nil) {
		txn.Discard()
		return ErrNotExist
	} else if err != nil {
		txn.Discard()
		return fmt.Errorf("get order: %w", err)
	}

	// Remove the order from the orders set
	if err := txn.SRem(ctx, "orders", key).Err(); err != nil {
		txn.Discard()
		return fmt.Errorf("failed to remove from orders set: %w", err)
	}

	// Execute the transaction
	if _, err := txn.Exec(ctx); err != nil {
		return fmt.Errorf("failed to exec: %w", err)
	}

	return nil
}

// Update modifies an existing order
func (r *RedisRepo) Update(ctx context.Context, order model.Order) error {
	data, err := serializeOrder(order)
	if err != nil {
		return err
	}

	key := orderIDKey(order.OrderID)

	// Update the order in Redis
	err = r.Client.SetXX(ctx, key, data, 0).Err()
	if errors.Is(err, redis.Nil) {
		return ErrNotExist
	} else if err != nil {
		return fmt.Errorf("set order: %w", err)
	}

	return nil
}

// FindAll retrieves multiple orders with pagination
func (r *RedisRepo) FindAll(ctx context.Context, page FindAllPage) (FindResult, error) {
	// Retrieve the order IDs from the orders set
	res := r.Client.SScan(ctx, "orders", page.Offset, "*", int64(page.Size))

	keys, cursor, err := res.Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get order ids: %w", err)
	}

	if len(keys) == 0 {
		return FindResult{
			Orders: []model.Order{},
		}, nil
	}

	// Retrieve the orders from Redis
	orders, err := r.Client.MGet(ctx, keys...).Result()
	if err != nil {
		return FindResult{}, fmt.Errorf("failed to get orders: %w", err)
	}

	orderList := make([]model.Order, len(orders))

	for i, orderData := range orders {
		order, err := parseOrder(orderData.(string))
		if err != nil {
			return FindResult{}, err
		}
		orderList[i] = order
	}

	return FindResult{
		Orders: orderList,
		Cursor: cursor,
	}, nil
}

// serializeOrder serializes an order to JSON
func serializeOrder(order model.Order) (string, error) {
	data, err := json.Marshal(order)
	if err != nil {
		return "", fmt.Errorf("failed to encode order: %w", err)
	}
	return string(data), nil
}

// parseOrder deserializes JSON data into an order
func parseOrder(data string) (model.Order, error) {
	var order model.Order
	err := json.Unmarshal([]byte(data), &order)
	if err != nil {
		return model.Order{}, fmt.Errorf("failed to decode order json: %w", err)
	}
	return order, nil
}

// orderIDKey returns the Redis key for an order ID
func orderIDKey(id uint64) string {
	return fmt.Sprintf("order:%d", id)
}
