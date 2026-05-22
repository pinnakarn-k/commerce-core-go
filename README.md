# commerce-core-go

A production-minded e-commerce backend built with Go.

This project focuses on practical backend engineering fundamentals and architecture patterns commonly used in real-world systems.

The system emphasizes transactional correctness, idempotency, and maintainable backend design over feature completeness.

Key concepts demonstrated in this project include:

- Layered architecture (`handler -> service -> repository`)
- Module-oriented design (`auth`, `cart`, `checkout`, `order`, `payment`, `product`, `user`)
- Explicit transaction management with PostgreSQL
- Idempotent checkout and payment flows
- Payment lifecycle and webhook processing
- Event recording for payment webhooks
- State transition guards for final payment states
- PostgreSQL constraints and data integrity
- JWT authentication
- OpenAPI / Swagger documentation
- Unit testing with mocks and small consumer-owned interfaces
- Authenticated HTTP integration testing
- Transaction-safe webhook processing
- Duplicate webhook event protection
- Thin HTTP handlers and business-focused service layers

The goal of this project is not to build a complete e-commerce platform, but to demonstrate clean, maintainable, and production-oriented backend engineering practices in Go.

---

## Tech Stack

- Go
- Gin HTTP framework
- gRPC
- Protocol Buffers (protobuf)
- PostgreSQL
- pgx PostgreSQL driver
- golang-migrate
- JWT authentication
- bcrypt password hashing
- Testify
- Docker
- Swagger / OpenAPI

---

## Getting Started

Install the following tools before starting:

- Go
- Docker
- Docker Compose
- make
- golang-migrate

Clone repository:

```bash
git clone https://github.com/pinnakarn-k/commerce-core-go.git
cd commerce-core-go
```

Local application runtime uses:

- `configs/.env.local`

Docker development environment uses:

- `configs/.env.docker`

Integration tests use:

- `configs/.env.test`

Create environment files:

Linux / macOS:

```bash
cp configs/.env.local.example configs/.env.local
cp configs/.env.docker.example configs/.env.docker
cp configs/.env.test.example configs/.env.test
```

Windows (CMD):

```cmd
copy configs\.env.local.example configs\.env.local
copy configs\.env.docker.example configs\.env.docker
copy configs\.env.test.example configs\.env.test
```

Start development environment:

```bash
make dev-up-build
```

Run database migrations:

```bash
make migrate-up
```

Run application locally:

```bash
make dev
```

Run application with hot reload:

```bash
make dev-watch
```

Application will be available at:

```text
http://localhost:8080/api/v1
```

---

## API Documentation

Interactive API documentation is available through Swagger UI.

After starting the development environment and running migrations:

```bash
make dev-up-build
make migrate-up
```

Open:
```text
http://localhost:8080/swagger/index.html
```

The Swagger UI includes:

- Request/response schemas
- Validation rules
- Authenticated endpoints
- JWT bearer authentication support
- Interactive request execution

---

## gRPC API

This project also exposes internal query APIs through gRPC.

Current gRPC services:

- `order.v1.OrderQueryService`

Example:

```bash
grpcurl -plaintext -d "{\"user_id\":1,\"order_id\":1}" localhost:9090 order.v1.OrderQueryService/GetOrderDetail
```

---

## Seed User

The following development user is available after running migrations and seed scripts:

```text
email: a@example.com
password: 0123456789
```

---

## Features

### Authentication

- User registration
- JWT-based login
- Protected routes
- Authentication context middleware

### Cart

- Add and update cart items
- Remove cart items
- List current cart items
- Selected item support

### Checkout

- Transaction-based checkout orchestration
- Idempotency key support
- Checkout item snapshot retrieval
- Stock deduction
- Order creation
- Order item snapshot creation
- Payment creation
- Purchased cart item marking

### Orders

- Order listing
- Order detail retrieval
- Order item inclusion in detail responses

### Payments

- Payment creation flow
- Provider payment ID support
- Payment webhook handling
- Payment event recording
- Duplicate webhook event protection
- Final-state transition guards
- Order payment status synchronization
- Mock payment provider support

---

## Architecture

The project follows a layered structure:

```text
handler -> service -> repository
```

### Handler

Responsible for:

- HTTP request/response
- Validation
- Authentication context

### Service

Responsible for:

- Business logic
- Application orchestration
- Transaction flow
- Transaction boundaries
- Domain/application error handling

### Repository

Responsible for:

- SQL execution
- Database interaction
- Data persistence
- Query-level data access

---

## Project Structure

