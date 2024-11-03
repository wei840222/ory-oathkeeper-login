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
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	login := r.Group("/login")
	{
		login.GET("/", func(c *gin.Context) {
			returnURL := c.GetHeader("X-Login-Server-Path")
			if returnURL == "" {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "required X-Login-Server-Path header"})
				return
			}
			c.Redirect(http.StatusFound, returnURL)
		})

		argoCDServer := resty.New().
			SetDebug(true).
			SetBaseURL(os.Getenv("ARGO_CD_SERVER_URL"))
		login.GET("/argo-cd", func(c *gin.Context) {
			if token, err := c.Cookie("argocd.token"); err == nil {
				res, err := argoCDServer.R().
					SetCookie(&http.Cookie{
						Name:  "argocd.token",
						Value: token,
					}).
					Get("/api/v1/session/userinfo")
				if err == nil && res.StatusCode() == http.StatusOK && gjson.GetBytes(res.Body(), "loggedIn").Bool() {
					c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
					return
				}
			}

			res, err := argoCDServer.R().
				SetBody(map[string]string{
					"username": "admin",
					"password": os.Getenv("ARGO_CD_ADMIN_PASSWORD"),
				}).
				Post("/api/v1/session")
			if err != nil {
				panic(err)
			}
			if res.IsError() {
				panic(fmt.Sprintf("failed to login to argo cd: %s %s", res.Status(), res))
			}

			c.SetCookie("argocd.token", gjson.GetBytes(res.Body(), "token").String(), 0, "/", "", true, true)
			c.Redirect(http.StatusFound, c.DefaultQuery("return_url", "/"))
		})
	}

	r.Run()
}
