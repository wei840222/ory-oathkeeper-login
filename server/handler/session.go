package handler

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptrace"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/tidwall/gjson"
	"go.opentelemetry.io/contrib/instrumentation/net/http/httptrace/otelhttptrace"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/wei840222/ory-oathkeeper-login/config"
	"github.com/wei840222/ory-oathkeeper-login/server"
)

type OrySession struct {
	Subject string `json:"subject"`
	Extra   struct {
		Email string `json:"email"`
	} `json:"extra"`
}

type SessionHandler struct {
	logger zerolog.Logger
	client *resty.Client
	cache  cache.CacheInterface[string]
}

func (h *SessionHandler) Proxmox(c *gin.Context) {
	if ticket, err := c.Request.Cookie("PVEAuthCookie"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("proxmox:%s", ticket.Value))
		if err == nil {
			h.logger.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				h.logger.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("proxmox:%s", ticket.Value)); err != nil {
					h.logger.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			h.logger.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().SetContext(c).
			SetCookie(&http.Cookie{
				Name:  "PVEAuthCookie",
				Value: ticket.Value,
			}).
			Get(JoinURL(viper.GetString(config.KeyProxmoxServerURL), "/api2/extjs/version"))

		if err == nil && res.IsSuccess() {
			var session OrySession
			session.Subject = viper.GetString(config.KeyProxmoxUsername)
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("proxmox:%s", ticket.Value), string(b), store.WithExpiration(viper.GetDuration(config.KeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, server.ErrorRes{Error: server.ErrInvalidSession.Error()})
}

func (h *SessionHandler) ArgoCD(c *gin.Context) {
	if sessionKey, err := c.Cookie("argocd.token"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("argo-cd:%s", sessionKey))
		if err == nil {
			h.logger.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				h.logger.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("argo-cd:%s", sessionKey)); err != nil {
					h.logger.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			h.logger.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().SetContext(c).
			SetCookie(&http.Cookie{
				Name:  "argocd.token",
				Value: sessionKey,
			}).
			Get(JoinURL(viper.GetString(config.KeyArgoCDServerURL), "/api/v1/session/userinfo"))

		if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
			var session OrySession
			session.Subject = gjson.GetBytes(res.Body(), "username").String()
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("argo-cd:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.KeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, server.ErrorRes{Error: server.ErrInvalidSession.Error()})
}

func (h *SessionHandler) Ghost(c *gin.Context) {
	if sessionKey, err := c.Cookie("ghost-admin-api-session"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("ghost:%s", sessionKey))
		if err == nil {
			h.logger.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				h.logger.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("ghost:%s", sessionKey)); err != nil {
					h.logger.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			h.logger.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().SetContext(c).
			SetHeaders(map[string]string{
				"X-Forwarded-Proto": "https",
				"Origin":            viper.GetString(config.KeyGhostOriginURL),
			}).
			SetCookie(&http.Cookie{
				Name:  "ghost-admin-api-session",
				Value: sessionKey,
			}).
			Get(JoinURL(viper.GetString(config.KeyGhostServerURL), "/ghost/api/admin/users/me/"))

		if err == nil && res.IsSuccess() {
			var session OrySession
			session.Subject = gjson.GetBytes(res.Body(), "users.0.id").String()
			session.Extra.Email = gjson.GetBytes(res.Body(), "users.0.email").String()
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("ghost:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.KeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, server.ErrorRes{Error: server.ErrInvalidSession.Error()})
}

func (h *SessionHandler) N8N(c *gin.Context) {
	if sessionKey, err := c.Cookie("n8n-auth"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("n8n:%s", sessionKey))
		if err == nil {
			h.logger.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				h.logger.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("n8n:%s", sessionKey)); err != nil {
					h.logger.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			h.logger.Debug().Err(err).Msg("session cache miss")
		}

		res, err := h.client.R().SetContext(c).
			SetHeader("Browser-Id", c.GetHeader("Browser-Id")).
			SetCookie(&http.Cookie{
				Name:  "n8n-auth",
				Value: sessionKey,
			}).
			Get(JoinURL(viper.GetString(config.KeyN8NServerURL), "/rest/login"))

		if err == nil && res.IsSuccess() {
			var session OrySession
			session.Subject = gjson.GetBytes(res.Body(), "data.id").String()
			session.Extra.Email = gjson.GetBytes(res.Body(), "data.email").String()
			b, err := json.Marshal(session)
			if err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			if err := h.cache.Set(c, fmt.Sprintf("n8n:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.KeyCacheTTL))); err != nil {
				c.Error(err)
				c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
				return
			}

			c.JSON(http.StatusOK, session)
			return
		}
	}

	c.JSON(http.StatusUnauthorized, server.ErrorRes{Error: server.ErrInvalidSession.Error()})
}

func (h *SessionHandler) NocoDB(c *gin.Context) {
	if sessionKey, err := c.Cookie("refresh_token"); err == nil {
		s, err := h.cache.Get(c, fmt.Sprintf("nocodb:%s", sessionKey))
		if err == nil {
			h.logger.Debug().Str("session", s).Msg("session cache hit")
			var session OrySession
			if err := json.Unmarshal([]byte(s), &session); err == nil {
				c.JSON(http.StatusOK, session)
				return
			} else {
				h.logger.Warn().Err(err).Msg("session cache hit but unmarshal failed")
				if err := h.cache.Delete(c, fmt.Sprintf("nocodb:%s", sessionKey)); err != nil {
					h.logger.Warn().Err(err).Msg("session cache delete failed")
				}
			}
		} else {
			h.logger.Debug().Err(err).Msg("session cache miss")
		}

		var session OrySession
		session.Subject = viper.GetString(config.KeyNocoDBUsername)
		b, err := json.Marshal(session)
		if err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
			return
		}

		if err := h.cache.Set(c, fmt.Sprintf("nocodb:%s", sessionKey), string(b), store.WithExpiration(viper.GetDuration(config.KeyCacheTTL))); err != nil {
			c.Error(err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, server.ErrorRes{Error: err.Error()})
			return
		}

		c.JSON(http.StatusOK, session)
		return
	}

	c.JSON(http.StatusUnauthorized, server.ErrorRes{Error: server.ErrInvalidSession.Error()})
}

func RegisterSessionHandler(e *gin.Engine, c cache.CacheInterface[string]) {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	h := &SessionHandler{
		logger: log.With().Str("logger", "sessionHandler").Logger(),
		client: resty.NewWithClient(&http.Client{
			Transport: otelhttp.NewTransport(
				customTransport,
				otelhttp.WithClientTrace(func(ctx context.Context) *httptrace.ClientTrace {
					return otelhttptrace.NewClientTrace(ctx)
				}),
			),
		}),
		cache: c,
	}

	session := e.Group("/session")
	{
		session.GET("/proxmox", h.Proxmox)
		session.GET("/argo-cd", h.ArgoCD)
		session.GET("/ghost", h.Ghost)
		session.GET("/n8n", h.N8N)
		session.GET("/nocodb", h.NocoDB)
	}
}
