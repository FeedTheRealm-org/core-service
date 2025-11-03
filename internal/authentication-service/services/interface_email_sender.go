package services


type EmailSenderService interface {
	SendVerificationEmail(toEmail string, verifyCode string) error
}
