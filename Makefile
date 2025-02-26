# Variables
PROJECT_NAME = gateway-svc
AUTHENTICATION_SERVICE_NAME = authentication-svc
PRODUCT_SERVICE_NAME = product-svc
ORDER_SERVICE_NAME = order-svc
PAYMENT_SERVICE_NAME = payment-svc
REMINDER_SERVICE_NAME = reminder-svc
GO = go
PROTO_DIR = internal/proto
PROTO_OUT = $(PROTO_DIR)
PORT = 8080

# Targets
.PHONY: build run proto clean

# Build the service
build:
	@echo "Building $(PROJECT_NAME)..."
	$(GO) build -o bin/$(PROJECT_NAME) ./cmd/main.go

# Run the service
run: build
	@echo "Running $(PROJECT_NAME) on port $(PORT)..."
	./bin/$(PROJECT_NAME)

# Run the service in development mode
dev:
	@echo "Running $(PROJECT_NAME) on port $(PORT) with live reload ..."
	air --build.cmd="$(GO) build -o bin/$(PROJECT_NAME) ./cmd/main.go" --build.bin="./bin/$(PROJECT_NAME)"

# Generate swagger docs
swag:
	@echo "Generating swagger docs..."
	swag init -g cmd/main.go

# Generate Go code from .proto file
proto:
	@echo "Generating Go code from Proto files..."
	protoc --go_out=$(PROTO_OUT) --go-grpc_out=$(PROTO_OUT) $(PROTO_DIR)/*.proto

# Copy proto file to related service
proto-copy:
	@echo "Copying Proto files to related service..."
	cp $(PROTO_DIR)/auth.proto ../$(AUTHENTICATION_SERVICE_NAME)/$(PROTO_DIR)

	cp $(PROTO_DIR)/product.proto ../$(PRODUCT_SERVICE_NAME)/$(PROTO_DIR)
	cp $(PROTO_DIR)/product.proto ../$(ORDER_SERVICE_NAME)/$(PROTO_DIR)

	cp $(PROTO_DIR)/order.proto ../$(ORDER_SERVICE_NAME)/$(PROTO_DIR)
	cp $(PROTO_DIR)/order.proto ../$(PAYMENT_SERVICE_NAME)/$(PROTO_DIR)

	cp $(PROTO_DIR)/payment.proto ../$(PAYMENT_SERVICE_NAME)/$(PROTO_DIR)
	cp $(PROTO_DIR)/payment.proto ../$(ORDER_SERVICE_NAME)/$(PROTO_DIR)

	cp $(PROTO_DIR)/reminder.proto ../$(REMINDER_SERVICE_NAME)/$(PROTO_DIR)

# Clean up build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf bin/$(PROJECT_NAME)
