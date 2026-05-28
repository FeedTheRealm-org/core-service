package gem_balances

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/FeedTheRealm-org/core-service/config"
	gem_balances_errors "github.com/FeedTheRealm-org/core-service/internal/payment-service/errors"
	"github.com/FeedTheRealm-org/core-service/internal/payment-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/utils/email_sender"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

type fakeGemBalancesRepo struct {
	balances       map[uuid.UUID]*models.GemBalance
	getErr         error
	allErr         error
	addErr         error
	addCalled      bool
	addDelta       int64
	upsertErr      error
	applyCalled    bool
	applyResponse  bool
	applyErr       error
	createErr      error
	createdBalance bool
}

func (f *fakeGemBalancesRepo) CreateGemBalance(userId uuid.UUID) error {
	if f.createErr != nil {
		return f.createErr
	}
	if f.balances == nil {
		f.balances = map[uuid.UUID]*models.GemBalance{}
	}
	f.balances[userId] = &models.GemBalance{UserId: userId, Gems: 0}
	f.createdBalance = true
	return nil
}

func (f *fakeGemBalancesRepo) GetAllGemBalances() ([]*models.GemBalance, error) {
	if f.allErr != nil {
		return nil, f.allErr
	}
	var list []*models.GemBalance
	for _, bal := range f.balances {
		list = append(list, bal)
	}
	return list, nil
}

func (f *fakeGemBalancesRepo) GetGemBalanceByUserId(userId uuid.UUID) (*models.GemBalance, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	bal, ok := f.balances[userId]
	if !ok {
		return nil, errors.New("not found")
	}
	return bal, nil
}

func (f *fakeGemBalancesRepo) AddToGemBalance(userId uuid.UUID, gems int64) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.addCalled = true
	f.addDelta = gems
	bal, ok := f.balances[userId]
	if ok {
		bal.Gems += gems
	}
	return nil
}

func (f *fakeGemBalancesRepo) ApplyStripeCheckoutCreditIfUnprocessed(userId uuid.UUID, gems int64, eventID string, sessionID string) (bool, error) {
	f.applyCalled = true
	if f.applyErr != nil {
		return false, f.applyErr
	}
	return f.applyResponse, nil
}

func (f *fakeGemBalancesRepo) UpsertGemBalance(userId uuid.UUID, gems int64) error {
	if f.upsertErr != nil {
		return f.upsertErr
	}
	if f.balances == nil {
		f.balances = map[uuid.UUID]*models.GemBalance{}
	}
	f.balances[userId] = &models.GemBalance{UserId: userId, Gems: gems}
	return nil
}

type fakeGemMetricsRepo struct {
	spentCalled  bool
	boughtCalled bool
}

func (f *fakeGemMetricsRepo) GetMetrics() (*models.GemMetrics, error) {
	return nil, nil
}

func (f *fakeGemMetricsRepo) AddGemsBoughtAndRevenue(gems int64, revenue float64) error {
	f.boughtCalled = true
	return nil
}

func (f *fakeGemMetricsRepo) AddGemsSpent(gems int64) error {
	f.spentCalled = true
	return nil
}

type fakeGemPacksRepo struct {
	pack   *models.GemPack
	getErr error
}

func (f *fakeGemPacksRepo) CreateGemPack(pkg *models.GemPack) (*models.GemPack, error) {
	return nil, nil
}

func (f *fakeGemPacksRepo) GetAllGemPacks() ([]*models.GemPack, error) {
	return nil, nil
}

func (f *fakeGemPacksRepo) GetGemPackById(id uuid.UUID) (*models.GemPack, error) {
	if f.getErr != nil {
		return nil, f.getErr
	}
	return f.pack, nil
}

func (f *fakeGemPacksRepo) UpdateGemPack(id uuid.UUID, updatedPkg *models.GemPack) error {
	return nil
}

func (f *fakeGemPacksRepo) DeleteGemPack(id uuid.UUID) error {
	return nil
}

type fakeCreatorBalancesRepo struct {
	addCalled bool
	addErr    error
}

func (f *fakeCreatorBalancesRepo) AddBalance(userId uuid.UUID, amount decimal.Decimal) error {
	f.addCalled = true
	return f.addErr
}

func (f *fakeCreatorBalancesRepo) GetBalance(userId uuid.UUID) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (f *fakeCreatorBalancesRepo) GetAllBalances() ([]models.CreatorBalance, error) {
	return nil, nil
}

