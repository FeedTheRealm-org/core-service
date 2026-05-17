package email_sender

type EmailSenderService interface {
	// CreateBaseEmailData creates a BaseEmailData struct with common email fields populated.
	CreateBaseEmailData(toEmail string) BaseEmailData

	// SendVerificationEmail sends a verification email to the user with the provided data.
	SendVerificationEmail(data VerificationEmailData) error

	// SendGemPurchaseEmail sends an email to the user confirming their gem purchase with the provided data.
	SendGemPurchaseEmail(data GemPurchaseEmailData) error

	// SendGemPurchaseFailedEmail sends an email to the user notifying them of a failed gem purchase with the provided data.
	SendGemPurchaseFailedEmail(data GemPurchaseFailedEmailData) error

	// SendSubscriptionStartedEmail sends an email to the user confirming their subscription start with the provided data.
	SendSubscriptionStartedEmail(data SubscriptionStartedData) error

	// SendPaymentRejectedEmail sends an email to the user notifying them of a failed subscription payment with the provided data.
	SendPaymentRejectedEmail(data SubscriptionPaymentRejectedData) error

	// SendPaymentSuccessfulEmail sends an email to the user confirming their successful subscription payment with the provided data.
	SendPaymentSuccessfulEmail(data SubscriptionPaymentSuccessfulData) error

	// SendSubscriptionReminderEmail sends a reminder email to the user about their upcoming subscription renewal with the provided data.
	SendSubscriptionReminderEmail(data SubscriptionReminderData) error

	// SendSubscriptionCancelledEmail sends an email to the user confirming their subscription cancellation with the provided data.
	SendSubscriptionCancelledEmail(data SubscriptionCancelledData) error
}
