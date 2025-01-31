package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"

	"github.com/wei840222/login-server/config"
)

type SessionHandler struct {
	client *resty.Client
}

func (h *SessionHandler) ArgoCD(c *gin.Context) {
	if token, err := c.Cookie("argocd.token"); err == nil {
		res, err := h.client.R().
			SetCookie(&http.Cookie{
				Name:  "argocd.token",
				Value: token,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyArgoCDServerURL), "/api/v1/session/userinfo"))
		if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
			c.JSON(http.StatusOK, gin.H{
				"subject": gjson.GetBytes(res.Body(), "username").String(),
				"extra": map[string]any{
					"email": gjson.GetBytes(res.Body(), "email").String(),
				},
			})
			return
		}
	}
	c.JSON(http.StatusUnauthorized, ErrorRes{Error: "invalid session"})
}

func (h *SessionHandler) Ghost(c *gin.Context) {
	if session, err := c.Cookie("ghost-admin-api-session"); err == nil {
		res, err := h.client.R().
			SetHeader("X-Forwarded-Proto", "https").
			SetCookie(&http.Cookie{
				Name:  "ghost-admin-api-session",
				Value: session,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyGhostServerURL), "/ghost/api/admin/users/me/"))
		if err == nil && res.IsSuccess() {
			c.JSON(http.StatusOK, gin.H{
				"subject": gjson.GetBytes(res.Body(), "users.0.id").String(),
				"extra": gin.H{
					"email": gjson.GetBytes(res.Body(), "users.0.email").String(),
				},
			})
			return
		}
	}
	c.JSON(http.StatusUnauthorized, ErrorRes{Error: "invalid session"})
}

func (h *SessionHandler) N8N(c *gin.Context) {
	if auth, err := c.Cookie("n8n-auth"); err == nil {
		res, err := h.client.R().
			SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
			SetCookie(&http.Cookie{
				Name:  "n8n-auth",
				Value: auth,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyN8NServerURL), "/rest/login"))
		if err == nil && res.IsSuccess() {
			c.JSON(http.StatusOK, gin.H{
				"subject": gjson.GetBytes(res.Body(), "data.id").String(),
				"extra": gin.H{
					"email": gjson.GetBytes(res.Body(), "data.email").String(),
				},
			})
			return
		}
	}
	c.JSON(http.StatusUnauthorized, ErrorRes{Error: "invalid session"})
}

func RegisterSessionHandler(e *gin.Engine) {
	h := &SessionHandler{
		client: resty.New(),
	}

	session := e.Group("/session")
	{
		session.GET("/argo-cd", h.ArgoCD)
		session.GET("/ghost", h.Ghost)
		session.GET("/n8n", h.N8N)
	}
}
