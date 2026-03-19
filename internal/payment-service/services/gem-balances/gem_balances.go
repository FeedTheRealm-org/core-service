package gem_balances

import (
	"encoding/json"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-balances"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
)

type gemBalancesService struct {
	conf            *config.Config
	gemBalancesRepo gem_balances.GemBalancesRepository
	packsRepo       gem_packs.GemPacksRepository
}

func NewGemBalancesService(conf *config.Config, gemBalancesRepo gem_balances.GemBalancesRepository, packsRepo gem_packs.GemPacksRepository) GemBalancesService {
	stripe.Key = conf.Stripe.StripeSecretKey
	return &gemBalancesService{conf: conf, gemBalancesRepo: gemBalancesRepo, packsRepo: packsRepo}
}

func (bs *gemBalancesService) GetAllGemBalances() ([]*models.GemBalance, error) {
	balances, err := bs.gemBalancesRepo.GetAllGemBalances()
	if err != nil {
		logger.Logger.Error("Failed to retrieve all balances: " + err.Error())
		return nil, err
	}

	logger.Logger.Info("Successfully retrieved all balances")

	return balances, nil
}

func (bs *gemBalancesService) GetGemBalanceByUserId(userId uuid.UUID) (*models.GemBalance, error) {
	balance, err := bs.gemBalancesRepo.GetGemBalanceByUserId(userId)
	if err != nil {
		logger.Logger.Error("Failed to retrieve balance for user " + userId.String() + ": " + err.Error())
		return nil, err
	}

	logger.Logger.Info("Successfully retrieved balance for user " + userId.String())
	return balance, nil
}

func (bs *gemBalancesService) CreateGemBalance(userId uuid.UUID) error {
	err := bs.gemBalancesRepo.CreateGemBalance(userId)
	if err != nil {
		logger.Logger.Error("Failed to create balance for user " + userId.String() + ": " + err.Error())
		return err
	}

	logger.Logger.Info("Successfully created balance for user " + userId.String())
	return nil
}

func (bs *gemBalancesService) UpdateGemBalance(userId uuid.UUID, gems int) error {
	err := bs.gemBalancesRepo.UpdateGemBalance(userId, gems)
	if err != nil {
		logger.Logger.Error("Failed to update balance for user " + userId.String() + ": " + err.Error())
		return err
	}

	logger.Logger.Info("Successfully updated balance for user " + userId.String())
	return nil
}

func (bs *gemBalancesService) CreateCheckoutSession(userId uuid.UUID, packId uuid.UUID, successUrl string, cancelUrl string) (string, error) {
	pack, err := bs.packsRepo.GetGemPackById(packId)
	if err != nil {
		logger.Logger.Error("Failed to retrieve pack with ID " + packId.String() + ": " + err.Error())
		return "", err
	}

	stripeParamsPriceData := &stripe.CheckoutSessionLineItemPriceDataParams{
		Currency: stripe.String("usd"),
		ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
			Name: stripe.String(pack.Name),
		},
		UnitAmount: stripe.Int64(int64(pack.Price * 100)),
	}

	stripeParamsSession := &stripe.CheckoutSessionLineItemParams{
		PriceData: stripeParamsPriceData,
		Quantity:  stripe.Int64(1),
	}

	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			stripeParamsSession,
		},
		Mode:       stripe.String(string(stripe.CheckoutSessionModePayment)),
		SuccessURL: stripe.String(successUrl),
		CancelURL:  stripe.String(cancelUrl),
		Metadata: map[string]string{
			"user_id": userId.String(),
			"pack_id": packId.String(),
		},
	}

	session, err := session.New(params)
	if err != nil {
		logger.Logger.Error("Failed to create Stripe checkout session: " + err.Error())
		return "", err
	}

	logger.Logger.Info("Stripe checkout session created successfully for user " + userId.String() + " and pack " + pack.Name)

	return session.URL, nil
}

func (bs *gemBalancesService) HandleWebhook(payload []byte, signature string) error {
	event, err := webhook.ConstructEvent(payload, signature, bs.conf.Stripe.StripeWebhookSecret)
	if err != nil {
		logger.Logger.Error("Failed to verify Stripe webhook signature: " + err.Error())
		return err
	}

	switch event.Type {
	case "checkout.session.completed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logger.Logger.Error("Failed to parse Stripe webhook event data: " + err.Error())
			return err
		}

		if session.PaymentStatus != stripe.CheckoutSessionPaymentStatusPaid {
			logger.Logger.Warn("Received Stripe checkout session completed event with non-paid status: " + session.PaymentStatus)
			return nil
		}

		userId, err := uuid.Parse(session.Metadata["user_id"])
		if err != nil {
			logger.Logger.Error("Failed to parse user ID from Stripe webhook event metadata: " + err.Error())
			return err
		}

		packId, err := uuid.Parse(session.Metadata["pack_id"])
		if err != nil {
			logger.Logger.Error("Failed to parse pack ID from Stripe webhook event metadata: " + err.Error())
			return err
		}

		pack, err := bs.packsRepo.GetGemPackById(packId)
		if err != nil {
			logger.Logger.Error("Failed to retrieve pack with ID " + packId.String() + ": " + err.Error())
			return err
		}

		balance, err := bs.gemBalancesRepo.GetGemBalanceByUserId(userId)
		if balance == nil {
			logger.Logger.Info("No existing balance found for user " + userId.String() + ", creating new balance record")
			if err := bs.gemBalancesRepo.CreateGemBalance(userId); err != nil {
				logger.Logger.Error("Failed to create balance for user " + userId.String() + ": " + err.Error())
				return err
			}
		} else if err != nil {
			logger.Logger.Error("Failed to retrieve balance for user " + userId.String() + ": " + err.Error())
			return err
		}

		if err := bs.gemBalancesRepo.AddToGemBalance(userId, pack.Gems); err != nil {
			logger.Logger.Error("Failed to update balance for user " + userId.String() + ": " + err.Error())
			return err
		}

		logger.Logger.Info("Processing Stripe checkout session completed event for session ID " + session.ID)
	case "checkout.session.async_payment_failed":
		var session stripe.CheckoutSession
		if err := json.Unmarshal(event.Data.Raw, &session); err != nil {
			logger.Logger.Error("Failed to parse Stripe webhook event data: " + err.Error())
			return err
		}

		logger.Logger.Info("Processing Stripe checkout session async payment failed event for session ID " + session.ID)
	default:
		logger.Logger.Warn("Received unhandled Stripe webhook event type: " + event.Type)
	}

	return nil
}
