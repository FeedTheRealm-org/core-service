package repositories_test

import (
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func repoTestEmail(prefix string) string {
	return prefix + "+" + uuid.NewString() + "@repo.local"
}

func cleanupRepoUsers(db *config.DB) {
	_ = db.Conn.Exec("DELETE FROM password_resets WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@repo.local');")
	_ = db.Conn.Exec("DELETE FROM account_verifications WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@repo.local');")
	_ = db.Conn.Exec("DELETE FROM users WHERE email LIKE '%@repo.local';")
}

func TestAccountRepository_CreateAccount(t *testing.T) {
	logger.InitLogger(false)

	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := repoTestEmail("john.doe")
	passwordHash := "hashed_password"

	user := &models.User{
		Email:    email,
		Password: passwordHash,
	}

	err = repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")

	result, err := repo.GetAccountById(user.Id)
	assert.Nil(t, err, "failed to get account by email")
	assert.NotNil(t, result, "expected user, got nil")
	assert.Equal(t, email, result.Email, "unexpected email")
	assert.Equal(t, passwordHash, result.Password, "unexpected password hash")
}

func TestAccountRepository_GetAccountByEmail_NotFound(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := repoTestEmail("notfound")
	user, err := repo.GetAccountByEmail(email)
	assert.NotNil(t, err, "expected error on getting non-existing user")
	assert.Error(t, err, "Account not found")
	assert.Nil(t, user, "expected no user to be found")
}

func TestAccountRepository_IsAccountVerified(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := repoTestEmail("johndoe")
	passwordHash := "hashed_password"

	user := &models.User{
		Email:    email,
		Password: passwordHash,
	}

	err = repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")
	assert.NotEmpty(t, user.Id.String())

	assert.Nil(t, err, "failed to check if account is verified")
	assert.False(t, user.Verified, "expected account to be unverified")
}

func TestAccountRepository_VerifyAccount(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := repoTestEmail("johndoe_verify_success")
	code := "verification_code"

	user := &models.User{
		Email:    email,
		Password: "hashed_password",
	}

	err = repo.CreateAccount(user, code)
	assert.Nil(t, err, "failed to create account")

	userFromDb, err := repo.GetAccountByEmail(email)
	assert.NoError(t, err, "failed to get account by email")

	err = repo.VerifyAccount(userFromDb, code, time.Now())
	assert.Nil(t, err, "failed to verify account")

	assert.True(t, userFromDb.Verified, "expected account to be verified")
}

func TestAccountRepository_VerifyAccount_Expired(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.Nil(t, err, "failed to connect to database")

	email := repoTestEmail("johndoe_verify_expired")
	code := "verification_code"

	user := &models.User{
		Email:    email,
		Password: "a",
	}

	err = repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")
	assert.NotEmpty(t, user.Id.String())

	err = repo.VerifyAccount(user, code, time.Now().Add(time.Hour))
	assert.Error(t, err, "expected error on verifying expired account")
}

func TestAccountRepository_RefreshVerificationCode_NewRecord(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.NoError(t, err)

	user := &models.User{Email: repoTestEmail("refresh"), Password: "hashed"}
	assert.NoError(t, db.Conn.Create(user).Error)

	expiresAt := time.Now().Add(time.Minute)
	err = repo.RefreshVerificationCode(user, "code123", expiresAt)
	assert.NoError(t, err)

	var record models.AccountVerification
	err = db.Conn.Where("user_id = ?", user.Id).First(&record).Error
	assert.NoError(t, err)
	assert.Equal(t, "code123", record.VerificationCode)
}

func TestAccountRepository_UpdateRefreshTokenUpdatedAtAndAdminStatus(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.NoError(t, err)

	user := &models.User{Email: repoTestEmail("admin"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	now := time.Now().UTC()
	assert.NoError(t, repo.UpdateRefreshTokenUpdatedAt(user.Id, now))
	assert.NoError(t, repo.UpdateAdminStatus(user.Id, true))

	stored, err := repo.GetAccountById(user.Id)
	assert.NoError(t, err)
	assert.True(t, stored.IsAdmin)
	assert.True(t, stored.RefreshTokenUpdatedAt.After(now.Add(-time.Second)))
}

func TestAccountRepository_ListAccounts_FilterAndVerified(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.NoError(t, err)

	userA := &models.User{Email: repoTestEmail("alpha"), Password: "hashed"}
	userB := &models.User{Email: repoTestEmail("beta"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(userA, "code"))
	assert.NoError(t, repo.CreateAccount(userB, "code"))

	_ = db.Conn.Model(&models.User{}).Where("id = ?", userA.Id).Update("verified", true).Error

	verified := true
	users, total, err := repo.ListAccounts("alpha", &verified, 0, 10)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, users, 1)
	assert.Equal(t, userA.Email, users[0].Email)
}

func TestAccountRepository_PasswordResetFlow(t *testing.T) {
	conf := config.CreateConfig()
	db, _ := config.NewDB(conf)
	cleanupRepoUsers(db)
	repo, err := repositories.NewAccountRepository(conf, db)
	assert.NoError(t, err)

	user := &models.User{Email: repoTestEmail("reset"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	reset, err := repo.CreatePasswordReset(user.Id, "otphash", time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, reset.Id)

	active, err := repo.GetActivePasswordResetByUserID(user.Id)
	assert.NoError(t, err)
	assert.Equal(t, reset.Id, active.Id)

	assert.NoError(t, repo.IncrementPasswordResetAttempts(reset.Id))
	assert.NoError(t, repo.MarkPasswordResetOTPVerified(reset.Id, "tokenhash", time.Now().Add(time.Minute)))

	byToken, err := repo.GetPasswordResetByTokenHash("tokenhash")
	assert.NoError(t, err)
	assert.Equal(t, reset.Id, byToken.Id)

	assert.NoError(t, repo.UpdatePassword(user.Id, "newhash"))
	assert.NoError(t, repo.InvalidateAllPasswordResets(user.Id))
}
