# Gateway Service

The **Gateway Service** is the API gateway for the Pharmakart platform. It routes incoming requests to the appropriate microservices, performs authentication checks, and provides additional features like rate limiting and logging.

---

## Table of Contents
1. [Overview](#overview)
2. [Features](#features)
3. [Prerequisites](#prerequisites)
4. [Setup and Installation](#setup-and-installation)
5. [Running the Service](#running-the-service)
6. [API Documentation](#api-documentation)
7. [Environment Variables](#environment-variables)
8. [Contributing](#contributing)
9. [License](#license)

---

## Overview

The Gateway Service acts as the entry point for all incoming requests. It:
- Routes requests to the appropriate microservices.
- Validates JWT tokens for authentication.
- Implements rate limiting and request logging.

---

## Features

- **Request Routing**: Route requests to the appropriate microservices.
- **Authentication**: Verify JWT tokens for protected routes.
- **Rate Limiting**: Prevent abuse by limiting request rates.
- **Logging**: Log incoming requests for monitoring and debugging.

---

## Prerequisites

Before setting up the service, ensure you have the following installed:
- **Docker**.
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

## API Documentation

The Gateway Service provides the following endpoints:
- **Health Check**: `GET /health`
- **Swagger UI**: `GET /swagger/index.html` (API documentation).

---

## Environment Variables

The service requires the following environment variables. Create a `.env` file in the `gateway-svc` directory with the following:

```env
GATEWAY_PORT=8080
AUTH_SERVICE_URL=http://authentication:50051
PRODUCT_SERVICE_URL=http://authentication:50052
ORDER_SERVICE_URL=http://authentication:50053
PAYMENT_SERVICE_URL=http://authentication:50054
REMINDER_SERVICE_URL=http://authentication:50055
STRIPE_WEBHOOK_SECRET=whsec_your_stripe_webhook_secret
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

