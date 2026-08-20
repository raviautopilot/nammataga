package handler

import (
	"encoding/csv"
	"net/http"
	"os"
	"taga-api/config"
	"taga-api/service"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// GetResourceCategories godoc
// @Summary Get all resource categories
// @Description Returns list of resource categories (without documents)
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Success 200 {array} model.ResourceCategory
// @Failure 500 {object} map[string]string
// @Router /api/resources [get]
func GetResourceCategories(c *gin.Context) {
	data, err := service.LoadResources()
	if err != nil {
		config.Logger.Error("ERROR in GetResourceCategories", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return only id & name
	var categories []gin.H
	for _, item := range data {
		categories = append(categories, gin.H{
			"id":   item.ID,
			"name": item.Name,
		})
	}

	c.JSON(http.StatusOK, categories)
}

// GetAllResources godoc
// @Summary Get all resources
// @Description Returns all resource categories along with their documents
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Success 200 {array} service.ResourceCategory
// @Failure 500 {object} map[string]string
// @Router /api/resources/all [get]
func GetAllResources(c *gin.Context) {
	data, err := service.LoadResources()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, data)
}

// GetDocumentsByCategory godoc
// @Summary Get documents by category
// @Description Returns all documents for a given resource category ID
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Param id path string true "Category ID"
// @Param subcategory query string false "Optional subcategory: Central or State"
// @Success 200 {array} model.Document
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/resources/{id} [get]
func GetDocumentsByCategory(c *gin.Context) {
	id := c.Param("id")
	subcategory := c.Query("subcategory") // new query param

	data, err := service.LoadResources()
	if err != nil {
		config.Logger.Error("ERROR in GetDocumentsByCategory", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, item := range data {
		if item.ID == id {
			// Filter documents by subcategory if provided
			var filtered []service.Document
			if subcategory != "" {
				for _, doc := range item.Documents {
					if doc.Subcategory == subcategory {
						filtered = append(filtered, doc)
					}
				}
				c.JSON(http.StatusOK, filtered)
				return
			}

			// If no subcategory, return all documents
			c.JSON(http.StatusOK, item.Documents)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
}

// GetResourcesBanner godoc
// @Summary Get resources banner image info
// @Description Get the relative path/url of the resources banner image
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string
// @Router /api/resources-banner [get]
func GetResourcesBanner(c *gin.Context) {
	c.JSON(200, gin.H{
		"image": "/api/images/resources-banner.jpg",
	})
}

// GetExternalLinks godoc
// @Summary Get external links from CSV
// @Description Returns list of external links (title + URL) from CSV file
// @Tags Resources
// @Produce json
// @Security BearerAuth
// @Success 200 {array} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/resources/external-links [get]
func GetExternalLinks(c *gin.Context) {
	filePath := "data/docs/External Links.csv"

	file, err := os.Open(filePath)
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to open CSV"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		c.JSON(500, gin.H{"error": "Unable to read CSV"})
		return
	}

	var links []gin.H

	for i, row := range records {
		if i == 0 {
			continue // skip header
		}

		if len(row) < 2 {
			continue
		}

		links = append(links, gin.H{
			"title": row[0],
			"url":   row[1],
		})
	}

	c.JSON(200, links)
}
