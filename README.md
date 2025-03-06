# Gateway Service

The **Gateway Service** is the API gateway for the PharmaKart backend. It routes incoming requests to the appropriate microservices, performs authentication checks, and provides additional features like rate limiting and logging.

---

## Table of Contents

1. [Overview](#overview)
2. [Features](#features)
3. [Prerequisites](#prerequisites)
4. [Setup and Installation](#setup-and-installation)
5. [Running the Service](#running-the-service)
6. [API Endpoints](#api-endpoints)
7. [Environment Variables](#environment-variables)
8. [Contributing](#contributing)
9. [License](#license)
10. [Support](#support)

---

## Overview

The Gateway Service acts as the entry point for all incoming requests. It:

- Routes requests to the appropriate microservices.
- Validates JWT tokens for authentication.
- Implements rate limiting and request logging.
- Provides API documentation via Swagger.

---

## Features

- **Request Routing**: Routes requests to the appropriate microservices.
- **Authentication**: Verifies JWT tokens for protected routes.
- **Rate Limiting**: Prevents abuse by limiting request rates.
- **Logging**: Logs incoming requests for monitoring and debugging.
- **API Documentation**: Provides Swagger UI for API reference.

---

## Prerequisites

Before setting up the service, ensure you have the following installed:

- **Docker** (for containerization and deployment).
- **Go** (for building and running the service).
- **Protobuf Compiler** (`protoc`) for generating gRPC/protobuf files.

---

## Setup and Installation

### 1. Clone the Repository

Clone the repository and navigate to the gateway service directory:

```bash
git clone https://github.com/PharmaKart/gateway-svc.git
cd gateway-svc
```

### 2. Generate Protobuf Files

Generate the protobuf files using the provided `Makefile`:

```bash
make proto
```

### 3. Install Dependencies

Run the following command to install dependencies:

```bash
go mod tidy
```

### 4. Build the Service

Build the Docker image for the service:

```bash
docker build -t gateway-service .
```

---

## Running the Service

### Start the Service

You can start the service using either of the following methods:

#### Using Docker Run

```bash
docker run -p 8080:8080 --env-file .env gateway-service
```

#### Using Make

```bash
make run
```

The service will be available at:

- **HTTP**: `http://localhost:8080`

### Stop the Service

To stop the service, press `Ctrl + C` if running manually, or stop the Docker container if running in Docker.

---

## API Endpoints

The Gateway Service provides the following endpoints:

### General Endpoints

- **Health Check**: `GET /health`
- **Swagger UI**: `GET /swagger/index.html`

### Authentication

- **User Registration**: `POST /api/v1/register`
- **User Login**: `POST /api/v1/login`

### Product Management

- **List Products**: `GET /api/v1/products`
- **Get Product by ID**: `GET /api/v1/products/:id`
- **Create Product (Admin)**: `POST /api/v1/admin/products`
- **Update Product (Admin)**: `PUT /api/v1/admin/products/:id`
- **Delete Product (Admin)**: `DELETE /api/v1/admin/products/:id`
- **Update Stock (Admin)**: `PUT /api/v1/admin/products/:id/stock`

### Order Management

- **Place Order**: `POST /api/v1/orders`
- **List Customer Orders**: `GET /api/v1/orders`
- **Get Order by ID**: `GET /api/v1/orders/:id`
- **Update Order Status**: `PUT /api/v1/orders/:id`
- **List All Orders (Admin)**: `GET /api/v1/admin/orders`
- **Get Order by ID (Admin)**: `GET /api/v1/admin/orders/:id`
- **Update Order Status (Admin)**: `PUT /api/v1/admin/orders/:id`

### Payment Processing

- **Payment Webhook**: `POST /api/v1/payment/webhook`
- **Get Payment Details**: `GET /api/v1/payment/:id`
- **Get Payment by Order ID**: `GET /api/v1/payment/order/:id`

### Reminder Service

- **Schedule Reminder**: `POST /api/v1/reminders`
- **List Customer Reminders**: `GET /api/v1/reminders`
- **Update Reminder**: `PUT /api/v1/reminders/:id`
- **Delete Reminder**: `DELETE /api/v1/reminders/:id`
- **Toggle Reminder**: `PATCH /api/v1/reminders/:id`
- **List Reminder Logs**: `GET /api/v1/reminders/:id/logs`
- **List All Reminders (Admin)**: `GET /api/v1/admin/reminders`

---

## Environment Variables

The service requires the following environment variables. Create a `.env` file in the `gateway-svc` directory with the following:

```env
PORT=8080
AUTH_SERVICE_URL=http://localhost:50051
PRODUCT_SERVICE_URL=http://localhost:50052
ORDER_SERVICE_URL=http://localhost:50053
PAYMENT_SERVICE_URL=http://localhost:50054
REMINDER_SERVICE_URL=http://localhost:50055
STRIPE_WEBHOOK_SECRET=whsec_your_stripe_webhook_secret
S3_BUCKET_NAME=your_s3_bucket_name
AWS_REGION=ca-central-1
FRONTEND_URL=http://localhost:3000
```

---

## Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bugfix.
3. Submit a pull request with a detailed description of your changes.

---

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

## Support

For any questions or issues, please open an issue in the repository or contact the maintainers.
