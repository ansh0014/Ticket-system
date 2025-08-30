package handler

import (
    "encoding/json"
    "net/http"

    "github.com/ansh0014/payment/model"
    "github.com/ansh0014/payment/service"
    "github.com/ansh0014/payment/utils"
    "github.com/gorilla/mux"
)

// CreatePaymentHandler handles payment initialization
func CreatePaymentHandler(w http.ResponseWriter, r *http.Request) {
    var req model.CreatePaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.BadRequestResponse(w, "Invalid request", nil)
        return
    }

    payment, err := service.CreatePayment(&req)
    if err != nil {
        utils.ServerErrorResponse(w, "Failed to create payment: "+err.Error())
        return
    }

    response := model.PaymentResponse{
        Payment:     *payment,
        RedirectURL: payment.PaymentURL,
        ExpiresIn:   1800, // 30 minutes
    }

    utils.CreatedResponse(w, "Payment created successfully", response)
}

// GetPaymentHandler retrieves payment details
func GetPaymentHandler(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    paymentID := vars["id"]

    payment, err := service.GetPayment(paymentID)
    if err != nil {
        utils.NotFoundResponse(w, "Payment not found")
        return
    }

    utils.OkResponse(w, "Payment retrieved successfully", payment)
}

// RefundPaymentHandler processes refunds
func RefundPaymentHandler(w http.ResponseWriter, r *http.Request) {
    var req model.RefundRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.BadRequestResponse(w, "Invalid request", nil)
        return
    }

    err := service.RefundPayment(&req)
    if err != nil {
        utils.BadRequestResponse(w, err.Error(), nil)
        return
    }

    utils.OkResponse(w, "Refund processed successfully", nil)
}

// VerifyPaymentHandler checks payment status
func VerifyPaymentHandler(w http.ResponseWriter, r *http.Request) {
    var req model.VerifyPaymentRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        utils.BadRequestResponse(w, "Invalid request", nil)
        return
    }

    payment, err := service.GetPayment(req.PaymentID)
    if err != nil {
        utils.NotFoundResponse(w, "Payment not found")
        return
    }

    utils.OkResponse(w, "Payment status retrieved", map[string]interface{}{
        "payment_id": payment.ID,
        "status":     payment.Status,
        "amount":     payment.Amount,
        "currency":   payment.Currency,
    })
}