package controllers

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"
	"stok-hadiah/services"

	"github.com/gin-gonic/gin"
)

type dashboardStatusForm struct {
	CounterID            int    `form:"counter_id" json:"counter_id" binding:"required"`
	Status               string `form:"status" json:"status" binding:"required"`
	InactiveStartedAtRaw string `form:"inactive_started_at" json:"inactive_started_at"`
	InactiveUntilRaw     string `form:"inactive_until" json:"inactive_until"`
	InactiveAnnouncement string `form:"inactive_announcement" json:"inactive_announcement"`
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

	inactiveStartedAt, err := parseDashboardStatusTime(form.InactiveStartedAtRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "Format inactive_started_at tidak valid"})
		return
	}
	inactiveUntil, err := parseDashboardStatusTime(form.InactiveUntilRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "Format inactive_until tidak valid"})
		return
	}

	counterStaffService := &services.CounterStaffService{Repo: &repositories.CounterStaffRepository{DB: config.DB}}
	detail, err := counterStaffService.UpdateStatusByCounterAndUser(models.CounterStaffStatusUpdateInput{
		CounterID:            form.CounterID,
		UserID:               userID,
		Status:               form.Status,
		InactiveStartedAt:    inactiveStartedAt,
		InactiveUntil:        inactiveUntil,
		InactiveAnnouncement: strings.TrimSpace(form.InactiveAnnouncement),
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": err.Error()})
		return
	}

	message := "Status loket diperbarui"
	if detail.Status == "ACTIVE" {
		message = "Status loket aktif"
	} else if detail.Status == "REST" {
		message = "Status loket istirahat"
	} else if detail.Status == "INACTIVE" {
		message = "Status loket non-aktif"
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":                    true,
		"status":                detail.Status,
		"inactive_started_at":   detail.InactiveStartedAt,
		"inactive_until":        detail.InactiveUntil,
		"inactive_announcement": detail.InactiveAnnouncement,
		"message":               message,
	})
}

func parseDashboardStatusTime(raw string) (*time.Time, error) {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil, nil
	}

	layouts := []string{
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
	}
	for _, layout := range layouts {
		if parsed, err := time.ParseInLocation(layout, trimmed, time.Local); err == nil {
			return &parsed, nil
		}
	}

	return nil, errors.New("format waktu tidak valid")
}
