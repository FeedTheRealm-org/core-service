package repositories_test

import (
	"os"
	"testing"
	"time"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/models"
	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/repositories"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sharedDB and sharedRepo are initialised once in TestMain and reused by all tests.
var (
	sharedDB   *config.DB
	sharedRepo repositories.AccountRepository
)

func TestMain(m *testing.M) {
	logger.InitLogger(false)

	conf := config.CreateConfig()

	var err error
	sharedDB, err = config.NewDB(conf)
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	sharedRepo, err = repositories.NewAccountRepository(conf, sharedDB)
	if err != nil {
		panic("failed to create account repository: " + err.Error())
	}

	os.Exit(m.Run())
}

func repoTestEmail(prefix string) string {
	return prefix + "+" + uuid.NewString() + "@repo.local"
}

func cleanupRepoUsers(db *config.DB) {
	_ = db.Conn.Exec("DELETE FROM password_resets WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@repo.local');")
	_ = db.Conn.Exec("DELETE FROM account_verifications WHERE user_id IN (SELECT id FROM users WHERE email LIKE '%@repo.local');")
	_ = db.Conn.Exec("DELETE FROM users WHERE email LIKE '%@repo.local';")
}

// setupTest cleans up leftover rows and returns the shared DB and repo.
func setupTest(t *testing.T) (*config.DB, repositories.AccountRepository) {
	t.Helper()
	require.NotNil(t, sharedDB, "sharedDB is nil — TestMain did not run")
	cleanupRepoUsers(sharedDB)
	return sharedDB, sharedRepo
}

func TestAccountRepository_CreateAccount(t *testing.T) {
	db, repo := setupTest(t)
	_ = db

	email := repoTestEmail("john.doe")
	passwordHash := "hashed_password"

	user := &models.User{
		Email:    email,
		Password: passwordHash,
	}

	err := repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")

	result, err := repo.GetAccountById(user.Id)
	assert.Nil(t, err, "failed to get account by id")
	assert.NotNil(t, result, "expected user, got nil")
	assert.Equal(t, email, result.Email, "unexpected email")
	assert.Equal(t, passwordHash, result.Password, "unexpected password hash")
}

func TestAccountRepository_GetAccountByEmail_NotFound(t *testing.T) {
	_, repo := setupTest(t)

	email := repoTestEmail("notfound")
	user, err := repo.GetAccountByEmail(email)
	assert.NotNil(t, err, "expected error on getting non-existing user")
	assert.Error(t, err, "Account not found")
	assert.Nil(t, user, "expected no user to be found")
}

func TestAccountRepository_IsAccountVerified(t *testing.T) {
	_, repo := setupTest(t)

	email := repoTestEmail("johndoe")
	user := &models.User{
		Email:    email,
		Password: "hashed_password",
	}

	err := repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")
	assert.NotEmpty(t, user.Id.String())
	assert.False(t, user.Verified, "expected account to be unverified")
}

func TestAccountRepository_VerifyAccount(t *testing.T) {
	_, repo := setupTest(t)

	email := repoTestEmail("johndoe_verify_success")
	code := "verification_code"

	user := &models.User{
		Email:    email,
		Password: "hashed_password",
	}

	err := repo.CreateAccount(user, code)
	assert.Nil(t, err, "failed to create account")

	userFromDb, err := repo.GetAccountByEmail(email)
	assert.NoError(t, err, "failed to get account by email")

	err = repo.VerifyAccount(userFromDb, code, time.Now())
	assert.Nil(t, err, "failed to verify account")
	assert.True(t, userFromDb.Verified, "expected account to be verified")
}

func TestAccountRepository_VerifyAccount_Expired(t *testing.T) {
	_, repo := setupTest(t)

	email := repoTestEmail("johndoe_verify_expired")
	code := "verification_code"

	user := &models.User{
		Email:    email,
		Password: "a",
	}

	err := repo.CreateAccount(user, "verification_code")
	assert.Nil(t, err, "failed to create account")
	assert.NotEmpty(t, user.Id.String())

	err = repo.VerifyAccount(user, code, time.Now().Add(time.Hour))
	assert.Error(t, err, "expected error on verifying expired account")
}

func TestAccountRepository_RefreshVerificationCode_NewRecord(t *testing.T) {
	db, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("refresh"), Password: "hashed"}
	assert.NoError(t, db.Conn.Create(user).Error)

	expiresAt := time.Now().Add(time.Minute)
	err := repo.RefreshVerificationCode(user, "code123", expiresAt)
	assert.NoError(t, err)

	var record models.AccountVerification
	err = db.Conn.Where("user_id = ?", user.Id).First(&record).Error
	assert.NoError(t, err)
	assert.Equal(t, "code123", record.VerificationCode)
}

