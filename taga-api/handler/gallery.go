package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"

	"taga-api/model"

	"github.com/gin-gonic/gin"
)

const (
	galleryFile = "data/gallery.json"
)

func loadGallery() ([]model.GalleryImage, error) {

	file, err := os.ReadFile(galleryFile)
	if err != nil {
		return nil, err
	}

	var images []model.GalleryImage
	err = json.Unmarshal(file, &images)

	if images == nil {
		images = []model.GalleryImage{}
	}

	return images, err
}

// GalleryYearsHandler godoc
// @Summary Get gallery years
// @Description Returns available gallery years
// @Tags gallery
// @Produce json
// @Success 200 {array} int
// @Router /api/gallery/years [get]
func GalleryYearsHandler(c *gin.Context) {

	images, err := loadGallery()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cannot read gallery",
		})
		return
	}

	yearMap := map[int]bool{}

	for _, img := range images {
		yearMap[img.Year] = true
	}

	var years []int

	for year := range yearMap {
		years = append(years, year)
	}

	// Sort years descending (latest first)
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	c.JSON(http.StatusOK, years)
}

// GalleryImagesHandler godoc
// @Summary Get gallery images by year
// @Description Returns gallery images for selected year
// @Tags gallery
// @Produce json
// @Param year query int false "Gallery Year (default: current year)"
// @Success 200 {array} model.GalleryImage
// @Router /api/gallery [get]
func GalleryImagesHandler(c *gin.Context) {

	yearParam := c.Query("year")

	var year int
	var err error

	if yearParam == "" {
		year = time.Now().Year()
	} else {
		year, err = strconv.Atoi(yearParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "invalid year",
			})
			return
		}
	}

	images, err := loadGallery()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cannot read gallery",
		})
		return
	}

	result := []model.GalleryImage{}

	for _, img := range images {
		if img.Year == year {
			result = append(result, img)
		}
	}

	c.JSON(http.StatusOK, result)
}
