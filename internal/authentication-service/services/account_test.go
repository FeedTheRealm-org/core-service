package services

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	repoerrs "github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/utils/hashing"
	"github.com/FeedTheRealm-org/core-service/internal/utils/session"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeAccountRepo struct {
	usersByEmail     map[string]*models.User
	usersByID        map[uuid.UUID]*models.User
	getByEmailErr    error
	getByIDErr       error
	createErr        error
	verifyErr        error
	refreshErr       error
	updateRefreshErr error
	updateAdminErr   error
	listUsers        []models.User
	listTotal        int64

	activeReset       *models.PasswordReset
	activeResetErr    error
	createResetErr    error
	incAttemptsErr    error
	markResetErr      error
	resetByToken      *models.PasswordReset
	resetByTokenErr   error
	invalidateErr     error
	updatePasswordErr error
}

func newFakeAccountRepo() *fakeAccountRepo {
	return &fakeAccountRepo{
		usersByEmail: make(map[string]*models.User),
		usersByID:    make(map[uuid.UUID]*models.User),
	}
}

func (f *fakeAccountRepo) GetAccountById(id uuid.UUID) (*models.User, error) {
	if f.getByIDErr != nil {
		return nil, f.getByIDErr
	}
	user, ok := f.usersByID[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (f *fakeAccountRepo) GetAccountByEmail(email string) (*models.User, error) {
	if f.getByEmailErr != nil {
		return nil, f.getByEmailErr
	}
	user, ok := f.usersByEmail[email]
	if !ok {
		return nil, errors.New("not found")
	}
	return user, nil
}

func (f *fakeAccountRepo) CreateAccount(user *models.User, verificationCode string) error {
	if f.createErr != nil {
		return f.createErr
	}
	if user.Id == uuid.Nil {
		user.Id = uuid.New()
	}
	f.usersByEmail[user.Email] = user
	f.usersByID[user.Id] = user
	return nil
}

func (f *fakeAccountRepo) VerifyAccount(user *models.User, code string, currentTime time.Time) error {
	if f.verifyErr != nil {
		return f.verifyErr
	}
	user.Verified = true
	return nil
}

func (f *fakeAccountRepo) RefreshVerificationCode(user *models.User, verificationCode string, expiresAt time.Time) error {
	if f.refreshErr != nil {
		return f.refreshErr
	}
	return nil
}

func (f *fakeAccountRepo) UpdateRefreshTokenUpdatedAt(id uuid.UUID, updatedAt time.Time) error {
	if f.updateRefreshErr != nil {
		return f.updateRefreshErr
	}
	user, ok := f.usersByID[id]
	if ok {
		user.RefreshTokenUpdatedAt = updatedAt
	}
	return nil
}

func (f *fakeAccountRepo) ListAccounts(query string, verified *bool, offset int, limit int) ([]models.User, int64, error) {
	return f.listUsers, f.listTotal, nil
}

func (f *fakeAccountRepo) UpdateAdminStatus(id uuid.UUID, isAdmin bool) error {
	if f.updateAdminErr != nil {
		return f.updateAdminErr
	}
	user, ok := f.usersByID[id]
	if ok {
		user.IsAdmin = isAdmin
	}
	return nil
}

func (f *fakeAccountRepo) CreatePasswordReset(userID uuid.UUID, otpHash string, expiresAt time.Time) (*models.PasswordReset, error) {
	if f.createResetErr != nil {
		return nil, f.createResetErr
	}
	reset := &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       userID,
		OTPHash:      otpHash,
		OTPExpiresAt: expiresAt,
	}
	f.activeReset = reset
	return reset, nil
}

func (f *fakeAccountRepo) GetActivePasswordResetByUserID(userID uuid.UUID) (*models.PasswordReset, error) {
	if f.activeResetErr != nil {
		return nil, f.activeResetErr
	}
	if f.activeReset == nil || f.activeReset.UserId != userID || f.activeReset.Used {
		return nil, errors.New("not found")
	}
	return f.activeReset, nil
}

func (f *fakeAccountRepo) IncrementPasswordResetAttempts(resetID uuid.UUID) error {
	if f.incAttemptsErr != nil {
		return f.incAttemptsErr
	}
	if f.activeReset != nil && f.activeReset.Id == resetID {
		f.activeReset.Attempts++
	}
	return nil
}

func (f *fakeAccountRepo) MarkPasswordResetOTPVerified(resetID uuid.UUID, resetTokenHash string, resetTokenExpiresAt time.Time) error {
	if f.markResetErr != nil {
		return f.markResetErr
	}
	if f.activeReset != nil && f.activeReset.Id == resetID {
		f.activeReset.OTPVerified = true
		f.activeReset.ResetTokenHash = resetTokenHash
		f.activeReset.ResetTokenExpiresAt = &resetTokenExpiresAt
	}
	return nil
}

func (f *fakeAccountRepo) GetPasswordResetByTokenHash(tokenHash string) (*models.PasswordReset, error) {
	if f.resetByTokenErr != nil {
		return nil, f.resetByTokenErr
	}
	if f.resetByToken == nil || f.resetByToken.ResetTokenHash != tokenHash {
		return nil, errors.New("not found")
	}
	return f.resetByToken, nil
}

func (f *fakeAccountRepo) InvalidateAllPasswordResets(userID uuid.UUID) error {
	if f.invalidateErr != nil {
		return f.invalidateErr
	}
	if f.activeReset != nil && f.activeReset.UserId == userID {
		f.activeReset.Used = true
	}
	return nil
}

func (f *fakeAccountRepo) UpdatePassword(userID uuid.UUID, hashedPassword string) error {
	if f.updatePasswordErr != nil {
		return f.updatePasswordErr
	}
	user, ok := f.usersByID[userID]
	if ok {
		user.Password = hashedPassword
	}
	return nil
}

func resetTokenToHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TestAccountService_LoginAccount_InvalidPassword(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	hashed, err := hashing.HashPassword("Password1")
	require.NoError(t, err)
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Password: string(hashed), Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	_, _, _, err = svc.LoginAccount("user@example.com", "Wrong1", false)
	assert.Error(t, err)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_LoginAccount_NotVerified(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	hashed, err := hashing.HashPassword("Password1")
	require.NoError(t, err)
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Password: string(hashed), Verified: false}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	_, _, _, err = svc.LoginAccount("user@example.com", "Password1", false)
	assert.Error(t, err)
	_, notVerified := err.(*AccountNotVerifiedError)
	assert.True(t, notVerified)
}

func TestAccountService_UpdateAdminStatus_InvalidID(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	err := svc.UpdateAdminStatus("not-a-uuid", true)
	assert.Error(t, err)
	_, invalid := err.(*AccountInvalidFormat)
	assert.True(t, invalid)
}

func TestAccountService_ValidateAccessToken_Expired(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", -time.Minute, time.Hour)
	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}

	expiredToken, err := jwtManager.GenerateAccessToken("user-id", "user@example.com", false)
	require.NoError(t, err)

	err = svc.ValidateAccessToken(expiredToken)
	assert.Error(t, err)
	_, expired := err.(*AccountSessionExpired)
	assert.True(t, expired)
}

