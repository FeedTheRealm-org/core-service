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

func (s *emailSenderService) SendVerificationEmail(toEmail string, verifyCode string) error {
	data, err := createPayloadForSendEmail(s.conf.EmailSenderAddress, toEmail, verifyCode, s.conf.EmailLogoURL)
	if err != nil {
		return err
	}

	req, err := createRequestForSendEmail(data, s.conf.BrevoAPIKey)
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
			logger.Logger.Warnf("SendVerificationEmail: failed to close response body: %v", cerr)
		}
	}()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to send verification email, status code: %d", resp.StatusCode)
	}

	return nil
}
