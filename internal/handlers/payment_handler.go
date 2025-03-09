package handlers

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/PharmaKart/gateway-svc/internal/grpc"
	"github.com/PharmaKart/gateway-svc/internal/proto"
	"github.com/PharmaKart/gateway-svc/pkg/config"
	"github.com/PharmaKart/gateway-svc/pkg/utils"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/webhook"
)

// HandleWebhook processes Stripe webhook events
// @Summary Process Stripe webhook
// @Description Processes incoming Stripe webhook events
// @Tags Payments
// @Accept json
// @Produce json
// @Success 200 {object} nil "OK"
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 503 {object} utils.ErrorResponse "Service Unavailable"
// @Router /api/v1/webhook [post]
func HandleWebhook(cfg *config.Config, paymentClient grpc.PaymentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		const MaxBodyBytes = int64(65536)

		// Read the body into a buffer
		var buf bytes.Buffer
		reader := io.TeeReader(c.Request.Body, &buf)

		// Read the body with max bytes limit
		payload, err := io.ReadAll(io.LimitReader(reader, MaxBodyBytes))
		if err != nil {
			utils.Error("Error reading request body", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusServiceUnavailable, utils.ErrorResponse{
				Type:    "SERVICE_UNAVAILABLE",
				Message: "Error reading request body",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		// This is your Stripe CLI webhook secret for testing your endpoint locally.
		endpointSecret := cfg.StripeWebhookSecret

		// Pass the request body and Stripe-Signature header to ConstructEvent, along
		// with the webhook signing key.
		event, err := webhook.ConstructEvent(payload, c.GetHeader("Stripe-Signature"), endpointSecret)
		if err != nil {
			utils.Error("Error verifying webhook signature", map[string]interface{}{
				"error": err,
			})
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "VALIDATION_ERROR",
				Message: "Error verifying webhook signature",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		// Unmarshal the event data into an appropriate struct depending on its Type
		switch event.Type {
		case "checkout.session.async_payment_failed":
			handleAsyncPaymentFailed(event, paymentClient)
		case "charge.succeeded":
			handleAsyncPaymentSucceeded(event)
		case "checkout.session.completed":
			handleCheckoutSessionCompleted(event, paymentClient)
		case "checkout.session.expired":
			handleCheckoutSessionExpired(event, paymentClient)
		default:
			utils.Warn("Unhandled event type", map[string]interface{}{
				"event": event.Type,
			})
		}

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": "Webhook processed successfully",
		})
	}
}

// Handler functions for different event types
func handleAsyncPaymentFailed(event stripe.Event, paymentClient grpc.PaymentClient) {
	// Handle async payment failed
	utils.Warn("Handling async payment failed event", map[string]interface{}{
		"event": event.ID,
	})

	_, err := paymentClient.StorePayment(context.Background(), &proto.StorePaymentRequest{
		TransactionId: event.ID,
		OrderId:       event.Data.Object["client_reference_id"].(string),
		CustomerId:    event.Data.Object["customer"].(string),
		Amount:        event.Data.Object["amount_total"].(float64),
		Status:        "failed",
	})

	if err != nil {
		utils.Error("Failed to store failed payment", map[string]interface{}{
			"error": err,
			"event": event.ID,
		})
	}
}

func handleAsyncPaymentSucceeded(event stripe.Event) *string {
	// Handle async payment succeeded

	var receiptUrl *string

	receiptURL, ok := event.Data.Object["receipt_url"].(string)
	if ok {
		receiptUrl = &receiptURL
	} else {
		utils.Warn("Receipt URL not found in event data", map[string]interface{}{
			"event": event.ID,
		})
	}

	utils.Info("Handling async payment succeeded event", map[string]interface{}{
		"event":       event.ID,
		"receipt_url": receiptUrl,
	})

	return receiptUrl
}

func handleCheckoutSessionCompleted(event stripe.Event, paymentClient grpc.PaymentClient) {
	metadata := event.Data.Object["metadata"].(map[string]interface{})

	// Extract individual fields safely
	orderID, ok := metadata["order_id"].(string)
	if !ok {
		// Handle error or log missing metadata
		utils.Warn("Order ID not found in metadata", map[string]interface{}{
			"event": event.ID,
		})
		return
	}

	customerID, ok := metadata["customer_id"].(string)
	if !ok {
		// Handle error or log missing metadata
		utils.Warn("Customer ID not found in metadata", map[string]interface{}{
			"event": event.ID,
		})
		return
	}

	amount, ok := event.Data.Object["amount_total"].(float64)
	if !ok {
		// Handle error or log missing amount
		utils.Warn("Amount not found in event data", map[string]interface{}{
			"event": event.ID,
		})
		return
	}

	status, ok := event.Data.Object["status"].(string)
	if !ok {
		// Handle error or log missing status
		utils.Warn("Status not found in event data", map[string]interface{}{
			"event": event.ID,
		})
		status = "completed" // Default status
	}

	// Handle checkout session completed
	_, err := paymentClient.StorePayment(context.Background(), &proto.StorePaymentRequest{
		TransactionId: event.ID,
		OrderId:       orderID,
		CustomerId:    customerID,
		Amount:        amount / 100,
		Status:        status,
	})

	if err != nil {
		utils.Error("Failed to store completed payment", map[string]interface{}{
			"error": err,
			"event": event.ID,
		})
		return
	}

	utils.Info("Handling checkout session completed event", map[string]interface{}{
		"event": event.ID,
	})
}

func handleCheckoutSessionExpired(event stripe.Event, paymentClient grpc.PaymentClient) {
	// Handle checkout session expired
	_, err := paymentClient.StorePayment(context.Background(), &proto.StorePaymentRequest{
		TransactionId: event.ID,
		OrderId:       event.Data.Object["client_reference_id"].(string),
		CustomerId:    event.Data.Object["customer"].(string),
		Amount:        event.Data.Object["amount_total"].(float64),
		Status:        "expired",
	})

	if err != nil {
		utils.Error("Failed to store expired payment", map[string]interface{}{
			"error": err,
			"event": event.ID,
		})
		return
	}

	utils.Warn("Handling checkout session expired event", map[string]interface{}{
		"event": event.ID,
	})
}

// GetPayment returns a payment by ID
// @Summary Get a payment
// @Description Retrieves a payment by ID
// @Tags Payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Payment ID"
// @Success 200 {object} proto.GetPaymentResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/payments/{id} [get]
func GetPayment(paymentClient grpc.PaymentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User Role not found in token",
			})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		customerID = userId.(string)
		if userRole == "admin" {
			customerID = "admin"
		}

		paymentID := c.Param("id")

		resp, err := paymentClient.GetPayment(c.Request.Context(), &proto.GetPaymentRequest{
			PaymentId:  paymentID,
			CustomerId: customerID,
		})
		if err != nil {
			utils.Error("Failed to get payment", map[string]interface{}{
				"error":      err,
				"payment_id": paymentID,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get payment",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get payment", map[string]interface{}{
				"error":      resp,
				"payment_id": paymentID,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			// Fallback if error structure is not available
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get payment",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}

// GetPaymentByOrderID returns a payment by order ID
// @Summary Get a payment by order ID
// @Description Retrieves a payment by order ID
// @Tags Payments
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param Authorization header string true "Bearer token"
// @Param id path string true "Order ID"
// @Success 200 {object} proto.GetPaymentResponse
// @Failure 400 {object} utils.ErrorResponse "Bad Request"
// @Failure 401 {object} utils.ErrorResponse "Unauthorized"
// @Failure 403 {object} utils.ErrorResponse "Forbidden"
// @Failure 404 {object} utils.ErrorResponse "Not Found"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /api/v1/payments/order/{id} [get]
func GetPaymentByOrderID(paymentClient grpc.PaymentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User Role not found in token",
			})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, utils.ErrorResponse{
				Type:    "AUTH_ERROR",
				Message: "User ID not found in token",
			})
			return
		}

		customerID = userId.(string)
		if userRole == "admin" {
			customerID = "admin"
		}

		orderID := c.Param("id")

		resp, err := paymentClient.GetPaymentByOrderID(c.Request.Context(), &proto.GetPaymentByOrderIDRequest{
			OrderId:    orderID,
			CustomerId: customerID,
		})
		if err != nil {
			utils.Error("Failed to get payment by order ID", map[string]interface{}{
				"error":    err,
				"order_id": orderID,
			})
			c.JSON(http.StatusInternalServerError, utils.ErrorResponse{
				Type:    "INTERNAL_ERROR",
				Message: "Failed to get payment by order ID",
				Details: map[string]string{"error": err.Error()},
			})
			return
		}

		if !resp.Success {
			utils.Error("Failed to get payment by order ID", map[string]interface{}{
				"error":    resp,
				"order_id": orderID,
			})

			if resp.Error != nil {
				errorResp, statusCode := utils.ConvertProtoErrorToResponse(resp.Error)
				c.JSON(statusCode, errorResp)
				return
			}

			// Fallback if error structure is not available
			c.JSON(http.StatusBadRequest, utils.ErrorResponse{
				Type:    "UNKNOWN_ERROR",
				Message: "Failed to get payment by order ID",
			})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
