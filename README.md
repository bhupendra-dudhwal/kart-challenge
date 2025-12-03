# About - Kart Challenge

A high-performance backend service built in Go, featuring a clean hexagonal architecture, Redis caching, PostgreSQL database, Docker support, and fast coupon validation using Bloom filters.

# Features

- High-performance HTTP API using Go
- PostgreSQL as primary database
- Redis cache with Bloom Filter for instant coupon lookup
- Docker & Docker Compose support
- Automatic DB Migration + Seed data
- Clean Hexagonal Architecture (Ports & Adapters)
- Parallel processing for coupon files
- OpenAPI Spec + Postman Collection

# Architecture Overview

- This project follows a clean hexagonal architecture:
- Ingress Layer – HTTP handlers, middleware
- Core Layer – Ports, services, models, DTOs
- Egress Layer – Database & cache repositories
- Builder Layer – Boot initialization (coupon processing, migration)
- Config Layer – YAML config-based setup

Coupon codes are preprocessed during boot and stored in Redis using Bloom filters for O(1) lookup.

# Project Structure
```.
├── cmd
│   ├── http
│   │   └── main.go
│   └── migration
│       └── migration.go
├── config
│   └── config.yaml
├── data
│   ├── couponbase1.gz
│   ├── couponbase2.gz
│   └── couponbase3.gz
├── docker-compose.yaml
├── Dockerfile
├── docs
│   ├── openapi
│   │   └── openapi.yaml
│   └── postman
│       └── kart-oolio.postman_collection.json
├── go.mod
├── go.sum
├── internal
│   ├── builder
│   │   ├── builder.go
│   │   ├── coupon.go
│   │   └── migration.go
│   ├── constants
│   │   ├── app.go
│   │   └── context.go
│   ├── core
│   │   ├── models
│   │   │   ├── config.go
│   │   │   └── ingress
│   │   │       ├── dto
│   │   │       │   ├── item_req.go
│   │   │       │   └── order_req.go
│   │   │       ├── item.go
│   │   │       ├── order.go
│   │   │       └── product.go
│   │   ├── ports
│   │   │   ├── egress
│   │   │   │   ├── cache.go
│   │   │   │   ├── database.go
│   │   │   │   ├── order.go
│   │   │   │   └── product.go
│   │   │   ├── ingress
│   │   │   │   ├── handler.go
│   │   │   │   ├── middleware.go
│   │   │   │   ├── migration
│   │   │   │   │   └── mogration.go
│   │   │   │   ├── order.go
│   │   │   │   └── product.go
│   │   │   └── logger.go
│   │   └── services
│   │       ├── migration
│   │       │   └── migration.go
│   │       ├── order.go
│   │       └── product.go
│   ├── egress
│   │   ├── cache
│   │   │   ├── connection.go
│   │   │   └── repository
│   │   │       └── cache.go
│   │   └── database
│   │       ├── connection.go
│   │       └── repository
│   │           ├── order.go
│   │           └── product.go
│   ├── ingress
│   │   └── http
│   │       ├── handler
│   │       │   └── handler.go
│   │       └── middleware
│   │           └── middleware.go
│   └── utils
│       ├── errors.go
│       └── utils.go
├── makefile
├── pkg
│   └── logger
│       └── logger.go
└── README.md
```
# Configuration

All env variables & ports can be modified in:

```
config/config.yaml
```

You may update ports if something is already in use.

# Running the Service – Docker (Recommended)
Brings up Postgres, Redis, and the Go service automatically.

Start services 
 ```
make up
```
Stop services
```
make down
```

# Running Locally (Without Docker)
1. Start Redis & PostgreSQL

Option A – Install locally
Or
Option B – Use Docker Compose:

```
docker compose up -d --build kart-cache
docker compose up -d --build kart-db
```

2. Run Migrations

This will create required tables and seed product data:
```
make migration
```

3. Build & Run the Service
```
make build && make run
```

# Coupon Code System (Bloom Filters)

During application boot:

1. .gz coupon files are uncompressed to .txt
2. Files are processed in parallel
3. Valid coupons are pushed to Redis using Bloom Filters
→ Enables O(1) lookup
4. This process takes a bit of time on startup, but runtime performance becomes extremely fast.

In production, this can be moved to a separate ETL pipeline.

# API Documentation

OpenAPI Spec

```
docs/openapi/openapi.yaml
```

Postman Collection

```
docs/postman/kart-oolio.postman_collection.json
```

# Available APIs
| Endpoint         | Method | Description               |
| ---------------- | ------ | ------------------------- |
| `/products`      | GET    | Get list of all products  |
| `/products/{id}` | GET    | Get product details by ID |
| `/orders`        | POST   | Create an order           |

# Makefile Commands
| Command          | Description                         |
| ---------------- | ----------------------------------- |
| `make up`        | Start Docker containers             |
| `make down`      | Stop Docker containers              |
| `make migration` | Apply DB migrations + seed products |
| `make build`     | Build Go binary                     |
| `make run`       | Run service locally                 |

# Technologies Used
- Go 1.24+
- PostgreSQL
- Redis + Bloom Filters
- Docker & Docker Compose
- GORM ORM
- Clean Architecture
- Makefile for automation
- OpenAPI + Postman

# Kart Challenge

Build an API server implementing our OpenAPI spec for food ordering API in [Go](https://go.dev).\
You can find our [API Documentation](https://orderfoodonline.deno.dev/public/openapi.html) here.

API documentation is based on [OpenAPI3.1](https://swagger.io/specification/v3/) specification.
You can also find spec file [here](https://orderfoodonline.deno.dev/public/openapi.yaml).

> The API immplementation example available to you at orderfoodonline.deno.dev/api is simplified and doesn't handle some edge cases intentionally.
> Use your best judgement to build a Robust API server.

## Basic Requirements

- Implement all APIs described in the OpenAPI specification
- Conform to the OpenAPI specification as close to as possible
- Implement all features our [demo API server](https://orderfoodonline.deno.dev) has implemented
- Validate promo code according to promo code validation logic described below

### Promo Code Validation

You will find three big files containing random text in this repositotory.\
A promo code is valid if the following rules apply:

1. Must be a string of length between 8 and 10 characters
2. It can be found in **at least two** files

> Files containing valid coupons are couponbase1.gz, couponbase2.gz and couponbase3.gz

You can download the files from here

[file 1](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase1.gz)
[file 2](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase2.gz)
[file 3](https://orderfoodonline-files.s3.ap-southeast-2.amazonaws.com/couponbase3.gz)

**Example Promo Codes**

Valid promo codes

- HAPPYHRS
- FIFTYOFF

Invalid promo codes

- SUPER100

> [!TIP]
> it should be noted that there are more valid and invalid promo codes that those shown above.

## Getting Started

You might need to configure Git LFS to clone this repository\
https://github.com/oolio-group/github.com/bhupendra-dudhwal/kart-challenge/tree/advanced-challenge/backend-challenge

1. Use this repository as a template and create a new repository in your account
2. Start coding
3. Share your repository