func TestAccountService_ValidateAccessToken_InvalidUser(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)
	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}

	token, err := jwtManager.GenerateAccessToken(uuid.New().String(), "user@example.com", false)
	require.NoError(t, err)

	err = svc.ValidateAccessToken(token)
	assert.Error(t, err)
	_, invalid := err.(*AccountSessionInvalid)
	assert.True(t, invalid)
}

func TestAccountService_ValidateAccessToken_FixedToken(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	conf.ServerFixedToken = "fixed-token"
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)
	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}

	err := svc.ValidateAccessToken("fixed-token")
	assert.NoError(t, err)
}

func TestAccountService_ValidateAccessToken_MissingUserID(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)
	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}

	claims := jwt.MapClaims{
		"email": "user@example.com",
		"exp":   time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte("access"))
	require.NoError(t, err)

	err = svc.ValidateAccessToken(signed)
	assert.Error(t, err)
	_, invalid := err.(*AccountSessionInvalid)
	assert.True(t, invalid)
}

func TestAccountService_ValidateRefreshToken_ExpiredByUpdate(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	user := &models.User{Id: uuid.New(), Email: "user@example.com", Verified: true}
	user.RefreshTokenUpdatedAt = time.Now().Add(time.Minute)
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	token, err := jwtManager.GenerateRefreshToken(user.Id.String(), user.Email, false)
	require.NoError(t, err)

	err = svc.ValidateRefreshToken(token, user.Email)
	assert.Error(t, err)
	_, expired := err.(*AccountSessionExpired)
	assert.True(t, expired)
}

