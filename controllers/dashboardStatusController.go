package controllers

import (
	"database/sql"
	"net/http"

	"stok-hadiah/config"
	"stok-hadiah/repositories"
	"stok-hadiah/services"

	"github.com/gin-gonic/gin"
)

type dashboardStatusForm struct {
	CounterID int    `form:"counter_id" json:"counter_id" binding:"required"`
	Status    string `form:"status" json:"status" binding:"required"`
}

func DashboardCounterStatus(c *gin.Context) {
	var form dashboardStatusForm
	if err := c.ShouldBind(&form); err != nil || form.CounterID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "counter_id tidak valid"})
		return
	}

	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counter, err := counterRepo.GetByID(form.CounterID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"ok": false, "message": "Loket tidak ditemukan"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	if !counter.IsActive {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "Loket tidak aktif"})
		return
	}

	allowedStoreIDs := getSessionStoreIDs(c)
	if !containsInt(allowedStoreIDs, counter.StoreID) {
		c.JSON(http.StatusForbidden, gin.H{"ok": false, "message": "Akses loket ditolak"})
		return
	}

	userID := getSessionUserID(c)
	if userID == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"ok": false, "message": "Sesi tidak valid"})
		return
	}

	counterStaffService := &services.CounterStaffService{Repo: &repositories.CounterStaffRepository{DB: config.DB}}
	status, err := counterStaffService.UpdateStatusByCounterAndUser(form.CounterID, userID, form.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
		return
	}

	message := "Status loket diperbarui"
	if status == "ACTIVE" {
		message = "Status loket aktif"
	} else if status == "REST" {
		message = "Status loket istirahat"
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":      true,
		"status":  status,
		"message": message,
	})
}
