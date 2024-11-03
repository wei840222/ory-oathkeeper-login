package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/go-resty/resty/v2"
	"github.com/tidwall/gjson"
)

func main() {
	argoCDServer := resty.New().
		SetDebug(true).
		SetBaseURL(os.Getenv("ARGO_CD_SERVER_URL"))
	ghostServer := resty.New().
		SetDebug(true).
		SetBaseURL(os.Getenv("GHOST_SERVER_URL"))

	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	session := r.Group("/session")
	{
		session.GET("/argo-cd", func(c *gin.Context) {
			if token, err := c.Cookie("argocd.token"); err == nil {
				res, err := argoCDServer.R().
					SetCookie(&http.Cookie{
						Name:  "argocd.token",
						Value: token,
					}).
					Get("/api/v1/session/userinfo")
				if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
					c.JSON(http.StatusOK, gin.H{
						"subject": gjson.GetBytes(res.Body(), "username").String(),
						"extra":   gin.H{},
					})
					return
				}
			}
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		})

		session.GET("/ghost", func(c *gin.Context) {
			if session, err := c.Cookie("ghost-admin-api-session"); err == nil {
				res, err := ghostServer.R().
					SetCookie(&http.Cookie{
						Name:  "ghost-admin-api-session",
						Value: session,
					}).
					Get("/ghost/api/admin/users/me")
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
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid session"})
		})
	}

	login := r.Group("/login")
	{
		login.GET("/argo-cd", func(c *gin.Context) {
			if token, err := c.Cookie("argocd.token"); err == nil {
				res, err := argoCDServer.R().
					SetCookie(&http.Cookie{
						Name:  "argocd.token",
						Value: token,
					}).
					Get("/api/v1/session/userinfo")
				if err == nil && res.IsSuccess() && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
					c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
					return
				}
			}

			res, err := argoCDServer.R().
				SetBody(map[string]string{
					"username": os.Getenv("ARGO_CD_USERNAME"),
					"password": os.Getenv("ARGO_CD_PASSWORD"),
				}).
				Post("/api/v1/session")
			if err != nil {
				panic(err)
			}
			if res.IsError() {
				panic(fmt.Sprintf("failed to login to argo-cd: %s %s", res.Status(), res))
			}

			c.Header("Set-Cookie", res.Header().Get("Set-Cookie"))
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
		})

		login.GET("/ghost", func(c *gin.Context) {
			if session, err := c.Cookie("ghost-admin-api-session"); err == nil {
				res, err := ghostServer.R().
					SetHeader("X-Forwarded-Proto", "https").
					SetCookie(&http.Cookie{
						Name:  "ghost-admin-api-session",
						Value: session,
					}).
					Get("/ghost/api/admin/users/me/")
				if err == nil && res.IsSuccess() {
					c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/ghost"))
					return
				}
			}

			res, err := ghostServer.R().
				SetHeader("X-Forwarded-Proto", "https").
				SetBody(map[string]string{
					"username": os.Getenv("GHOST_USERNAME"),
					"password": os.Getenv("GHOST_PASSWORD"),
				}).
				Post("/ghost/api/admin/session")
			if err != nil {
				panic(err)
			}
			if res.IsError() {
				panic(fmt.Sprintf("failed to login to ghost: %s %s", res.Status(), res))
			}

			c.Header("Set-Cookie", res.Header().Get("Set-Cookie"))
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/ghost"))
		})
	}

	r.Run()
}