func TestAccountService_RefreshVerificationCode_AlreadyVerified(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.RefreshVerificationCode(user.Email)
	assert.Error(t, err)
	_, already := err.(*AccountAlreadyVerifiedError)
	assert.True(t, already)
}

func TestAccountService_ForgotPassword_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	otp, err := svc.ForgotPassword(user.Email)
	assert.NoError(t, err)
	assert.Len(t, otp, 6)
}

func TestAccountService_ForgotPassword_NoUser(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	otp, err := svc.ForgotPassword("missing@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "", otp)
}

func TestAccountService_VerifyPasswordResetCode_Invalid(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	hash, err := hashing.HashPassword("123456")
	require.NoError(t, err)
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPHash:      string(hash),
		OTPExpiresAt: time.Now().Add(time.Minute),
	}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err = svc.VerifyPasswordResetCode(user.Email, "000000")
	assert.Error(t, err)
	_, invalid := err.(*InvalidPasswordResetCodeError)
	assert.True(t, invalid)
}

func TestAccountService_VerifyPasswordResetCode_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	hash, err := hashing.HashPassword("123456")
	require.NoError(t, err)
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPHash:      string(hash),
		OTPExpiresAt: time.Now().Add(time.Minute),
	}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	token, err := svc.VerifyPasswordResetCode(user.Email, "123456")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.True(t, repo.activeReset.OTPVerified)
}

// newResetPasswordRepo builds a fake repo pre-loaded with a user and a
// PasswordReset whose ResetTokenHash matches whatever hash the service will
// compute from rawToken before calling GetPasswordResetByTokenHash.
func newResetPasswordRepo(rawToken string, expiresAt time.Time) (*fakeAccountRepo, *models.User) {
	repo := newFakeAccountRepo()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByID[user.Id] = user
	repo.resetByToken = &models.PasswordReset{
		UserId:              user.Id,
		ResetTokenHash:      resetTokenToHash(rawToken), // match what the service will look up
		ResetTokenExpiresAt: &expiresAt,
	}
	return repo, user
}

func TestAccountService_ResetPassword_ExpiredToken(t *testing.T) {
	expiredAt := time.Now().Add(-time.Minute)
	repo, _ := newResetPasswordRepo("sometoken", expiredAt)
	conf := config.CreateConfig()

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.ResetPassword("sometoken", "Password1")
	assert.Error(t, err)
	_, expired := err.(*PasswordResetTokenExpiredError)
	assert.True(t, expired)
}

func TestAccountService_ResetPassword_InvalidPassword(t *testing.T) {
	validUntil := time.Now().Add(time.Minute)
	repo, _ := newResetPasswordRepo("sometoken", validUntil)
	conf := config.CreateConfig()

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.ResetPassword("sometoken", "short")
	assert.Error(t, err)
	_, invalid := err.(*AccountInvalidFormat)
	assert.True(t, invalid)
}

func TestAccountService_ResetPassword_Success(t *testing.T) {
	validUntil := time.Now().Add(time.Minute)
	repo, _ := newResetPasswordRepo("sometoken", validUntil)
	conf := config.CreateConfig()

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.ResetPassword("sometoken", "Password1")
	assert.NoError(t, err)
}

func TestAccountService_CreateAccount_InvalidEmail(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, err := svc.CreateAccount("", "Password1", false)
	assert.Error(t, err)
	invalidErr, ok := err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Empty email", invalidErr.Msg)

	_, _, err = svc.CreateAccount("user@", "Password1", false)
	assert.Error(t, err)
	invalidErr, ok = err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Invalid email", invalidErr.Msg)
}

func TestAccountService_CreateAccount_InvalidPassword(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, err := svc.CreateAccount("user@example.com", "", false)
	assert.Error(t, err)
	invalidErr, ok := err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Empty password", invalidErr.Msg)

	_, _, err = svc.CreateAccount("user@example.com", "short", false)
	assert.Error(t, err)
	invalidErr, ok = err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Password is too short", invalidErr.Msg)

	_, _, err = svc.CreateAccount("user@example.com", "12345678", false)
	assert.Error(t, err)
	invalidErr, ok = err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Password must contain at least one letter", invalidErr.Msg)

	_, _, err = svc.CreateAccount("user@example.com", "PasswordOnly", false)
	assert.Error(t, err)
	invalidErr, ok = err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Password must contain at least one number", invalidErr.Msg)
}

