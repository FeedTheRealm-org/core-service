package email_sender

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
)

const brevoSendEmailURL = "https://api.brevo.com/v3/smtp/email"
const templatesDir = "templates"

type emailSenderService struct {
	conf *config.Config
}

type BaseEmailData struct {
	ToEmail      string
	LogoURL      string
	SupportEmail string
}

func NewEmailSenderService(conf *config.Config) EmailSenderService {
	return &emailSenderService{conf: conf}
}

func (s *emailSenderService) CreateBaseEmailData(toEmail string) BaseEmailData {
	return BaseEmailData{
		ToEmail:      toEmail,
		LogoURL:      s.conf.EmailLogoURL,
		SupportEmail: s.conf.SupportEmail,
	}
}

func renderAndSend(conf *config.Config, toEmail, subject, templateName string, data any) error {
	templatePath := filepath.Join(templatesDir, templateName+".html")
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("read template %q: %w", templateName, err)
	}

	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("parse template %q: %w", templateName, err)
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return fmt.Errorf("execute template %q: %w", templateName, err)
	}

	payload := map[string]any{
		"sender": map[string]string{
			"name":  "Feed The Realm",
			"email": conf.EmailSenderAddress,
		},
		"to":          []map[string]string{{"email": toEmail}},
		"subject":     subject,
		"htmlContent": htmlBuffer.String(),
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal email payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, brevoSendEmailURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", conf.BrevoAPIKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Logger.Warnf("sendEmail: failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("send email %q failed with status %d: %s", templateName, resp.StatusCode, string(body))
	}

	return nil
}

type VerificationEmailData struct {
	BaseEmailData
	VerifyCode string
}

func (s *emailSenderService) SendVerificationEmail(data VerificationEmailData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Verification Code", "verification_email", data)
}

type PasswordResetEmailData struct {
	BaseEmailData
	VerifyCode string
}

func (s *emailSenderService) SendPasswordResetEmail(data PasswordResetEmailData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Password Reset Code", "password_reset_email", data)
}

type GemPurchaseEmailData struct {
	BaseEmailData
	GemAmount     int64
	TotalGems     int64
	AmountCharged string
	TransactionID string
	PurchaseDate  string
}

func (s *emailSenderService) SendGemPurchaseEmail(data GemPurchaseEmailData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Gems Purchased", "gem_successful", data)
}

type GemPurchaseFailedEmailData struct {
	BaseEmailData
	GemAmount     int64
	AmountCharged string
	TransactionID string
	PurchaseDate  string
}

func (s *emailSenderService) SendGemPurchaseFailedEmail(data GemPurchaseFailedEmailData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Gem Purchase Failed", "gem_rejected", data)
}

type SubscriptionStartedData struct {
	BaseEmailData
	ZoneCount             int64
	Amount                string
	FirstBillingDate      string
	ManageSubscriptionURL string
}

func (s *emailSenderService) SendSubscriptionStartedEmail(data SubscriptionStartedData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Subscription Started", "subscription_started", data)
}

type SubscriptionUpdatedData struct {
	BaseEmailData
	OldZoneCount    int64
	OldAmount       string
	NewZoneCount    int64
	NewAmount       string
	NextBillingDate string
}

func (s *emailSenderService) SendSubscriptionUpdatedEmail(data SubscriptionUpdatedData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Subscription Updated", "subscription_updated", data)
}

type SubscriptionPaymentRejectedData struct {
	BaseEmailData
	ZoneCount int64
	Amount    string
}

func (s *emailSenderService) SendPaymentRejectedEmail(data SubscriptionPaymentRejectedData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Payment Rejected", "subscription_payment_failure",
		data)
}

type SubscriptionPaymentSuccessfulData struct {
	BaseEmailData
	ZoneCount       int
	Amount          string
	PaymentDate     string
	InvoiceID       string
	NextBillingDate string
}

func (s *emailSenderService) SendPaymentSuccessfulEmail(data SubscriptionPaymentSuccessfulData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Subscription Payment Successful", "subscription_payment_successful", data)
}

type SubscriptionCancelledData struct {
	BaseEmailData
	ZoneCount int
}

func (s *emailSenderService) SendSubscriptionCancelledEmail(data SubscriptionCancelledData) error {
	return renderAndSend(s.conf, data.ToEmail, "Feed The Realm - Subscription Cancelled", "subscription_cancelled", data)
}
