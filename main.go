package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

// URLSet is the root element of the sitemap file
type URLSet struct {
	XMLName xml.Name `xml:"http://www.sitemaps.org/schemas/sitemap/0.9 urlset"`
	URLs    []URL    `xml:"url"`
}
type EmailConfig struct {
	SMTPHost string
	SMTPPort int
	Username string
	Password string
}

var emailConfig EmailConfig

// URL represents a single <url> entry in the sitemap
type URL struct {
	Loc        string  `xml:"loc"`
	LastMod    string  `xml:"lastmod"`
	ChangeFreq string  `xml:"changefreq"`
	Priority   float64 `xml:"priority"`
}

// Response structure from the FindIP API
type FindIPResponse struct {
	Country struct {
		IsoCode string `json:"iso_code"`
	} `json:"country"`
}

func main() {
	// Create a new Gin engine without default logger for performance

	// Initialize memcache client
	//mc = memcache.New("127.0.0.1:11211")

	router := gin.New()
	router.SetTrustedProxies([]string{"0.0.0.0/0"})
	router.Use(geoIPMiddleware)
	router.Use(gin.Recovery())            // Keep recovery middleware
	router.Use(gzip.Gzip(gzip.BestSpeed)) // Add gzip compression

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

	router.GET("/debug/ip", func(c *gin.Context) {
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			headers[k] = strings.Join(v, ", ")
		}

		log.Println("=== Incoming Request Headers ===")
		for k, v := range headers {
			log.Printf("%s: %s\n", k, v)
		}

		c.JSON(http.StatusOK, gin.H{
			"client_ip":       c.ClientIP(),
			"remote_addr":     c.Request.RemoteAddr,
			"x_forwarded_for": c.GetHeader("X-Forwarded-For"),
			"x_real_ip":       c.GetHeader("X-Real-IP"),
			"headers":         headers,
		})
	})
	// This route handles the incoming form data from the browser.
	router.POST("/contact-submit", func(c *gin.Context) {
		// Extract all the parameters from the form submission.
		// These names ("name", "email", etc.) MUST match the `name` attribute in your HTML form inputs.
		name := c.PostForm("name")
		email := c.PostForm("email")
		phone := c.PostForm("phone")
		country := c.PostForm("country")
		plan := c.PostForm("plan")
		bestTime := c.PostForm("calling-hours")

		// Call your sendEmail function with the extracted data.
		err := sendEmail(email, name, plan, country, bestTime, phone)

		// Check if the sendEmail function returned an error.
		if err != nil {
			// If there was an error, log it on the server and send an error response to the browser.
			log.Println("Error sending email:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Sorry, there was an error sending your message. Please try again later.",
			})
			return
		}

		// If everything was successful, send a success response back to the browser.
		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Your message has been sent successfully!",
		})
	})
	router.POST("/loss-recovery-submit", func(c *gin.Context) {
		// 1. Extract all the string parameters from the form.
		name := c.PostForm("name")
		email := c.PostForm("email")
		country := c.PostForm("country")
		whatsapp := c.PostForm("whatsapp")
		accountSize := c.PostForm("account-size")
		pastLoss := c.PostForm("past-loss")

		// 2. Handle the checkbox. A checkbox sends "on" if checked, and nothing if not.
		// We convert this to a true/false boolean.
		hasRunningTrades := c.PostForm("running-trades") == "on"

		// 3. Call your new sendLossRecoveryEmail function with the data.
		err := sendLossRecoveryEmail(name, email, country, whatsapp, accountSize, pastLoss, hasRunningTrades)

		// 4. Handle the response, just like the other form.
		if err != nil {
			log.Println("Error sending loss recovery email:", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  "error",
				"message": "Sorry, there was an error submitting your details. Please try again.",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":  "success",
			"message": "Your details have been submitted successfully!",
		})
	})

	// --- NEW: SITEMAP.XML HANDLER ---
	router.GET("/sitemap.xml", func(c *gin.Context) {
		baseURL := "https://fxopus.com"

		// Define all the URLs of your site
		pages := []URL{
			{Loc: baseURL + "/", ChangeFreq: "weekly", Priority: 1.0},
			{Loc: baseURL + "/about", ChangeFreq: "monthly", Priority: 0.8},
			{Loc: baseURL + "/pricing", ChangeFreq: "monthly", Priority: 0.9},
			{Loc: baseURL + "/payment", ChangeFreq: "yearly", Priority: 0.5},
			{Loc: baseURL + "/forex", ChangeFreq: "monthly", Priority: 0.7},
			{Loc: baseURL + "/commodities", ChangeFreq: "monthly", Priority: 0.7},
			{Loc: baseURL + "/crypto", ChangeFreq: "monthly", Priority: 0.7},
			{Loc: baseURL + "/loss-recovery", ChangeFreq: "monthly", Priority: 0.9},
			{Loc: baseURL + "/free-trial", ChangeFreq: "weekly", Priority: 0.9},
			{Loc: baseURL + "/blog", ChangeFreq: "weekly", Priority: 0.9},
			{Loc: baseURL + "/blog-post", ChangeFreq: "daily", Priority: 0.7}, // Example for a recent blog post
			// As you add more blog posts, you would dynamically add them to this list
		}

		// Get the current date for the LastMod field
		lastMod := time.Now().UTC().Format("2006-01-02")
		for i := range pages {
			pages[i].LastMod = lastMod
		}

		// Create the root URLSet element
		urlSet := URLSet{
			URLs: pages,
		}

		// Marshal the Go struct into XML
		xmlBytes, err := xml.MarshalIndent(urlSet, "", "  ")
		if err != nil {
			c.String(http.StatusInternalServerError, "Error generating sitemap")
			return
		}

		// Prepend the XML header
		sitemap := []byte(xml.Header + string(xmlBytes))

		// Serve the sitemap with the correct XML content type
		c.Data(http.StatusOK, "application/xml", sitemap)
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

	// Initialize email configuration using environment variables
	emailConfig = EmailConfig{
		SMTPHost: "smtp.zoho.in", // e.g., smtp.zoho.in
		SMTPPort: 587,            // Typical port for TLS
		Username: "contact@fxopus.com",
		Password: "bEqgys-xojrej-qowqu3",
	} //
}

// Function to get country code (with caching)
func getCountryCodeFromIP(ip string) (string, error) {
	cacheKey := fmt.Sprintf("country_%s", ip)
	const ipAPIEndpoint = "https://api.findip.net/%s/?token=d324417b69d24048a33807051cad5814"

	// Check cache
	if val, ok := ipCache.Load(cacheKey); ok {
		return val.(string), nil
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
	ipCache.Store(cacheKey, countryCode)
	return countryCode, nil
}

var ipCache sync.Map

func geoIPMiddleware(c *gin.Context) {

	log.Printf("ClientIP: %s, RemoteAddr: %s, X-Forwarded-For: %s\n",
		c.ClientIP(),
		c.Request.RemoteAddr,
		c.GetHeader("X-Forwarded-For"),
	)

	ip := c.GetHeader("X-Forwarded-For")
	if ip == "" {
		ip = c.ClientIP()
	} else {
		// If multiple IPs (comma-separated), get the first
		ip = strings.Split(ip, ",")[0]
	}
	log.Printf("[GeoBlock] Detected IP: %s", ip)
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