func TestAccountService_CreateAccount_AlreadyExists(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	existing := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[existing.Email] = existing
	repo.usersByID[existing.Id] = existing

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, _, err := svc.CreateAccount(existing.Email, "Password1", false)
	assert.Error(t, err)
	_, exists := err.(*AccountAlreadyExistsError)
	assert.True(t, exists)
}

func TestAccountService_CreateAccount_RepoError(t *testing.T) {
	repo := newFakeAccountRepo()
	repo.createErr = errors.New("db")
	conf := config.CreateConfig()
	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, err := svc.CreateAccount("user@example.com", "Password1", false)
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateError)
	assert.True(t, failed)
}

func TestAccountService_RefreshToken_NotVerified(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Verified: false}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, _, err := svc.RefreshToken(user.Email)
	assert.Error(t, err)
	_, notVerified := err.(*AccountNotVerifiedError)
	assert.True(t, notVerified)
}

func TestAccountService_RefreshToken_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	accessToken, refreshToken, err := svc.RefreshToken(user.Email)
	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
}

func TestAccountService_VerifyAccount_ErrorMapping(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.verifyErr = &repoerrs.AccountVerificationExpired{}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyAccount(user.Email, "code")
	assert.Error(t, err)
	_, expired := err.(*VerificationCodeExpiredError)
	assert.True(t, expired)
}

func TestAccountService_VerifyAccount_InvalidCode(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.verifyErr = &repoerrs.AccountNotVerifiedError{}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyAccount(user.Email, "bad")
	assert.Error(t, err)
	_, invalid := err.(*InvalidVerificationCodeError)
	assert.True(t, invalid)
}

func TestAccountService_ListAccounts(t *testing.T) {
	repo := newFakeAccountRepo()
	repo.listUsers = []models.User{{Email: "a@example.com"}, {Email: "b@example.com"}}
	repo.listTotal = 2

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	verified := true
	users, total, err := svc.ListAccounts("a", &verified, 0, 10)
	assert.NoError(t, err)
	assert.Len(t, users, 2)
	assert.Equal(t, int64(2), total)
}

func TestAccountService_RefreshVerificationCode_RepoError(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com", Verified: false}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.refreshErr = errors.New("boom")

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.RefreshVerificationCode(user.Email)
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateError)
	assert.True(t, failed)
}

func TestAccountService_ResetPassword_InvalidToken(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	err := svc.ResetPassword("token", "Password1")
	assert.Error(t, err)
	_, notFound := err.(*PasswordResetNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_ResetPassword_RepoError(t *testing.T) {
	validUntil := time.Now().Add(time.Minute)
	repo, _ := newResetPasswordRepo("sometoken", validUntil)
	repo.updatePasswordErr = errors.New("boom")
	conf := config.CreateConfig()

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.ResetPassword("sometoken", "Password1")
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateError)
	assert.True(t, failed)
}

func TestAccountService_ForgotPassword_CreateError(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.createResetErr = errors.New("boom")

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.ForgotPassword(user.Email)
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateError)
	assert.True(t, failed)
}

func TestAccountService_VerifyPasswordResetCode_MaxAttempts(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPHash:      "hash",
		OTPExpiresAt: time.Now().Add(time.Minute),
		Attempts:     5,
	}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyPasswordResetCode(user.Email, "123456")
	assert.Error(t, err)
	_, maxed := err.(*PasswordResetMaxAttemptsError)
	assert.True(t, maxed)
}

func TestAccountService_VerifyPasswordResetCode_Expired(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "user@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPHash:      "hash",
		OTPExpiresAt: time.Now().Add(-time.Minute),
	}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyPasswordResetCode(user.Email, "123456")
	assert.Error(t, err)
	_, expired := err.(*PasswordResetExpiredError)
	assert.True(t, expired)
}

func TestAccountService_CreateAccount_InvalidPasswordFallback(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, err := svc.CreateAccount("user@example.com", "Password@", false)
	assert.Error(t, err)
	invalidErr, ok := err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Password must contain at least one number", invalidErr.Msg)
}

func TestAccountService_CreateAccount_EmailInvalidError(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, err := svc.CreateAccount("user@invalid_domain", "Password1", false)
	assert.Error(t, err)
	invalidErr, ok := err.(*AccountInvalidFormat)
	assert.True(t, ok)
	assert.Equal(t, "Invalid email", invalidErr.Msg)
}

