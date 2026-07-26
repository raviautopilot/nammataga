package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"taga-api/model"

	"github.com/gin-gonic/gin"
)

const (
	eventsFile = "data/events.json"
)



// UpcomingEventsHandler godoc
// @Summary Get upcoming events
// @Description Returns upcoming events
// @Tags Events
// @Produce json
// @Success 200 {array} model.Event
// @Router /api/events/upcoming [get]
func UpcomingEventsHandler(c *gin.Context) {
	file, err := os.ReadFile(eventsFile)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cannot read events",
		})
		return
	}

	var events []model.Event
	if err := json.Unmarshal(file, &events); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "cannot parse events",
		})
		return
	}

	var upcoming []model.Event
	for _, event := range events {
		layout := "2006-01-02 15:04"
		eventTime, err := time.Parse(layout, event.Date)
		if err != nil {
			// If date doesn't parse, include it anyway to avoid silently dropping events
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

	c.JSON(http.StatusOK, upcoming)
}


