package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"sort"
	"strconv"
	"taga-api/model"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	eventsFile  = "data/events.json"
	galleryFile = "data/gallery.json"
)

// UpcomingEventsHandler retrieves upcoming events
// @Summary Get upcoming events
// @Tags Events
// @Produce json
// @Success 200 {array} model.Event
// @Router /api/events/upcoming [get]
func UpcomingEventsHandler(c *gin.Context) {
	file, err := os.ReadFile(eventsFile)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "cannot read events")
		return
	}

	var events []model.Event
	if err = json.Unmarshal(file, &events); err != nil {
		respondError(c, http.StatusInternalServerError, "cannot parse events")
		return
	}

	var upcoming []model.Event
	for _, event := range events {
		layout := "2006-01-02 15:04"
		eventTime, err := time.Parse(layout, event.Date)
		if err != nil {
			upcoming = append(upcoming, event)
			continue
		}
		eventDate := eventTime.Truncate(24 * time.Hour)
		currentDate := time.Now().Truncate(24 * time.Hour)

		if !eventDate.Before(currentDate) {
			upcoming = append(upcoming, event)
		}
	}

	if upcoming == nil {
		upcoming = []model.Event{}
	}

	respondOK(c, upcoming)
}

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

// GalleryYearsHandler retrieves available gallery years
// @Summary Get gallery years
// @Tags gallery
// @Produce json
// @Success 200 {array} int
// @Router /api/gallery/years [get]
func GalleryYearsHandler(c *gin.Context) {
	images, err := loadGallery()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "cannot read gallery")
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

	sort.Sort(sort.Reverse(sort.IntSlice(years)))
	respondOK(c, years)
}

// GalleryImagesHandler retrieves gallery images for a specific year
// @Summary Get gallery images by year
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
			respondError(c, http.StatusBadRequest, "invalid year")
			return
		}
	}

	images, err := loadGallery()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "cannot read gallery")
		return
	}

	result := []model.GalleryImage{}
	for _, img := range images {
		if img.Year == year {
			result = append(result, img)
		}
	}

	respondOK(c, result)
}
