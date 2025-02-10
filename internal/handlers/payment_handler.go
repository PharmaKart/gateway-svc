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
			c.Status(http.StatusServiceUnavailable)
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
			c.Status(http.StatusBadRequest) // Return a 400 error on a bad signature
			return
		}

		// Unmarshal the event data into an appropriate struct depending on its Type
		switch event.Type {
		case "checkout.session.async_payment_failed":
			handleAsyncPaymentFailed(event, paymentClient)
		case "checkout.session.async_payment_succeeded":
			handleAsyncPaymentSucceeded(event, paymentClient)
		case "checkout.session.completed":
			handleCheckoutSessionCompleted(event)
		case "checkout.session.expired":
			handleCheckoutSessionExpired(event, paymentClient)
		default:
			utils.Warn("Unhandled event type", map[string]interface{}{
				"event": event.Type,
			})
		}

		c.Status(http.StatusOK)
	}
}

// Handler functions for different event types
func handleAsyncPaymentFailed(event stripe.Event, paymentClient grpc.PaymentClient) {
	// Handle async payment failed
	utils.Warn("Handling async payment failed event", map[string]interface{}{
		"event": event.ID,
	})

	paymentClient.StorePayment(context.Background(), &proto.StorePaymentRequest{
		TransactionId: event.ID,
		OrderId:       event.Data.Object["client_reference_id"].(string),
		CustomerId:    event.Data.Object["customer"].(string),
		Amount:        event.Data.Object["amount_total"].(float64),
		Status:        "failed",
	})
}

func handleAsyncPaymentSucceeded(event stripe.Event, paymentClient grpc.PaymentClient) {
	// Handle async payment succeeded
	paymentClient.StorePayment(context.Background(), &proto.StorePaymentRequest{
		TransactionId: event.ID,
		OrderId:       event.Data.Object["client_reference_id"].(string),
		CustomerId:    event.Data.Object["customer"].(string),
		Amount:        event.Data.Object["amount_total"].(float64),
		Status:        "succeeded",
	})
	utils.Info("Handling async payment succeeded event", map[string]interface{}{
		"event": event.ID,
	})
}

func handleCheckoutSessionCompleted(event stripe.Event) {
	// Handle checkout session completed
	utils.Info("Handling checkout session completed event", map[string]interface{}{
		"event": event.ID,
	})
}

func handleCheckoutSessionExpired(event stripe.Event, paymentClient grpc.PaymentClient) {
	// Handle checkout session expired
	paymentClient.StorePayment(context.Background(), &proto.StorePaymentRequest{
		TransactionId: event.ID,
		OrderId:       event.Data.Object["client_reference_id"].(string),
		CustomerId:    event.Data.Object["customer"].(string),
		Amount:        event.Data.Object["amount_total"].(float64),
		Status:        "expired",
	})
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
// @Router /api/v1/payments/{id} [get]
func GetPayment(paymentClient grpc.PaymentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User Role not found in token"})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
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
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
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
// @Router /api/v1/payments/order/{id} [get]
func GetPaymentByOrderID(paymentClient grpc.PaymentClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole, ok := c.Get("user_role")
		var customerID string
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User Role not found in token"})
			return
		}

		userId, ok := c.Get("user_id")
		if !ok {
			c.JSON(http.StatusUnauthorized, ErrorResponse{Error: "User ID not found in token"})
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
			c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	}
}
