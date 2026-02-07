package controllers

import (
	"database/sql"
	"net/http"
	"strconv"

	"stok-hadiah/config"
	"stok-hadiah/repositories"

	"github.com/gin-gonic/gin"
)

func DashboardQueueState(c *gin.Context) {
	counterID, _ := strconv.Atoi(c.Query("counter_id"))
	if counterID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "counter_id tidak valid"})
		return
	}

	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counter, err := counterRepo.GetByID(counterID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "message": "Loket tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	allowedStoreIDs := getSessionStoreIDs(c)
	if !containsInt(allowedStoreIDs, counter.StoreID) {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "message": "Akses loket ditolak"})
		return
	}

	var staffStatus string
	userID := getSessionUserID(c)
	if userID > 0 {
		counterStaffRepo := &repositories.CounterStaffRepository{DB: config.DB}
		status, _, statusErr := counterStaffRepo.GetStatusByCounterAndUser(counter.ID, userID)
		if statusErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": statusErr.Error()})
			return
		}
		staffStatus = status
	}

	serving, waitingItems, waitingTotal, err := buildDashboardQueueState(counter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":            true,
		"counter_id":    counter.ID,
		"staff_status":  staffStatus,
		"serving":       serving,
		"waiting_items": waitingItems,
		"waiting_total": waitingTotal,
	})
}
