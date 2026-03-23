package controllers

import (
	"net/http"

	"github.com/FeedTheRealm-org/core-service/internal/authentication-service/services"
	"github.com/FeedTheRealm-org/core-service/internal/utils/logger"
	"github.com/gin-gonic/gin"
)

type AdminLoginController struct {
	accountService services.AccountService
}

func NewAdminLoginController(accountService services.AccountService) *AdminLoginController {
	return &AdminLoginController{
		accountService: accountService,
	}
}

func (ac *AdminLoginController) AdminLoginPageHandler(c *gin.Context) {
	logger.Logger.Info("Serving admin login page")
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<form method="POST" action="/auth">
			<input type="email"    name="email"    placeholder="Email" />
			<input type="password" name="password" placeholder="Password" />
			<button type="submit">Login</button>
		</form>
	`))
}

func (ac *AdminLoginController) AdminLoginHandler(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	_, token, err := ac.accountService.LoginAccount(email, password)
	if err != nil {
		_ = c.Error(err)
	}

	c.SetCookie("jwt", token, 3600, "/swagger", "", true, true)
	c.Redirect(http.StatusFound, "/swagger/index.html")
}
