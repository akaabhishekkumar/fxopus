package main

import (
	"log"
	"net/http"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize Gin
	router := gin.Default()

	// Tell Gin where to find your static files (CSS, JS, etc.)
	router.Static("/static", "./static")

	// Tell Gin to load all your HTML files from the 'templates' folder.
	router.LoadHTMLGlob("templates/*")

	// --- SETUP THE .HTML REDIRECT ROUTES ---
	// This list contains all your pages that should have redirects.
	pages := []string{
		"about.html",
		"pricing.html",
		"payment.html",
		"forex.html",
		"commodities.html",
		"crypto.html",
		"loss-recovery.html",
		"free-trial.html",
		"blog.html",
		"blog-post.html",
	}

	for _, page := range pages {
		// For each page, create a GET route that matches the .html filename
		router.GET("/"+page, func(c *gin.Context) {
			// Get the requested filename (e.g., "about.html")
			filename := path.Base(c.Request.URL.Path)
			// Remove the .html extension to get the clean route (e.g., "about")
			newPath := "/" + filename[:len(filename)-len(path.Ext(filename))]

			// Issue a permanent redirect to the clean URL
			c.Redirect(http.StatusMovedPermanently, newPath)
		})
	}

	// --- SETUP THE OFFICIAL PAGE-SERVING ROUTES ---

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.GET("/index.html", func(c *gin.Context) {
		// Get the raw URL from the original request, which includes the query and fragment
		rawURL := c.Request.URL.String()

		// The new path is just the part of the string after "/index.html"
		// For "/index.html#contact", this will result in "/#contact"
		// For "/index.html", this will result in "/"
		newPath := rawURL[len("/index.html"):]
		if newPath == "" {
			newPath = "/"
		}

		c.Redirect(http.StatusMovedPermanently, newPath)
	})

	//router.GET("/index", func(c *gin.Context) {
	//	c.Redirect(http.StatusMovedPermanently, "/")
	//})
	router.GET("/about", func(c *gin.Context) {
		c.HTML(http.StatusOK, "about.html", nil)
	})

	router.GET("/pricing", func(c *gin.Context) {
		c.HTML(http.StatusOK, "pricing.html", nil)
	})

	// Add a clean route for every other page...
	router.GET("/payment", func(c *gin.Context) { c.HTML(http.StatusOK, "payment.html", nil) })
	router.GET("/forex", func(c *gin.Context) { c.HTML(http.StatusOK, "forex.html", nil) })
	router.GET("/commodities", func(c *gin.Context) { c.HTML(http.StatusOK, "commodities.html", nil) })
	router.GET("/crypto", func(c *gin.Context) { c.HTML(http.StatusOK, "crypto.html", nil) })
	router.GET("/loss-recovery", func(c *gin.Context) { c.HTML(http.StatusOK, "loss-recovery.html", nil) })
	router.GET("/free-trial", func(c *gin.Context) { c.HTML(http.StatusOK, "free-trial.html", nil) })
	router.GET("/blog", func(c *gin.Context) { c.HTML(http.StatusOK, "blog.html", nil) })
	router.GET("/blog-post", func(c *gin.Context) { c.HTML(http.StatusOK, "blog-post.html", nil) })

	// --- Your Backend Logic Routes ---
	router.POST("/contact-submit", func(c *gin.Context) {
		// Your form handling logic...
		log.Println("Form submitted successfully.")
		c.String(http.StatusOK, "Thank you!")
	})

	// Start the server
	//log.Println("Server starting on http://localhost:8080")
	//	router.Run(":8090")

	// Get the port from the environment variable provided by Leapcell
	port := os.Getenv("PORT")
	// If the PORT environment variable is not set, default to 8080 for local development
	if port == "" {
		port = "8080"
	}

	// Start the server on the correct port
	log.Println("Server starting on port " + port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Server failed to start:", err)
	}

}