func TestAccountService_RefreshVerificationCode_NotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, err := svc.RefreshVerificationCode("missing@example.com")
	assert.Error(t, err)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_ValidateRefreshToken_NotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	err := svc.ValidateRefreshToken("token", "missing@example.com")
	assert.Error(t, err)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_LoginAccount_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	hashed, err := hashing.HashPassword("Password1")
	require.NoError(t, err)
	user := &models.User{Id: uuid.New(), Email: "ok@example.com", Password: string(hashed), Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	gotUser, access, refresh, err := svc.LoginAccount("ok@example.com", "Password1", false)
	assert.NoError(t, err)
	assert.NotNil(t, gotUser)
	assert.NotEmpty(t, access)
	assert.NotEmpty(t, refresh)
}

func TestAccountService_LoginAccount_UserNotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, _, err := svc.LoginAccount("missing@example.com", "Password1", false)
	assert.Error(t, err)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_LoginAccount_UpdateRefreshError(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	hashed, err := hashing.HashPassword("Password1")
	require.NoError(t, err)
	user := &models.User{Id: uuid.New(), Email: "u@example.com", Password: string(hashed), Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.updateRefreshErr = errors.New("db error")

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	_, _, _, err = svc.LoginAccount("u@example.com", "Password1", false)
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateTokenError)
	assert.True(t, failed)
}

// ─── GetUserByEmail ──────────────────────────────────────────────────────────

func TestAccountService_GetUserByEmail_Found(t *testing.T) {
	repo := newFakeAccountRepo()
	user := &models.User{Id: uuid.New(), Email: "found@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	got, err := svc.GetUserByEmail("found@example.com")
	assert.NoError(t, err)
	assert.Equal(t, user.Email, got.Email)
}

func TestAccountService_GetUserByEmail_NotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	got, err := svc.GetUserByEmail("missing@example.com")
	assert.Error(t, err)
	assert.Nil(t, got)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

// ─── ValidateRefreshToken ────────────────────────────────────────────────────

func TestAccountService_ValidateRefreshToken_InvalidClaims(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	user := &models.User{Id: uuid.New(), Email: "u@example.com", Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	// Token firmado con clave distinta → inválido
	token, err := session.NewJWTManager("other", "other", time.Minute, time.Hour).GenerateRefreshToken(user.Id.String(), user.Email, false)
	require.NoError(t, err)

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	err = svc.ValidateRefreshToken(token, user.Email)
	assert.Error(t, err)
	_, invalid := err.(*AccountSessionInvalid)
	assert.True(t, invalid)
}

func TestAccountService_ValidateRefreshToken_UserNotInDB(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	jwtManager := session.NewJWTManager("access", "refresh", time.Minute, time.Hour)

	knownID := uuid.New()
	user := &models.User{Id: knownID, Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	// No lo añadimos a usersByID → GetAccountById fallará

	token, err := jwtManager.GenerateRefreshToken(knownID.String(), user.Email, false)
	require.NoError(t, err)

	svc := &accountService{conf: conf, repo: repo, jwt: jwtManager}
	err = svc.ValidateRefreshToken(token, user.Email)
	assert.Error(t, err)
	_, invalid := err.(*AccountSessionInvalid)
	assert.True(t, invalid)
}

// ─── VerifyAccount ────────────────────────────────────────────────────────────

func TestAccountService_VerifyAccount_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	verified, err := svc.VerifyAccount(user.Email, "code")
	assert.NoError(t, err)
	assert.True(t, verified)
}

func TestAccountService_VerifyAccount_NotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, err := svc.VerifyAccount("missing@example.com", "code")
	assert.Error(t, err)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

// ─── UpdateAdminStatus ───────────────────────────────────────────────────────

func TestAccountService_UpdateAdminStatus_RepoError(t *testing.T) {
	repo := newFakeAccountRepo()
	repo.updateAdminErr = errors.New("db error")
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.UpdateAdminStatus(user.Id.String(), true)
	assert.Error(t, err)
}

// ─── RefreshToken ─────────────────────────────────────────────────────────────

func TestAccountService_RefreshToken_UpdateRefreshError(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "u@example.com", Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.updateRefreshErr = errors.New("boom")

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, _, err := svc.RefreshToken(user.Email)
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateTokenError)
	assert.True(t, failed)
}

func TestAccountService_RefreshToken_UserNotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, _, err := svc.RefreshToken("missing@example.com")
	assert.Error(t, err)
	_, notFound := err.(*AccountNotFoundError)
	assert.True(t, notFound)
}

// ─── CreateAccount (extra paths) ─────────────────────────────────────────────

