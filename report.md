# API

## List

```bash
curl localhost:3000/orders | jq
```

**Others pages:**

```bash
curl "localhost:3000/orders?cursor=" | jq
```

## Get by id

```bash
curl localhost:3000/orders/7169153026499611436 | jq
```

## Delete

```bash
curl -X DELETE localhost:3000/orders/17821569105372609616
```

## Create

```bash
curl -X POST -H "Content-Type: application/json" -d '{"customer_id": "'$(uuidgen)'", "line_items": [{"item_id": "'$(uuidgen)'", "quantity": 1, "price": 1000}]}' localhost:3000/orders | jq
```

## Update

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"status": "shipped"}' localhost:3000/orders/6500788779582153952 | jq
```

```bash
curl -X PUT -H "Content-Type: application/json" -d '{"status": "completed"}' localhost:3000/orders/6500788779582153952 | jq
```
