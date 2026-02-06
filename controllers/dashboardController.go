package controllers

import (
	"net/http"
	"strconv"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"

	"github.com/gin-gonic/gin"
)

func DashboardIndex(c *gin.Context) {
	storeIDs := getSessionStoreIDs(c)
	userID := getSessionUserID(c)
	counterRepo := &repositories.CounterRepository{DB: config.DB}
	var counters []models.Counter
	var err error
	if userID > 0 {
		counters, err = counterRepo.GetByStoreIDsAndUserID(storeIDs, userID)
	} else {
		counters = []models.Counter{}
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	selectedCounter := models.Counter{}
	if counterIDStr := c.Query("counter_id"); counterIDStr != "" {
		if counterID, parseErr := strconv.Atoi(counterIDStr); parseErr == nil && counterID > 0 {
			for _, counter := range counters {
				if counter.ID == counterID {
					selectedCounter = counter
					break
				}
			}
		}
	}

	if selectedCounter.ID == 0 {
		for _, counter := range counters {
			if counter.IsActive {
				selectedCounter = counter
				break
			}
		}
	}

	var (
		serving       models.DashboardServing
		waitingItems  []models.DashboardQueueItem
		waitingTotal  int
		stateLoadErr  error
		selectedIDVal int
	)

	if selectedCounter.ID > 0 {
		serving, waitingItems, waitingTotal, stateLoadErr = buildDashboardQueueState(selectedCounter)
		selectedIDVal = selectedCounter.ID
		if stateLoadErr != nil {
			c.String(http.StatusInternalServerError, stateLoadErr.Error())
			return
		}
	}

	Render(c, "dashboard.html", gin.H{
		"Title":             "Dashboard",
		"Page":              "dashboard",
		"Counters":          counters,
		"SelectedCounterID": selectedIDVal,
		"Serving":           serving,
		"WaitingItems":      waitingItems,
		"WaitingTotal":      waitingTotal,
	})

}
