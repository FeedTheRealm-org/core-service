package services

import (
	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/resend/resend-go/v2"
)

type emailSenderService struct {
	conf *config.Config
}

func NewEmailSenderService(conf *config.Config) EmailSenderService {
	return &emailSenderService{
		conf: conf,
	}
}

func (s *emailSenderService) SendVerificationEmail(toEmail string, verifyCode string) error {
	client := resend.NewClient(s.conf.ResendAPIKey)

	params := &resend.SendEmailRequest{
		From:    "onboarding@resend.dev",
		To:      []string{toEmail},
		Subject: "Verification Code",
		Html:    "<p>Your verification code is: <strong>" + verifyCode + "</strong>.</p>",
	}

	_, err := client.Emails.Send(params)
	if err != nil {
		return err
	}

	return nil
}
