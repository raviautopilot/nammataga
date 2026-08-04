package router

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// RegisterStaticRoutes registers Swagger docs and static file serving endpoints
func RegisterStaticRoutes(r *gin.Engine) {
	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Static file serving - Images
	fmt.Println("✅ Serving /api/images from ./data/image")
	r.Static("/api/images", "./data/image")

	// PDF & Static document serving
	wd, _ := os.Getwd()
	docsPath := filepath.Join(wd, "data", "docs")

	docsHandler := func(c *gin.Context) {
		relPath := c.Param("filepath")
		fullPath := filepath.Join(docsPath, relPath)

		info, err := os.Stat(fullPath)
		if err != nil || info.IsDir() {
			c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
			return
		}

		if strings.ToLower(filepath.Ext(fullPath)) == ".pdf" {
			c.Writer.Header().Set("Content-Type", "application/pdf")
			c.Writer.Header().Set("Content-Disposition", "inline")
		}
		c.File(fullPath)
	}

	r.GET("/docs/*filepath", docsHandler)
	r.GET("/api/docs/*filepath", docsHandler)
	r.GET("/data/docs/*filepath", docsHandler)
	r.GET("/api/data/docs/*filepath", docsHandler)
}
