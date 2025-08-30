package handler

import (
    "encoding/json"
    "net/http"

    "github.com/ansh0014/payment/model"
    "github.com/ansh0014/payment/service"
    "github.com/ansh0014/payment/utils"
)

// WebhookHandler processes payment gateway callbacks
func WebhookHandler(w http.ResponseWriter, r *http.Request) {
    var webhook model.WebhookRequest
    if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
        utils.BadRequestResponse(w, "Invalid webhook payload", nil)
        return
    }

    // Process the webhook
    err := service.ProcessWebhook(&webhook)
    if err != nil {
        utils.ServerErrorResponse(w, "Failed to process webhook: "+err.Error())
        return
    }

    // Always return 200 OK to payment gateway
    utils.OkResponse(w, "Webhook processed successfully", nil)
}