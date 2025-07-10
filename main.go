package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// Response structure from the FindIP API
type FindIPResponse struct {
	Country struct {
		IsoCode string `json:"iso_code"`
	} `json:"country"`
}

var mc *memcache.Client

func main() {
	// Create a new Gin engine without default logger for performance

	// Initialize memcache client
	mc = memcache.New("127.0.0.1:11211")

	router := gin.New()
	router.Use(gin.Recovery())            // Keep recovery middleware
	router.Use(gzip.Gzip(gzip.BestSpeed)) // Add gzip compression
	router.Use(geoIPMiddleware)
	// Cache headers for static files
	router.Use(func(c *gin.Context) {
		if c.Request.Method == http.MethodGet {
			ext := strings.ToLower(filepath.Ext(c.Request.URL.Path))
			switch ext {
			case ".js", ".css", ".png", ".jpg", ".jpeg", ".gif", ".svg", ".webp":
				c.Header("Cache-Control", "public, max-age=31536000")
			}
		}
		c.Next()
	})

	// Serve static assets
	router.Static("/static", "./static")

	// Preload HTML templates once
	tmpl := template.Must(template.ParseGlob("templates/*.html"))
	router.SetHTMLTemplate(tmpl)

	// HTML redirect routes (e.g., /about.html → /about)
	pages := []string{
		"about.html", "pricing.html", "payment.html", "forex.html",
		"commodities.html", "crypto.html", "loss-recovery.html",
		"free-trial.html", "blog.html", "blog-post.html",
	}

	for _, page := range pages {
		router.GET("/"+page, func(c *gin.Context) {
			filename := path.Base(c.Request.URL.Path)
			cleanRoute := "/" + strings.TrimSuffix(filename, filepath.Ext(filename))
			c.Redirect(http.StatusMovedPermanently, cleanRoute)
		})
	}

	// Clean URL routes
	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})
	router.GET("/sorry-not-available", func(c *gin.Context) {
		c.HTML(http.StatusOK, "sorry.html", nil)
	})
	router.GET("/index.html", func(c *gin.Context) {
		rawURL := c.Request.URL.String()
		newPath := rawURL[len("/index.html"):]
		if newPath == "" {
			newPath = "/"
		}
		c.Redirect(http.StatusMovedPermanently, newPath)
	})

	router.GET("/about", func(c *gin.Context) { c.HTML(http.StatusOK, "about.html", nil) })
	router.GET("/pricing", func(c *gin.Context) { c.HTML(http.StatusOK, "pricing.html", nil) })
	router.GET("/payment", func(c *gin.Context) { c.HTML(http.StatusOK, "payment.html", nil) })
	router.GET("/forex", func(c *gin.Context) { c.HTML(http.StatusOK, "forex.html", nil) })
	router.GET("/commodities", func(c *gin.Context) { c.HTML(http.StatusOK, "commodities.html", nil) })
	router.GET("/crypto", func(c *gin.Context) { c.HTML(http.StatusOK, "crypto.html", nil) })
	router.GET("/loss-recovery", func(c *gin.Context) { c.HTML(http.StatusOK, "loss-recovery.html", nil) })
	router.GET("/free-trial", func(c *gin.Context) { c.HTML(http.StatusOK, "free-trial.html", nil) })
	router.GET("/blog", func(c *gin.Context) { c.HTML(http.StatusOK, "blog.html", nil) })
	router.GET("/blog-post", func(c *gin.Context) { c.HTML(http.StatusOK, "blog-post.html", nil) })

	// Example form submission
	router.POST("/contact-submit", func(c *gin.Context) {
		log.Println("Form submitted successfully.")
		c.String(http.StatusOK, "Thank you!")
	})

	// Get PORT from env (for Leapcell or Docker)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on port", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}

// Function to get country code (with caching)
func getCountryCodeFromIP(ip string) (string, error) {
	cacheKey := fmt.Sprintf("country_%s", ip)
	const ipAPIEndpoint = "https://api.findip.net/%s/?token=d324417b69d24048a33807051cad5814"

	// Check cache
	item, err := mc.Get(cacheKey)
	if err == nil {
		// Cache hit: return cached country code
		fmt.Printf("Cache hit: Retrieved country code for IP %s from cache\n", ip)
		return string(item.Value), nil
	} else {
		fmt.Printf("Cache miss: No cache data for IP %s, err: %s\n", ip, err.Error())
	}

	// Make API request
	url := fmt.Sprintf(ipAPIEndpoint, ip)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Decode JSON response using json.NewDecoder (Recommended in Go 1.16+)
	var ipData FindIPResponse
	if err := json.NewDecoder(resp.Body).Decode(&ipData); err != nil {
		return "", fmt.Errorf("failed to parse JSON: %v", err)
	}

	countryCode := strings.ToUpper(ipData.Country.IsoCode) // Normalize to uppercase

	// Store in cache
	err = mc.Set(&memcache.Item{Key: cacheKey, Value: []byte(countryCode), Expiration: 3600}) // 1 hour
	if err == nil {
		fmt.Printf("Cache write: Stored country code %s for IP %s\n", countryCode, ip)
	} else {
		fmt.Printf("Cache write failed for IP %s: %s\n", ip, err.Error())
	}

	return countryCode, nil
}

func geoIPMiddleware(c *gin.Context) {
	ip := c.ClientIP()

	countryCode, err := getCountryCodeFromIP(ip)
	if err != nil {
		log.Printf("Failed to get country code from IP %s: %v", ip, err)
		c.Next()
		return
	}

	if countryCode == "IN" {
		c.Redirect(http.StatusTemporaryRedirect, "/sorry-not-available")
		c.Abort()
		return
	}

	c.Next()
}
