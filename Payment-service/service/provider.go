package service

import (
    "bytes"
    "encoding/json"
    "errors"
    "fmt"
    "net/http"
    "os"
    "strconv"
    "time"

    "github.com/ansh0014/payment/config"
    "github.com/ansh0014/payment/model"
)

// PaymentProvider is an interface for different payment gateways
type PaymentProvider interface {
    CreatePayment(amount float64, currency string, description string, metadata map[string]string) (string, string, error)
    VerifyPayment(gatewayReference string) (model.PaymentStatus, error)
    RefundPayment(gatewayReference string, amount float64, reason string) (string, error)
}

// RazorpayProvider implements PaymentProvider for Razorpay
type RazorpayProvider struct {
    APIKey    string
    SecretKey string
    BaseURL   string
    IsTest    bool
}

// StripeProvider implements PaymentProvider for Stripe
type StripeProvider struct {
    APIKey  string
    BaseURL string
    IsTest  bool
}

// PayUProvider implements PaymentProvider for PayU
type PayUProvider struct {
    MerchantID string
    SecretKey  string
    BaseURL    string
    IsTest     bool
}

// NewPaymentProvider creates a new payment provider based on configuration
func NewPaymentProvider() PaymentProvider {
    gatewayConfig := config.GetPaymentGatewayConfig()

    switch gatewayConfig.Name {
    case "razorpay":
        return &RazorpayProvider{
            APIKey:    gatewayConfig.APIKey,
            SecretKey: gatewayConfig.SecretKey,
            BaseURL:   gatewayConfig.BaseURL,
            IsTest:    gatewayConfig.IsTest,
        }
    // Uncomment these once implemented
    
    // case "stripe":
    //     return &StripeProvider{
    //         APIKey:  gatewayConfig.APIKey,
    //         BaseURL: gatewayConfig.BaseURL,
    //         IsTest:  gatewayConfig.IsTest,
    //     }
    // case "payu":
    //     return &PayUProvider{
    //         MerchantID: gatewayConfig.APIKey,
    //         SecretKey:  gatewayConfig.SecretKey,
    //         BaseURL:    gatewayConfig.BaseURL,
    //         IsTest:     gatewayConfig.IsTest,
    //     }
    
    default:
        // For development/testing, use a mock provider
        return &MockProvider{}
    }
}

// MockProvider for testing without real payment gateway
type MockProvider struct{}

// CreatePayment implements payment creation for MockProvider
func (p *MockProvider) CreatePayment(amount float64, currency string, description string, metadata map[string]string) (string, string, error) {
    // Generate a mock reference
    reference := fmt.Sprintf("mock_%d", time.Now().UnixNano())
    
    // Mock payment URL
    callbackURL := "http://localhost:8003/api/webhook"
    if url, ok := metadata["callback_url"]; ok && url != "" {
        callbackURL = url
    }
    
    paymentURL := fmt.Sprintf("http://localhost:3000/mock-payment?ref=%s&amount=%s&callback=%s", 
        reference, 
        strconv.FormatFloat(amount, 'f', 2, 64),
        callbackURL)
    
    return reference, paymentURL, nil
}

// VerifyPayment implements payment verification for MockProvider
func (p *MockProvider) VerifyPayment(gatewayReference string) (model.PaymentStatus, error) {
    // For testing, always return success
    return model.PaymentStatusCompleted, nil
}

// RefundPayment implements refund for MockProvider
func (p *MockProvider) RefundPayment(gatewayReference string, amount float64, reason string) (string, error) {
    // Generate a mock refund reference
    refundRef := fmt.Sprintf("refund_%s_%d", gatewayReference, time.Now().UnixNano())
    return refundRef, nil
}

// CreatePayment implements payment creation for RazorpayProvider
func (p *RazorpayProvider) CreatePayment(amount float64, currency string, description string, metadata map[string]string) (string, string, error) {
    if p.IsTest {
        // In test mode, use predefined test values
        fmt.Println("[TEST MODE] Creating Razorpay payment")
        reference := fmt.Sprintf("rzp_test_%d", time.Now().UnixNano())
        paymentURL := fmt.Sprintf("https://checkout.razorpay.com/v1/checkout.html?key=%s&amount=%d&order_id=%s", 
            p.APIKey, 
            int(amount*100), // Razorpay uses lowest currency unit (paise)
            reference)
        return reference, paymentURL, nil
    }

    // Prepare Razorpay order request
    razorpayAmount := int(amount * 100) // Convert to paise

    reqBody := map[string]interface{}{
        "amount":   razorpayAmount,
        "currency": currency,
        "receipt":  fmt.Sprintf("rcpt_%d", time.Now().UnixNano()),
        "notes":    metadata,
    }

    jsonData, _ := json.Marshal(reqBody)

    // Create HTTP request
    req, err := http.NewRequest("POST", p.BaseURL+"/orders", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", "", err
    }

    // Set headers
    req.SetBasicAuth(p.APIKey, p.SecretKey)
    req.Header.Set("Content-Type", "application/json")

    // Send request
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", "", err
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return "", "", fmt.Errorf("razorpay error: status code %d", resp.StatusCode)
    }

    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", "", err
    }

    // Extract order ID
    orderID, ok := result["id"].(string)
    if !ok {
        return "", "", errors.New("invalid razorpay response: missing order ID")
    }

    // Generate payment URL
    paymentURL := fmt.Sprintf("https://checkout.razorpay.com/v1/checkout.html?key=%s&amount=%d&order_id=%s", 
        p.APIKey, 
        razorpayAmount,
        orderID)

    return orderID, paymentURL, nil
}

// VerifyPayment implements payment verification for RazorpayProvider
func (p *RazorpayProvider) VerifyPayment(gatewayReference string) (model.PaymentStatus, error) {
    // In a real implementation, you would verify the payment with Razorpay
    // For now, return completed
    return model.PaymentStatusCompleted, nil
}

// RefundPayment implements refund for RazorpayProvider
func (p *RazorpayProvider) RefundPayment(gatewayReference string, amount float64, reason string) (string, error) {
    if p.IsTest {
        // In test mode, use predefined test values
        fmt.Println("[TEST MODE] Processing Razorpay refund")
        return fmt.Sprintf("rfnd_test_%d", time.Now().UnixNano()), nil
    }

    // Prepare refund request
    reqBody := map[string]interface{}{
        "amount": int(amount * 100), // Convert to paise
        "notes": map[string]string{
            "reason": reason,
        },
    }

    jsonData, _ := json.Marshal(reqBody)

    // Create HTTP request
    req, err := http.NewRequest("POST", p.BaseURL+"/payments/"+gatewayReference+"/refund", bytes.NewBuffer(jsonData))
    if err != nil {
        return "", err
    }

    // Set headers
    req.SetBasicAuth(p.APIKey, p.SecretKey)
    req.Header.Set("Content-Type", "application/json")

    // Send request
    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

    // Check response
    if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
        return "", fmt.Errorf("razorpay error: status code %d", resp.StatusCode)
    }

    // Parse response
    var result map[string]interface{}
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", err
    }

    // Extract refund ID
    refundID, ok := result["id"].(string)
    if !ok {
        return "", errors.New("invalid razorpay response: missing refund ID")
    }

    return refundID, nil
}

// Implement Stripe and PayU provider methods similarly
// For brevity, they are omitted here

// GetProviderName returns the configured payment provider name
func GetProviderName() string {
    return os.Getenv("PAYMENT_GATEWAY_NAME")
}