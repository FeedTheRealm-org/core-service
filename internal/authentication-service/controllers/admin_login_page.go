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

// AdminLoginPageHandler godoc
// @Summary      Admin login page
// @Description  Serves the HTML form for admin login.
// @Tags         authentication-service
// @Produce      text/html
// @Success      200  {string}  string "HTML form"
// @Router       /auth [get]
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

// AdminLoginHandler godoc
// @Summary      Admin login submission
// @Description  Processes admin login form submission, sets a JWT cookie, and redirects to Swagger UI.
// @Tags         authentication-service
// @Accept       x-www-form-urlencoded
// @Produce      text/html
// @Param        email formData string true "Admin Email"
// @Param        password formData string true "Admin Password"
// @Success      302  {string}  string "Redirect to /swagger/index.html"
// @Success      200  {string}  string "Fallback on error"
// @Router       /auth [post]
func (ac *AdminLoginController) AdminLoginHandler(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	_, token, err := ac.accountService.LoginAccount(email, password, false)
	if err != nil {
		_ = c.Error(err)
	}

	c.SetCookie("jwt", token, 3600, "/swagger", "", true, true)
	c.Redirect(http.StatusFound, "/swagger/index.html")
}
