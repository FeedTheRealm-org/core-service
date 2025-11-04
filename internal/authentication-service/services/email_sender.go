package services

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/FeedTheRealm-org/core-service/config"
)

const brevoSendEmailURL = "https://api.brevo.com/v3/smtp/email"

type emailSenderService struct {
	conf *config.Config
}

func NewEmailSenderService(conf *config.Config) EmailSenderService {
	return &emailSenderService{
		conf: conf,
	}
}

func createPayloadForSendEmail(fromEmail string, toEmail string, verifyCode string) *bytes.Buffer {
	data := map[string]interface{}{
		"sender": map[string]string{
			"name":  "Feed The Realm",
			"email": fromEmail,
		},
		"to": []map[string]string{
			{
				"email": toEmail,
			},
		},
		"subject":     "Verification Code",
		"htmlContent": "<p>Your verification code is: <strong>" + verifyCode + "</strong>.</p>",
	}
	jsonData, _ := json.Marshal(data)
	return bytes.NewBuffer(jsonData)
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
	data := createPayloadForSendEmail(s.conf.EmailSenderAddress, toEmail, verifyCode)
	req, err := createRequestForSendEmail(data, s.conf.BrevoAPIKey)
	if err != nil {
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		_ = resp.Body.Close()
		return err
	}

	err = resp.Body.Close()
	if err != nil {
		return err
	}

	return nil
}
