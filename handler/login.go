package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"github.com/wei840222/login-server/config"
)

type LoginHandler struct {
	client *resty.Client
}

func (h *LoginHandler) ArgoCD(c *gin.Context) {
	if token, err := c.Cookie("argocd.token"); err == nil {
		res, err := h.client.R().
			SetCookie(&http.Cookie{
				Name:  "argocd.token",
				Value: token,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyArgoCDServerURL), "/api/v1/session/userinfo"))
		if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
			return
		}
	}

	res, err := h.client.R().
		SetBody(map[string]string{
			"username": viper.GetString(config.ConfigKeyArgoCDUsername),
			"password": viper.GetString(config.ConfigKeyArgoCDPassword),
		}).
		Post(JoinURL(viper.GetString(config.ConfigKeyArgoCDServerURL), "/api/v1/session"))
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
		return
	}
	if res.IsError() {
		err := fmt.Errorf("failed to login to argo-cd: %s %s", res.Status(), res)
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
		return
	}

	c.Header("Set-Cookie", res.Header().Get("Set-Cookie"))
	c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
}

func (h *LoginHandler) Ghost(c *gin.Context) {
	if session, err := c.Cookie("ghost-admin-api-session"); err == nil {
		res, err := h.client.R().
			SetHeader("X-Forwarded-Proto", "https").
			SetCookie(&http.Cookie{
				Name:  "ghost-admin-api-session",
				Value: session,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyGhostServerURL), "/ghost/api/admin/users/me/"))
		if err == nil && res.IsSuccess() {
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/ghost"))
			return
		}
	}

	res, err := h.client.R().
		SetHeader("X-Forwarded-Proto", "https").
		SetBody(map[string]string{
			"username": viper.GetString(config.ConfigKeyGhostUsername),
			"password": viper.GetString(config.ConfigKeyGhostPassword),
		}).
		Post(JoinURL(viper.GetString(config.ConfigKeyGhostServerURL), "/ghost/api/admin/session"))
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
		return
	}
	if res.IsError() {
		err := fmt.Errorf("failed to login to ghost: %s %s", res.Status(), res)
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
		return
	}

	c.Header("Set-Cookie", res.Header().Get("Set-Cookie"))
	c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/ghost"))
}

func (h *LoginHandler) N8N(c *gin.Context) {
	if auth, err := c.Cookie("n8n-auth"); err == nil {
		res, err := h.client.R().
			SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
			SetCookie(&http.Cookie{
				Name:  "n8n-auth",
				Value: auth,
			}).
			Get("/rest/login")
		if err == nil && res.IsSuccess() {
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
			return
		}
	}

	res, err := h.client.R().
		SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
		SetBody(map[string]string{
			"email":    viper.GetString(config.ConfigKeyN8NUsername),
			"password": viper.GetString(config.ConfigKeyN8NPassword),
		}).
		Post(JoinURL(viper.GetString(config.ConfigKeyN8NServerURL), "/rest/login"))
	if err != nil {
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
		return
	}
	if res.IsError() {
		err := fmt.Errorf("failed to login to n8n: %s %s", res.Status(), res)
		c.Error(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
		return
	}

	c.Header("Set-Cookie", res.Header().Get("Set-Cookie"))
	c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
}

func RegisterLoginHandler(e *gin.Engine) {
	h := &LoginHandler{
		client: resty.New(),
	}

	login := e.Group("/login")
	{
		login.GET("/argo-cd", h.ArgoCD)
		login.GET("/ghost", h.Ghost)
		login.GET("/n8n", h.N8N)
	}
}