```text
cmd/
  api/                # application entrypoint

internal/
  app/                # application bootstrap
  config/             # configuration loading

  modules/
    auth/
    cart/
    checkout/
    order/
    payment/
    product/
    user/

  platform/
    apperror/         # shared application errors
    authcontext/      # authenticated user context
    database/         # postgres setup
    grpc/             # gRPC server bootstrap
    httpmiddleware/   # HTTP middleware
    logger/           # structured logging
    pagination/       # pagination helpers
    response/         # shared HTTP responses
    token/            # JWT token utilities
    validator/        # validation helpers

  testutil/           # shared test helpers and fixtures

configs/              # environment configurations

deployments/docker/   # docker compose and dockerfiles

docs/                 # swagger generated files

migrations/           # database migrations

.github/
  workflows/          # github actions CI workflows

proto/
  order/v1/
```

---

## Core Flows

### Checkout

The checkout flow is transactionally safe and idempotent.

Flow:

1. Validate checkout command
2. Check existing order by idempotency key
3. Load selected cart items
4. Deduct product stock
5. Create order
6. Create order item snapshots
7. Create payment record
8. Mark cart items as purchased
9. Commit transaction

Guarantees:

- The same idempotency key returns the existing order/payment
- Product stock is deducted only once
- Cart items are marked as purchased only after successful order creation
- Failed checkout rolls back order, payment, cart, and stock changes

### Payment Webhook

The payment webhook flow is transactionally safe and duplicate-safe.

Flow:

1. Begin transaction
2. Store payment event
3. Load payment by provider payment ID
4. Guard final payment states
5. Update payment status
6. Synchronize order payment status
7. Commit transaction

Guarantees:

- Duplicate provider events are ignored safely
- Final payment states are not overwritten
- Payment and order state updates happen in the same transaction
- Failed webhook processing rolls back all changes

---

## Design Decisions

### Why idempotency keys are required for checkout

Checkout operations involve stock deduction and payment creation.

Idempotency guarantees prevent duplicate orders and duplicate stock mutations during retries or network failures.

### Why checkout and webhook flows use transactions

Checkout and payment webhook flows mutate multiple entities such as stock, orders, payments, and cart items.

Database transactions ensure these changes succeed or fail atomically, preventing partial state updates during failures.

### Why payment events are stored

Payment providers may retry webhook delivery multiple times.

Persisting provider event IDs enables duplicate detection and safe webhook replay handling.

---

## Database Design Notes

This project intentionally uses database constraints to enforce data correctness and integrity.

Examples:

```sql
check (stock_qty >= 0)
check (line_total_amount = quantity * unit_price_amount)
unique (idempotency_key)
unique (provider, provider_event_id)
```

The application performs calculations and orchestration, while PostgreSQL enforces correctness and consistency.

---

## Testing Strategy

Integration tests require the test database to be running and migrated.

```bash
docker compose -f deployments/docker/docker-compose.test.yml up -d
make migrate-test-up
```

Run test migrations:

```bash
make migrate-test-up
```

Run all tests:

```bash
make test
```

Run integration tests only:

```bash
make test-integration
```

Run verbose integration tests:

```bash
make test-integration-v
```

Current tests focus on:

- Service layer behavior
- Business flow validation
- Error handling and error mapping
- Checkout orchestration
- Payment webhook orchestration
- Payment state transitions
- Idempotency behavior
- Final-state guards
- Repository mocking through small consumer-owned interfaces

Integration tests cover:

- Repository SQL correctness
- Database constraints
- Row mapping
- Transaction behavior
- Checkout end-to-end persistence flow
- Payment webhook end-to-end persistence flow
- HTTP API integration behavior
- Real application wiring verification
- Authentication middleware behavior
- Protected route verification
- JWT-based authenticated requests
- Idempotency behavior with real database state
- Duplicate webhook event handling
- Rollback behavior on checkout failure
- State transition persistence
- Foreign key and consistency constraints

---

## CI

This project uses GitHub Actions for automated verification.

CI checks include:

- Build verification
- Test suite verification
- Linting
- Integration and HTTP integration tests

Protected branch rules require all checks to pass before merging into `main`.

---

## Developer Commands

### Local Development

Run application locally:
```bash
make dev
```

Run application with hot reload:
```bash
make dev-watch
```

Build application binary:
```bash
make build
```

Run all tests:

```bash
make test
```

Run integration tests only:

```bash
make test-integration
```

### Docker Development

Build and start development containers:

```bash
make dev-up-build
```

Stop and remove development containers and volumes:

```bash
make dev-down-v
```
