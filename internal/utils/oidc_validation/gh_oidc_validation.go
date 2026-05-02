package oidc_validation

import (
	"errors"
	"strings"

	"github.com/FeedTheRealm-org/core-service/config"
	"github.com/MicahParks/keyfunc/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	GITHUB_ISSUER   = "https://token.actions.githubusercontent.com"
	GITHUB_JWKS_URL = "https://token.actions.githubusercontent.com/.well-known/jwks"
	MAIN_BRANCH_REF = "refs/heads/main"
)

type GitHubOIDCVerifier struct {
	audience string
	repo     string
	jwks     *keyfunc.JWKS
}

type GitHubClaims struct {
	jwt.RegisteredClaims
	BaseRef    string `json:"base_ref"`
	Repository string `json:"repository"`
	RefType    string `json:"ref_type"`
}

func NewGitHubOIDCVerifier(conf *config.Config) (*GitHubOIDCVerifier, error) {
	jwks, err := keyfunc.Get(GITHUB_JWKS_URL, keyfunc.Options{})
	if err != nil {
		return nil, err
	}

	return &GitHubOIDCVerifier{
		audience: conf.Github.GithubAudienceWebhook,
		repo:     conf.Github.GithubRepoURLWebhook,
		jwks:     jwks,
	}, nil
}

func (v *GitHubOIDCVerifier) IsValidToken(tokenString string) (GitHubClaims, error) {
	claims := &GitHubClaims{}

	token, err := jwt.ParseWithClaims(tokenString, claims, v.jwks.Keyfunc,
		jwt.WithAudience(v.audience),
		jwt.WithIssuer(GITHUB_ISSUER),
		jwt.WithExpirationRequired(),
	)
	if err != nil || !token.Valid {
		return *claims, errors.New("invalid token")
	}

	return *claims, nil
}

func (v *GitHubOIDCVerifier) IsValidRepo(claims *GitHubClaims) bool {
	return strings.EqualFold(claims.Repository, v.repo)
}

func (v *GitHubOIDCVerifier) IsTriggerATag(claims *GitHubClaims) bool {
	return claims.RefType == "tag"
}

func (v *GitHubOIDCVerifier) IsTriggerFromMain(claims *GitHubClaims) bool {
	return strings.EqualFold(claims.BaseRef, MAIN_BRANCH_REF)
}
