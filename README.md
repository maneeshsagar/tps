# TPS - Transaction Processing System & Interface

Internal transfers System with an HTTP Interface


## Stack

- Go, Gin, GORM
- PostgreSQL
- Kafka (async transfers)
- Docker


## Description
 The system is designed using Hexagonal Architecture at the low level. Transaction processing is supported in both synchronous and asynchronous modes, and I implemented both implementations end-to-end.


## Run
Clone the repository
```bash
https://github.com/maneeshsagar/tps.git
```

```bash
cd tps 
docker-compose up --build -d
```

- API: http://localhost:8080
- Kafka: localhost:9092

## API

### Accounts

```bash
# create
curl -X POST localhost:8080/accounts -H "Content-Type: application/json" \
  -d '{"account_id": 1, "initial_balance": "1000"}'

# get
curl localhost:8080/accounts/1
```

### Sync Transfer

Blocks until complete.

```bash
curl -X POST localhost:8080/transactions -H "Content-Type: application/json" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100"}'
```

### Async Transfer

Returns immediately, processes via Kafka consumer.

```bash
# submit
curl -X POST localhost:8080/async-transactions -H "Content-Type: application/json" \
  -d '{"source_account_id": 1, "destination_account_id": 2, "amount": "100"}'

# check status
curl localhost:8080/async-transactions/{id}/status
```

Status: pending → completed or failed


## Concurrency

- PostgreSQL advisory locks (sorted order to prevent deadlocks)
- Atomic transactions via GORM

## Tables
The system uses three core tables, which act as the source of truth:
- **accounts** : Stores account-level information, including **account_id** and **balance**.
- **transactions** : Stores details of successful transactions.
- **async_transactions_status** : Stores the status and metadata of submitted asynchronous transactions.
## Failed Asynsc Transaction
- If a business validation failure occurs, the transaction is immediately marked as **failed**, along with the failure reason, in the **async_transactions_status** table.
- If a transient failure occurs, the message is retried 3 times. After 3 unsuccessful retries, it is pushed to transactions-dlq, and the transaction is marked as **failed** in the async_transactions_status table.

## Kafka Topics 
1. **transactions**
2. **transactions-dlq**

## Postman Collection
[TPS Postman collection](tps.postman_collection.json)


## Notes & Assumptions

- Amounts in rupees, stored as paise (int64)
- Max 2 decimal places(rupees can only be valid till two decimal places)
- No overdrafts
- No self-transfers
- Tables auto-migrate on startup
