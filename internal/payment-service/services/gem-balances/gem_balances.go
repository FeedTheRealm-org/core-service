package gem_balances

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stripe/stripe-go/v84"
	"github.com/stripe/stripe-go/v84/checkout/session"
	"github.com/stripe/stripe-go/v84/webhook"

	"github.com/FeedTheRealm-org/core-service/config"
	gem_balances_errors "github.com/FeedTheRealm-org/core-service/internal/payment-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	creator_balances_repo "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/creator-balances"
	gem_balances "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-balances"
	gem_packs "github.com/FeedTheRealm-org/core-service/internal/payment-service/repositories/gem-packs"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
)

type gemBalancesService struct {
	conf                *config.Config
	gemBalancesRepo     gem_balances.GemBalancesRepository
	packsRepo           gem_packs.GemPacksRepository
	creatorBalancesRepo creator_balances_repo.CreatorBalancesRepository
}

func NewGemBalancesService(conf *config.Config, gemBalancesRepo gem_balances.GemBalancesRepository, packsRepo gem_packs.GemPacksRepository, creatorBalancesRepo creator_balances_repo.CreatorBalancesRepository) GemBalancesService {
	stripe.Key = conf.Stripe.StripeApiKey
	return &gemBalancesService{
		conf:                conf,
		gemBalancesRepo:     gemBalancesRepo,
		packsRepo:           packsRepo,
		creatorBalancesRepo: creatorBalancesRepo,
	}
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

func (bs *gemBalancesService) PurchaseCosmetic(userId uuid.UUID, cosmeticId uuid.UUID) error {
	price, creatorId, err := bs.fetchCosmeticPrice(cosmeticId)
	if err != nil {
		return err
	}

	if err := bs.ensureSufficientBalance(userId, price); err != nil {
		return err
	}

	if err := bs.issueCosmeticPurchase(userId, cosmeticId); err != nil {
		return err
	}

	if err := bs.gemBalancesRepo.AddToGemBalance(userId, -price); err != nil {
		logger.Logger.Error("Failed to deduct gems after purchase: " + err.Error())
		return err
	}

	if creatorId != uuid.Nil && price > 0 {
		creatorEarnings := float64(price) * bs.conf.Server.CreatorRevenuePercent
		if creatorEarnings > 0 {
			if err := bs.creatorBalancesRepo.AddBalance(creatorId, creatorEarnings); err != nil {
				logger.Logger.Error(fmt.Sprintf("Failed to add revenue balance for creator %s: %s", creatorId, err.Error()))
			}
		}
	}

	logger.Logger.Info(fmt.Sprintf("Successfully purchased cosmetic %s for user %s", cosmeticId, userId))
	return nil
}

func (bs *gemBalancesService) fetchCosmeticPrice(cosmeticId uuid.UUID) (int, uuid.UUID, error) {
	url := fmt.Sprintf("http://127.0.0.1:%d/assets/internal/cosmetics/%s", bs.conf.Server.Port, cosmeticId.String())
	resp, err := http.Get(url)
	if err != nil {
		logger.Logger.Error("Failed to fetch cosmetic details: " + err.Error())
		return 0, uuid.Nil, err
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			logger.Logger.Error("Failed to close response body: " + closeErr.Error())
		}
	}()

	if resp.StatusCode == http.StatusNotFound {
		return 0, uuid.Nil, gem_balances_errors.NewCosmeticNotFound("cosmetic not found")
	} else if resp.StatusCode != http.StatusOK {
		errStr := fmt.Sprintf("Failed to get cosmetic, status code: %d", resp.StatusCode)
		logger.Logger.Error(errStr)
		return 0, uuid.Nil, fmt.Errorf("%s", errStr)
	}

	var cosmeticResp struct {
		Data struct {
			CosmeticId    uuid.UUID `json:"cosmetic_id"`
			CosmeticPrice float64   `json:"cosmetic_price"`
			CreatedBy     uuid.UUID `json:"created_by"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&cosmeticResp); err != nil {
		logger.Logger.Error("Failed to decode cosmetic response: " + err.Error())
		return 0, uuid.Nil, err
	}

	return int(cosmeticResp.Data.CosmeticPrice), cosmeticResp.Data.CreatedBy, nil
}

func (bs *gemBalancesService) ensureSufficientBalance(userId uuid.UUID, price int) error {
	balance, err := bs.gemBalancesRepo.GetGemBalanceByUserId(userId)
	if err != nil {
		logger.Logger.Error("Failed to get user balance: " + err.Error())
		return err
	}

	if balance.Gems < price {
		return gem_balances_errors.NewInsufficientGems("insufficient gems to purchase this cosmetic")
	}

	return nil
}

func (bs *gemBalancesService) issueCosmeticPurchase(userId uuid.UUID, cosmeticId uuid.UUID) error {
	purchaseReqBody := map[string]string{
		"cosmetic_id": cosmeticId.String(),
	}
	reqBodyBytes, err := json.Marshal(purchaseReqBody)
	if err != nil {
		return err
	}

	purchaseUrl := fmt.Sprintf("http://127.0.0.1:%d/assets/internal/users/%s/cosmetics", bs.conf.Server.Port, userId.String())
	postResp, err := http.Post(purchaseUrl, "application/json", bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		logger.Logger.Error("Failed to issue cosmetic purchase: " + err.Error())
		return err
	}
	defer func() {
		if closeErr := postResp.Body.Close(); closeErr != nil {
			logger.Logger.Error("Failed to close response body: " + closeErr.Error())
		}
	}()

	if postResp.StatusCode == http.StatusConflict {
		return gem_balances_errors.NewCosmeticAlreadyPurchased("cosmetic was already purchased by this user")
	} else if postResp.StatusCode != http.StatusCreated {
		errStr := fmt.Sprintf("Failed to record cosmetic purchase, status code: %d", postResp.StatusCode)
		logger.Logger.Error(errStr)
		return fmt.Errorf("%s", errStr)
	}

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
		UnitAmount: stripe.Int64(pack.Price.Mul(decimal.NewFromInt(100)).IntPart()),
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

		applied, err := bs.gemBalancesRepo.ApplyStripeCheckoutCreditIfUnprocessed(userId, pack.Gems, event.ID, session.ID)
		if err != nil {
			logger.Logger.Error("Failed to apply idempotent Stripe credit for user " + userId.String() + ": " + err.Error())
			return err
		}

		if !applied {
			logger.Logger.Info("Ignoring duplicate Stripe checkout session completed event: " + event.ID)
			return nil
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
