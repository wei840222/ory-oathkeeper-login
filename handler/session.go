package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"

	"github.com/wei840222/login-server/config"
)

type OrySession struct {
	Subject string `json:"subject"`
	Extra   struct {
		Email string `json:"email"`
	} `json:"extra"`
}

type SessionHandler struct {
	client *resty.Client
	cache  cache.CacheInterface[string]
}

func (h *SessionHandler) ArgoCD(c *gin.Context) {
	if sessionKey, err := c.Cookie("argocd.token"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("argo-cd:%s", sessionKey))
		if err == nil {
			log.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				log.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("argo-cd:%s", sessionKey)); err != nil {
					log.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			log.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().
			SetCookie(&http.Cookie{
				Name:  "argocd.token",
				Value: sessionKey,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyArgoCDServerURL), "/api/v1/session/userinfo"))

		if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
			var session OrySession
			session.Subject = gjson.GetBytes(res.Body(), "username").String()
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("argo-cd:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.ConfigKeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, ErrorRes{Error: "invalid session"})
}

func (h *SessionHandler) Ghost(c *gin.Context) {
	if sessionKey, err := c.Cookie("ghost-admin-api-session"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("ghost:%s", sessionKey))
		if err == nil {
			log.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				log.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("ghost:%s", sessionKey)); err != nil {
					log.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			log.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().
			SetHeaders(map[string]string{
				"X-Forwarded-Proto": "https",
				"Origin":            viper.GetString(config.ConfigKeyGhostOriginURL),
			}).
			SetCookie(&http.Cookie{
				Name:  "ghost-admin-api-session",
				Value: sessionKey,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyGhostServerURL), "/ghost/api/admin/users/me/"))

		if err == nil && res.IsSuccess() {
			var session OrySession
			session.Subject = gjson.GetBytes(res.Body(), "users.0.id").String()
			session.Extra.Email = gjson.GetBytes(res.Body(), "users.0.email").String()
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("ghost:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.ConfigKeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, ErrorRes{Error: "invalid session"})
}

func (h *SessionHandler) N8N(c *gin.Context) {
	if sessionKey, err := c.Cookie("n8n-auth"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("n8n:%s", sessionKey))
		if err == nil {
			log.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				log.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("n8n:%s", sessionKey)); err != nil {
					log.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			log.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().
			SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
			SetCookie(&http.Cookie{
				Name:  "n8n-auth",
				Value: sessionKey,
			}).
			Get(JoinURL(viper.GetString(config.ConfigKeyN8NServerURL), "/rest/login"))

		if err == nil && res.IsSuccess() {
			var session OrySession
			session.Subject = gjson.GetBytes(res.Body(), "data.id").String()
			session.Extra.Email = gjson.GetBytes(res.Body(), "data.email").String()
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("n8n:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.ConfigKeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, ErrorRes{Error: "invalid session"})
}

func RegisterSessionHandler(e *gin.Engine, c cache.CacheInterface[string]) {
	h := &SessionHandler{
		client: resty.New(),
		cache:  c,
	}

	session := e.Group("/session")
	{
		session.GET("/argo-cd", h.ArgoCD)
		session.GET("/ghost", h.Ghost)
		session.GET("/n8n", h.N8N)
	}
}
