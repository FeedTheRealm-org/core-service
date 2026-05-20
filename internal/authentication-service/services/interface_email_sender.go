package services

type EmailSenderService interface {
	SendVerificationEmail(toEmail string, verifyCode string) error
	SendPasswordResetEmail(toEmail string, otpCode string) error
}
