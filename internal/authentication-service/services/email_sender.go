package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"

	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"

	"github.com/FeedTheRealm-org/core-service/config"
)

const brevoSendEmailURL = "https://api.brevo.com/v3/smtp/email"
const filepathTemplates = "templates"
const templateVerificationEmail = "verification_email.html"
const templateName = "verification_email"
const templatePasswordResetEmail = "password_reset_email.html"
const templateNamePasswordReset = "password_reset_email"

// EmailTemplateData holds the data for email templates
type EmailTemplateData struct {
	VerifyCode string
	ToEmail    string
	LogoURL    string
}

type emailSenderService struct {
	conf *config.Config
}

func NewEmailSenderService(conf *config.Config) EmailSenderService {
	return &emailSenderService{
		conf: conf,
	}
}

func createEmailTemplate() (*template.Template, error) {
	templatePath := filepath.Join(filepathTemplates, templateVerificationEmail)
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	tmpl, err := template.New(templateName).Parse(string(templateContent))
	if err != nil {
		return nil, err
	}

	return tmpl, nil
}

func createPayloadForSendEmail(fromEmail string, toEmail string, verifyCode string, logoURL string) (*bytes.Buffer, error) {
	tmpl, err := createEmailTemplate()
	if err != nil {
		return nil, err
	}

	data := EmailTemplateData{
		VerifyCode: verifyCode,
		ToEmail:    toEmail,
		LogoURL:    logoURL,
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return nil, err
	}

	emailData := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Feed The Realm",
			"email": fromEmail,
		},
		"to": []map[string]string{
			{"email": toEmail},
		},
		"subject":     "Feed The Realm - Verification Code",
		"htmlContent": htmlBuffer.String(),
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return nil, err
	}

	return bytes.NewBuffer(jsonData), nil
}

func createRequestForSendEmail(payload *bytes.Buffer, apiKey string) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodPost, brevoSendEmailURL, payload)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("api-key", apiKey)

	return req, nil
}

func (s *emailSenderService) sendEmail(payload *bytes.Buffer, logTag string) error {
	req, err := createRequestForSendEmail(payload, s.conf.BrevoAPIKey)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			logger.Logger.Warnf("%s: failed to close response body: %v", logTag, cerr)
		}
	}()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send email (%s), status code: %d", logTag, resp.StatusCode)
	}

	return nil
}

func (s *emailSenderService) SendVerificationEmail(toEmail string, verifyCode string) error {
	data, err := createPayloadForSendEmail(s.conf.EmailSenderAddress, toEmail, verifyCode, s.conf.EmailLogoURL)
	if err != nil {
		return err
	}
	return s.sendEmail(data, "SendVerificationEmail")
}

func (s *emailSenderService) SendPasswordResetEmail(toEmail string, otpCode string) error {
	templatePath := filepath.Join(filepathTemplates, templatePasswordResetEmail)
	templateContent, err := os.ReadFile(templatePath)
	if err != nil {
		return err
	}

	tmpl, err := template.New(templateNamePasswordReset).Parse(string(templateContent))
	if err != nil {
		return err
	}

	data := EmailTemplateData{
		VerifyCode: otpCode,
		ToEmail:    toEmail,
		LogoURL:    s.conf.EmailLogoURL,
	}

	var htmlBuffer bytes.Buffer
	if err := tmpl.Execute(&htmlBuffer, data); err != nil {
		return err
	}

	emailData := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Feed The Realm",
			"email": s.conf.EmailSenderAddress,
		},
		"to": []map[string]string{
			{"email": toEmail},
		},
		"subject":     "Feed The Realm - Password Reset Code",
		"htmlContent": htmlBuffer.String(),
	}

	jsonData, err := json.Marshal(emailData)
	if err != nil {
		return err
	}

	return s.sendEmail(bytes.NewBuffer(jsonData), "SendPasswordResetEmail")
}