func TestAccountRepository_RefreshVerificationCode_UpdateExisting(t *testing.T) {
	db, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("refresh-update"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "old"))

	assert.NoError(t, repo.RefreshVerificationCode(user, "newcode", time.Now().Add(time.Minute)))

	var record models.AccountVerification
	err := db.Conn.Where("user_id = ?", user.Id).First(&record).Error
	assert.NoError(t, err)
	assert.Equal(t, "newcode", record.VerificationCode)
}

func TestAccountRepository_UpdateRefreshTokenUpdatedAtAndAdminStatus(t *testing.T) {
	_, repo := setupTest(t)

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
	db, repo := setupTest(t)

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
	_, repo := setupTest(t)

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

func TestAccountRepository_VerifyAccount_InvalidCodeAttempts(t *testing.T) {
	db, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("attempts"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "correct"))

	for i := 0; i < 3; i++ {
		err := repo.VerifyAccount(user, "wrong", time.Now())
		assert.Error(t, err)
		_, notVerified := err.(*repositories.AccountNotVerifiedError)
		assert.True(t, notVerified)
	}

	var record models.AccountVerification
	err := db.Conn.Where("user_id = ?", user.Id).First(&record).Error
	assert.Error(t, err)
}

func TestAccountRepository_GetActivePasswordResetByUserID_NotFound(t *testing.T) {
	_, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("missing-reset"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	reset, err := repo.GetActivePasswordResetByUserID(user.Id)
	assert.Error(t, err)
	assert.Nil(t, reset)
	_, notFound := err.(*repositories.AccountNotFoundError)
	assert.True(t, notFound)
}

func TestAccountRepository_GetPasswordResetByTokenHash_NotFound(t *testing.T) {
	_, repo := setupTest(t)

	reset, err := repo.GetPasswordResetByTokenHash("missing")
	assert.Error(t, err)
	assert.Nil(t, reset)
	_, notFound := err.(*repositories.AccountNotFoundError)
	assert.True(t, notFound)
}

func TestAccountRepository_GetAccountByEmail_Found(t *testing.T) {
	_, repo := setupTest(t)

	email := repoTestEmail("found")
	user := &models.User{Email: email, Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	found, err := repo.GetAccountByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, email, found.Email)
}

func TestAccountRepository_GetAccountById_NotFound(t *testing.T) {
	_, repo := setupTest(t)

	result, err := repo.GetAccountById(uuid.New())
	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestAccountRepository_ListAccounts_NoFilter(t *testing.T) {
	db, repo := setupTest(t)
	_ = db

	for i := 0; i < 3; i++ {
		u := &models.User{Email: repoTestEmail("list-nofilter"), Password: "hashed"}
		assert.NoError(t, repo.CreateAccount(u, "code"))
	}

	users, total, err := repo.ListAccounts("", nil, 0, 100)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(3))
	assert.GreaterOrEqual(t, len(users), 3)
}

func TestAccountRepository_ListAccounts_VerifiedFalse(t *testing.T) {
	db, repo := setupTest(t)
	_ = db

	u := &models.User{Email: repoTestEmail("unverified"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(u, "code"))

	unverified := false
	users, total, err := repo.ListAccounts("", &unverified, 0, 100)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, total, int64(1))
	for _, usr := range users {
		assert.False(t, usr.Verified)
	}
}

func TestAccountRepository_ListAccounts_Pagination(t *testing.T) {
	db, repo := setupTest(t)
	_ = db

	for i := 0; i < 5; i++ {
		u := &models.User{Email: repoTestEmail("paginate"), Password: "hashed"}
		assert.NoError(t, repo.CreateAccount(u, "code"))
	}

	page1, total, err := repo.ListAccounts("paginate", nil, 0, 2)
	assert.NoError(t, err)
	assert.Equal(t, int64(5), total)
	assert.Len(t, page1, 2)

	page2, _, err := repo.ListAccounts("paginate", nil, 2, 2)
	assert.NoError(t, err)
	assert.Len(t, page2, 2)
}

func TestAccountRepository_UpdateAdminStatus_Toggle(t *testing.T) {
	_, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("toggle-admin"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	assert.NoError(t, repo.UpdateAdminStatus(user.Id, true))
	stored, err := repo.GetAccountById(user.Id)
	assert.NoError(t, err)
	assert.True(t, stored.IsAdmin)

	assert.NoError(t, repo.UpdateAdminStatus(user.Id, false))
	stored, err = repo.GetAccountById(user.Id)
	assert.NoError(t, err)
	assert.False(t, stored.IsAdmin)
}

func TestAccountRepository_IncrementPasswordResetAttempts_Multiple(t *testing.T) {
	_, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("inc-attempts"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	reset, err := repo.CreatePasswordReset(user.Id, "hash", time.Now().Add(time.Minute))
	assert.NoError(t, err)

	for i := 0; i < 3; i++ {
		assert.NoError(t, repo.IncrementPasswordResetAttempts(reset.Id))
	}

	active, err := repo.GetActivePasswordResetByUserID(user.Id)
	assert.NoError(t, err)
	assert.Equal(t, 3, active.Attempts)
}

func TestAccountRepository_UpdatePassword_ChangesHash(t *testing.T) {
	_, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("update-pwd"), Password: "original_hash"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	assert.NoError(t, repo.UpdatePassword(user.Id, "new_hash"))

	stored, err := repo.GetAccountById(user.Id)
	assert.NoError(t, err)
	assert.Equal(t, "new_hash", stored.Password)
}

func TestAccountRepository_VerifyAccount_WrongCode_ThenCorrect(t *testing.T) {
	_, repo := setupTest(t)

	email := repoTestEmail("wrong-then-correct")
	code := "rightcode"
	user := &models.User{Email: email, Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, code))

	// Un intento fallido
	err := repo.VerifyAccount(user, "wrongcode", time.Now())
	assert.Error(t, err)

	// El código correcto debe funcionar aun así
	userFromDb, err := repo.GetAccountByEmail(email)
	assert.NoError(t, err)
	err = repo.VerifyAccount(userFromDb, code, time.Now())
	assert.NoError(t, err)
	assert.True(t, userFromDb.Verified)
}

func TestAccountRepository_CreatePasswordReset_MultipleInvalidated(t *testing.T) {
	_, repo := setupTest(t)

	user := &models.User{Email: repoTestEmail("multi-reset"), Password: "hashed"}
	assert.NoError(t, repo.CreateAccount(user, "code"))

	reset1, err := repo.CreatePasswordReset(user.Id, "hash1", time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.NotNil(t, reset1)

	reset2, err := repo.CreatePasswordReset(user.Id, "hash2", time.Now().Add(time.Minute))
	assert.NoError(t, err)
	assert.NotNil(t, reset2)

	// Invalidar todos
	assert.NoError(t, repo.InvalidateAllPasswordResets(user.Id))

	// Ninguno debe ser activo
	active, err := repo.GetActivePasswordResetByUserID(user.Id)
	assert.Error(t, err)
	assert.Nil(t, active)
}
