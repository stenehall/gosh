package web

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/contrib/secure"
	"github.com/gin-gonic/gin"
	"github.com/stenehall/gosh/internal/config"
)

const (
	CSP = "default-src 'unsafe-inline' 'self'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://cdn.tailwindcss.com"
)

// Ginning is the main entry for the web server functionality of Gosh.
func Ginning(gosh config.Gosh) error {
	env := os.Getenv("APP_ENV")
	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(secure.Secure(secure.Options{
		ContentSecurityPolicy: CSP,
	}))

	router.Static("/assets", "./assets")
	router.Static("/favicons", "./favicons")
	router.LoadHTMLGlob("web/*.gohtml")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.gohtml", gin.H{
			"title":      gosh.Title,
			"sets":       gosh.Sets,
			"showTitle":  gosh.ShowTitle,
			"background": gosh.Background,
		})
	})

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "OK",
		})
	})

	log.Printf("Serving gosh on port %d\n\n", gosh.Port)
	if err := router.Run(":" + strconv.Itoa(gosh.Port)); err != nil {
		return fmt.Errorf("error starting gin %w", err)
	}

	return nil
}
