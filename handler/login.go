package handler

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptrace"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"github.com/wei840222/ory-oathkeeper-login/config"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type LoginHandler struct {
	client *resty.Client
}

func (h *LoginHandler) ArgoCD(c *gin.Context) {
	if token, err := c.Cookie("argocd.token"); err == nil {
		res, err := h.client.R().SetContext(c).
			SetCookie(&http.Cookie{
				Name:  "argocd.token",
				Value: token,
			}).
			Get(JoinURL(viper.GetString(config.KeyArgoCDServerURL), "/api/v1/session/userinfo"))
		if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
			return
		}
	}

	res, err := h.client.R().SetContext(c).
		SetBody(map[string]string{
			"username": viper.GetString(config.KeyArgoCDUsername),
			"password": viper.GetString(config.KeyArgoCDPassword),
		}).
		Post(JoinURL(viper.GetString(config.KeyArgoCDServerURL), "/api/v1/session"))
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
		res, err := h.client.R().SetContext(c).
			SetHeaders(map[string]string{
				"X-Forwarded-Proto": "https",
				"Origin":            viper.GetString(config.KeyGhostOriginURL),
			}).
			SetCookie(&http.Cookie{
				Name:  "ghost-admin-api-session",
				Value: session,
			}).
			Get(JoinURL(viper.GetString(config.KeyGhostServerURL), "/ghost/api/admin/users/me/"))
		if err == nil && res.IsSuccess() {
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/ghost"))
			return
		}
	}

	res, err := h.client.R().SetContext(c).
		SetHeaders(map[string]string{
			"X-Forwarded-Proto": "https",
			"Origin":            viper.GetString(config.KeyGhostOriginURL),
		}).
		SetBody(map[string]string{
			"username": viper.GetString(config.KeyGhostUsername),
			"password": viper.GetString(config.KeyGhostPassword),
		}).
		Post(JoinURL(viper.GetString(config.KeyGhostServerURL), "/ghost/api/admin/session"))
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
		res, err := h.client.R().SetContext(c).
			SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
			SetCookie(&http.Cookie{
				Name:  "n8n-auth",
				Value: auth,
			}).
			Get(JoinURL(viper.GetString(config.KeyN8NServerURL), "/rest/login"))
		if err == nil && res.IsSuccess() {
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
			return
		}
	}

	res, err := h.client.R().SetContext(c).
		SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
		SetBody(map[string]string{
			"emailOrLdapLoginId": viper.GetString(config.KeyN8NUsername),
			"password":           viper.GetString(config.KeyN8NPassword),
		}).
		Post(JoinURL(viper.GetString(config.KeyN8NServerURL), "/rest/login"))
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
		client: resty.NewWithClient(&http.Client{
			Transport: otelhttp.NewTransport(
				http.DefaultTransport,
				otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
					return otelhttptrace.NewClientTrace(ctx)
				}),
			),
		}),
	}

	login := e.Group("/login")
	{
		login.GET("/argo-cd", h.ArgoCD)
		login.GET("/ghost", h.Ghost)
		login.GET("/n8n", h.N8N)
	}
}
