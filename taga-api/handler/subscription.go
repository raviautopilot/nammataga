package handler

import (
	"net/http"
	"taga-api/config"
	"taga-api/utils"

	"github.com/gin-gonic/gin"
)

type SubscriptionPlanResponse struct {
	ID                string  `json:"id"`
	Name              string  `json:"name"`
	Description       string  `json:"description"`
	Amount            float64 `json:"amount,omitempty"`
	Frequency         string  `json:"frequency"`
	Status            string  `json:"status"`
	LastPaidDate      string  `json:"lastPaidDate,omitempty"`
	NextDueDate       string  `json:"nextDueDate,omitempty"`
	AllowCustomAmount bool    `json:"allowCustomAmount,omitempty"`
	OneTime           bool    `json:"oneTime"`
	NeedBased         bool    `json:"needBased"`
}

// GetSubscriptions godoc
// @Summary      Get subscriptions
// @Description  Fetch all subscription details for members
// @Tags         Subscription
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {array}  SubscriptionPlanResponse
// @Failure      500  {object}  map[string]string
// @Router       /api/subscriptions [get]
func GetSubscriptions(c *gin.Context) {
	var subscriptions []map[string]interface{}

	err := utils.ReadJSON(config.Config.Data.Config.SubscriptionType, &subscriptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read subscriptions",
		})
		return
	}

	response := make([]SubscriptionPlanResponse, len(subscriptions))
	for i, sub := range subscriptions {
		response[i] = SubscriptionPlanResponse{
			ID:                utils.GetStringFromMap(sub, "id"),
			Name:              utils.GetStringFromMap(sub, "name"),
			Description:       utils.GetStringFromMap(sub, "description"),
			Amount:            getFloatFromMap(sub, "amount"),
			Frequency:         utils.GetStringFromMap(sub, "frequency"),
			Status:            utils.GetStringFromMap(sub, "status"),
			LastPaidDate:      utils.GetStringFromMap(sub, "lastPaidDate"),
			NextDueDate:       utils.GetStringFromMap(sub, "nextDueDate"),
			AllowCustomAmount: getBoolFromMap(sub, "allowCustomAmount"),
			OneTime:           getBoolFromMap(sub, "oneTime"),
			NeedBased:         getBoolFromMap(sub, "needBased"),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
	})
}

// getFloatFromMap safely extracts a float64 value from a map
func getFloatFromMap(m map[string]interface{}, key string) float64 {
	if val, ok := m[key]; ok && val != nil {
		switch v := val.(type) {
		case float64:
			return v
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

// getBoolFromMap safely extracts a bool value from a map
func getBoolFromMap(m map[string]interface{}, key string) bool {
	if val, ok := m[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return b
		}
	}
	return false
}