type fakeEmailSender struct {
}

func (f *fakeEmailSender) CreateBaseEmailData(toEmail string) email_sender.BaseEmailData {
	return email_sender.BaseEmailData{ToEmail: toEmail}
}

func (f *fakeEmailSender) SendPasswordResetEmail(data email_sender.PasswordResetEmailData) error {
	return nil
}

func (f *fakeEmailSender) SendVerificationEmail(data email_sender.VerificationEmailData) error {
	return nil
}

func (f *fakeEmailSender) SendGemPurchaseEmail(data email_sender.GemPurchaseEmailData) error {
	return nil
}

func (f *fakeEmailSender) SendGemPurchaseFailedEmail(data email_sender.GemPurchaseFailedEmailData) error {
	return nil
}

func (f *fakeEmailSender) SendSubscriptionStartedEmail(data email_sender.SubscriptionStartedData) error {
	return nil
}

func (f *fakeEmailSender) SendSubscriptionUpdatedEmail(data email_sender.SubscriptionUpdatedData) error {
	return nil
}

func (f *fakeEmailSender) SendPaymentRejectedEmail(data email_sender.SubscriptionPaymentRejectedData) error {
	return nil
}

func (f *fakeEmailSender) SendPaymentSuccessfulEmail(data email_sender.SubscriptionPaymentSuccessfulData) error {
	return nil
}

func (f *fakeEmailSender) SendSubscriptionCancelledEmail(data email_sender.SubscriptionCancelledData) error {
	return nil
}

func TestGemBalancesService_GetAllGemBalances_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemBalancesRepo{allErr: errors.New("boom")}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: repo}

	list, err := service.GetAllGemBalances()
	assert.Error(t, err)
	assert.Nil(t, list)
}

func TestGemBalancesService_GetGemBalanceByUserId_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemBalancesRepo{getErr: errors.New("boom")}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: repo}

	bal, err := service.GetGemBalanceByUserId(uuid.New())
	assert.Error(t, err)
	assert.Nil(t, bal)
}

func TestGemBalancesService_CreateGemBalance_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemBalancesRepo{createErr: errors.New("boom")}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: repo}

	err := service.CreateGemBalance(uuid.New())
	assert.Error(t, err)
}

func TestGemBalancesService_UpdateGemBalance_Error(t *testing.T) {
	conf := config.CreateConfig()
	repo := &fakeGemBalancesRepo{upsertErr: errors.New("boom")}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: repo}

	err := service.UpdateGemBalance(uuid.New(), 10)
	assert.Error(t, err)
}

func TestGemBalancesService_PurchaseCosmetic_Success(t *testing.T) {
	userID := uuid.New()
	cosmeticID := uuid.New()
	creatorID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/assets/internal/cosmetics/") {
			payload := map[string]any{
				"data": map[string]any{
					"cosmetic_id":    cosmeticID,
					"cosmetic_price": int64(10),
					"created_by":     creatorID,
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
			return
		}
		if strings.Contains(r.URL.Path, "/assets/internal/users/") {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port
	conf.Server.CreatorRevenuePercent = 1.0
	conf.Server.DollarsGemsRatio = 1.0

	gemRepo := &fakeGemBalancesRepo{balances: map[uuid.UUID]*models.GemBalance{userID: {UserId: userID, Gems: 20}}}
	metricsRepo := &fakeGemMetricsRepo{}
	creatorRepo := &fakeCreatorBalancesRepo{}
	service := &gemBalancesService{
		conf:                conf,
		gemBalancesRepo:     gemRepo,
		gemMetricsRepo:      metricsRepo,
		creatorBalancesRepo: creatorRepo,
		emailSender:         &fakeEmailSender{},
	}

	err := service.PurchaseCosmetic(userID, cosmeticID)
	assert.NoError(t, err)
	assert.True(t, gemRepo.addCalled)
	assert.Equal(t, int64(-10), gemRepo.addDelta)
	assert.True(t, creatorRepo.addCalled)
	assert.True(t, metricsRepo.spentCalled)
}

func TestGemBalancesService_PurchaseCosmetic_CosmeticNotFound(t *testing.T) {
	userID := uuid.New()
	cosmeticID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	gemRepo := &fakeGemBalancesRepo{balances: map[uuid.UUID]*models.GemBalance{userID: {UserId: userID, Gems: 10}}}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: gemRepo}

	err := service.PurchaseCosmetic(userID, cosmeticID)
	assert.Error(t, err)
	_, notFound := err.(*gem_balances_errors.CosmeticNotFound)
	assert.True(t, notFound)
}

func TestGemBalancesService_PurchaseCosmetic_InsufficientBalance(t *testing.T) {
	userID := uuid.New()
	cosmeticID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload := map[string]any{
			"data": map[string]any{
				"cosmetic_id":    cosmeticID,
				"cosmetic_price": int64(50),
				"created_by":     uuid.Nil,
			},
		}
		_ = json.NewEncoder(w).Encode(payload)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	gemRepo := &fakeGemBalancesRepo{balances: map[uuid.UUID]*models.GemBalance{userID: {UserId: userID, Gems: 10}}}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: gemRepo}

	err := service.PurchaseCosmetic(userID, cosmeticID)
	assert.Error(t, err)
	_, insufficient := err.(*gem_balances_errors.InsufficientGems)
	assert.True(t, insufficient)
}

