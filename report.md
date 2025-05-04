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
curl -X POST localhost:3000/orders -H "Content-Type: application/json" -d '{"customer_id": "'$(uuidgen)'", "line_items": [{"item_id": "'$(uuidgen)'", "quantity": 1, "price": 1000}]}' | jq
```

## Update

```bash
curl -X PUT localhost:3000/orders/11182650741222677110 -H "Content-Type: application/json" -d '{"status": "shipped"}' | jq
```

```bash
curl -X PUT localhost:3000/orders/11182650741222677110 -H "Content-Type: application/json" -d '{"status": "completed"}' | jq
```