func TestAccountService_CreateAccount_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	user, code, err := svc.CreateAccount("new@example.com", "Password1", false)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.NotEmpty(t, code)
}

// ─── ForgotPassword (extra paths) ────────────────────────────────────────────

func TestAccountService_ForgotPassword_InvalidateError_Continues(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "u@example.com", Verified: true}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.invalidateErr = errors.New("warning-only")

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	otp, err := svc.ForgotPassword(user.Email)
	assert.NoError(t, err)
	assert.Len(t, otp, 6)
}

// ─── VerifyPasswordResetCode (extra paths) ────────────────────────────────────

func TestAccountService_VerifyPasswordResetCode_NotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, err := svc.VerifyPasswordResetCode("missing@example.com", "123456")
	assert.Error(t, err)
	_, notFound := err.(*PasswordResetNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_VerifyPasswordResetCode_AlreadyVerified(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPVerified:  true,
		OTPExpiresAt: time.Now().Add(time.Minute),
	}

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyPasswordResetCode(user.Email, "123456")
	assert.Error(t, err)
	_, notFound := err.(*PasswordResetNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_VerifyAccount_GenericRepoError(t *testing.T) {
	repo := newFakeAccountRepo()
	conf := config.CreateConfig()
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	// Forzamos un error genérico que no sea ni expirado ni inválido
	repo.verifyErr = errors.New("unexpected database crash")

	svc := &accountService{conf: conf, repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyAccount(user.Email, "code")
	assert.Error(t, err)
	assert.EqualError(t, err, "unexpected database crash")
}

func TestAccountService_UpdateAdminStatus_Success(t *testing.T) {
	repo := newFakeAccountRepo()
	user := &models.User{Id: uuid.New(), Email: "admin@example.com", IsAdmin: false}
	repo.usersByID[user.Id] = user

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.UpdateAdminStatus(user.Id.String(), true)
	assert.NoError(t, err)
	assert.True(t, user.IsAdmin)
}

func TestAccountService_UpdateAdminStatus_UserNotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	// UUID válido pero no existe en el fake → UpdateAdminStatus retorna nil (no verifica existencia)
	validUUID := uuid.New().String()

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	err := svc.UpdateAdminStatus(validUUID, true)
	assert.NoError(t, err)
}

func TestAccountService_VerifyPasswordResetCode_UserNotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}

	_, err := svc.VerifyPasswordResetCode("missing@example.com", "123456")
	assert.Error(t, err)
	_, notFound := err.(*PasswordResetNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_VerifyPasswordResetCode_ActiveResetNotFound(t *testing.T) {
	repo := newFakeAccountRepo()
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	// No le seteamos repo.activeReset, simulando que no hay pedido de reset activo

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyPasswordResetCode(user.Email, "123456")
	assert.Error(t, err)
	_, notFound := err.(*PasswordResetNotFoundError)
	assert.True(t, notFound)
}

func TestAccountService_VerifyPasswordResetCode_RepoIncAttemptsError(t *testing.T) {
	repo := newFakeAccountRepo()
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPHash:      "somehash",
		OTPExpiresAt: time.Now().Add(time.Minute),
	}
	// Forzamos error al intentar incrementar intentos en la BD
	repo.incAttemptsErr = errors.New("db error")

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyPasswordResetCode(user.Email, "wrongcode")
	assert.Error(t, err)
	_, maxed := err.(*PasswordResetMaxAttemptsError)
	assert.True(t, maxed)
}

func TestAccountService_VerifyPasswordResetCode_RepoMarkVerifiedError(t *testing.T) {
	repo := newFakeAccountRepo()
	user := &models.User{Id: uuid.New(), Email: "u@example.com"}
	repo.usersByEmail[user.Email] = user
	repo.usersByID[user.Id] = user

	hash, _ := hashing.HashPassword("123456")
	repo.activeReset = &models.PasswordReset{
		Id:           uuid.New(),
		UserId:       user.Id,
		OTPHash:      string(hash),
		OTPExpiresAt: time.Now().Add(time.Minute),
	}
	// El código es correcto, pero falla al guardar el estado verificado en BD
	repo.markResetErr = errors.New("db error")

	svc := &accountService{conf: config.CreateConfig(), repo: repo, jwt: session.NewJWTManager("a", "b", time.Minute, time.Hour)}
	_, err := svc.VerifyPasswordResetCode(user.Email, "123456")
	assert.Error(t, err)
	_, failed := err.(*AccountFailedToCreateTokenError)
	assert.True(t, failed)
}