func TestGemBalancesService_PurchaseCosmetic_AlreadyPurchased(t *testing.T) {
	userID := uuid.New()
	cosmeticID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/assets/internal/cosmetics/") {
			payload := map[string]any{
				"data": map[string]any{
					"cosmetic_id":    cosmeticID,
					"cosmetic_price": int64(1),
					"created_by":     uuid.Nil,
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
			return
		}
		if strings.Contains(r.URL.Path, "/assets/internal/users/") {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	gemRepo := &fakeGemBalancesRepo{balances: map[uuid.UUID]*models.GemBalance{userID: {UserId: userID, Gems: 10}}}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: gemRepo}

	err := service.PurchaseCosmetic(userID, cosmeticID)
	assert.Error(t, err)
	_, conflict := err.(*gem_balances_errors.CosmeticAlreadyPurchased)
	assert.True(t, conflict)
}

func TestGemBalancesService_PurchaseCosmetic_AddBalanceError(t *testing.T) {
	userID := uuid.New()
	cosmeticID := uuid.New()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/assets/internal/cosmetics/") {
			payload := map[string]any{
				"data": map[string]any{
					"cosmetic_id":    cosmeticID,
					"cosmetic_price": int64(1),
					"created_by":     uuid.Nil,
				},
			}
			_ = json.NewEncoder(w).Encode(payload)
			return
		}
		if strings.Contains(r.URL.Path, "/assets/internal/users/") {
			w.WriteHeader(http.StatusCreated)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	gemRepo := &fakeGemBalancesRepo{addErr: errors.New("boom"), balances: map[uuid.UUID]*models.GemBalance{userID: {UserId: userID, Gems: 10}}}
	service := &gemBalancesService{conf: conf, gemBalancesRepo: gemRepo}

	err := service.PurchaseCosmetic(userID, cosmeticID)
	assert.Error(t, err)
}

func TestGemBalancesService_FetchCosmeticPrice_BadJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("not-json"))
	}))
	defer server.Close()

	conf := config.CreateConfig()
	portStr := strings.TrimPrefix(server.URL, "http://127.0.0.1:")
	port, _ := strconv.Atoi(portStr)
	conf.Server.Port = port

	service := &gemBalancesService{conf: conf}
	_, _, err := service.fetchCosmeticPrice(uuid.New())
	assert.Error(t, err)
}

func TestGemBalancesService_CreateCheckoutSession_PackError(t *testing.T) {
	conf := config.CreateConfig()
	service := &gemBalancesService{
		conf:      conf,
		packsRepo: &fakeGemPacksRepo{getErr: errors.New("missing")},
	}

	url, err := service.CreateCheckoutSession(uuid.New(), "user@example.com", uuid.New(), "ok", "cancel")
	assert.Error(t, err)
	assert.Equal(t, "", url)
}

func TestGemBalancesService_HandleWebhook_InvalidSignature(t *testing.T) {
	conf := config.CreateConfig()
	service := &gemBalancesService{conf: conf}

	err := service.HandleWebhook([]byte("{}"), "invalid")
	assert.Error(t, err)
}

func TestGemBalancesService_GetTodayDate_InvalidTimezone(t *testing.T) {
	conf := config.CreateConfig()
	conf.Stripe.StripeBillingTimezone = "Invalid/Zone"
	service := &gemBalancesService{conf: conf}

	date := service.getTodayDate()
	assert.NotEmpty(t, date)
}
