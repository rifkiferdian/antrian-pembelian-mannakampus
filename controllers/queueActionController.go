package controllers

import (
	"database/sql"
	"net/http"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/realtime"
	"stok-hadiah/repositories"
	"stok-hadiah/services"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type queueActionForm struct {
	CounterID int `form:"counter_id" json:"counter_id" binding:"required"`
}

type queueActionFn func(service *services.QueueService, storeID, counterID, userID int) (int64, string, error)

func QueueCallNext(c *gin.Context) {
	handleQueueAction(c, "call", "Nomor berikutnya dipanggil", func(service *services.QueueService, storeID, counterID, userID int) (int64, string, error) {
		return service.CallNext(storeID, counterID, userID)
	})
}

func QueueRecall(c *gin.Context) {
	handleQueueAction(c, "recall", "Panggil ulang berhasil", func(service *services.QueueService, storeID, counterID, userID int) (int64, string, error) {
		return service.Recall(storeID, counterID, userID)
	})
}

func QueueDone(c *gin.Context) {
	handleQueueAction(c, "done", "Nomor selesai dilayani", func(service *services.QueueService, storeID, counterID, userID int) (int64, string, error) {
		return service.Finish(storeID, counterID, userID)
	})
}

func QueueSkip(c *gin.Context) {
	handleQueueAction(c, "skip", "Nomor dilewati", func(service *services.QueueService, storeID, counterID, userID int) (int64, string, error) {
		return service.Skip(storeID, counterID, userID)
	})
}

func handleQueueAction(c *gin.Context, action, successMsg string, fn queueActionFn) {
	var form queueActionForm
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

	service := &services.QueueService{Repo: &repositories.QueueActionRepository{DB: config.DB}}
	_, ticketNo, err := fn(service, counter.StoreID, form.CounterID, userID)
	if err != nil {
		switch err {
		case services.ErrNoWaitingTicket:
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "Tidak ada antrian menunggu"})
			return
		case services.ErrNoCalledTicket:
			c.JSON(http.StatusBadRequest, gin.H{"ok": false, "message": "Tidak ada antrian yang sedang dipanggil"})
			return
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
	}

	payload, payloadErr := buildQueueViewPayload(counter.StoreID, action)
	if payloadErr == nil {
		realtime.QueueHub.Broadcast(counter.StoreID, payload)
	}

	resp := gin.H{
		"ok":        true,
		"message":   successMsg,
		"ticket_no": ticketNo,
	}
	if payloadErr == nil {
		resp["data"] = payload
	} else {
		resp["warning"] = "Berhasil, tetapi gagal memuat status antrian terbaru"
	}

	c.JSON(http.StatusOK, resp)
}

func getSessionUserID(c *gin.Context) int {
	sess := sessions.Default(c)
	if v := sess.Get("user_id"); v != nil {
		switch id := v.(type) {
		case int:
			return id
		case int64:
			return int(id)
		case float64:
			return int(id)
		}
	}

	if u := sess.Get("user"); u != nil {
		switch val := u.(type) {
		case models.SessionUser:
			return val.UserID
		case map[string]interface{}:
			if id, ok := val["user_id"]; ok {
				return normalizeSessionID(id)
			}
			if id, ok := val["UserID"]; ok {
				return normalizeSessionID(id)
			}
		case gin.H:
			if id, ok := val["user_id"]; ok {
				return normalizeSessionID(id)
			}
			if id, ok := val["UserID"]; ok {
				return normalizeSessionID(id)
			}
		}
	}

	return 0
}

func normalizeSessionID(val interface{}) int {
	switch id := val.(type) {
	case int:
		return id
	case int64:
		return int(id)
	case float64:
		return int(id)
	default:
		return 0
	}
}
