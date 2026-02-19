package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/realtime"
	"stok-hadiah/repositories"

	"github.com/gin-gonic/gin"
)

func GuestQueuePage(c *gin.Context) {
	storeRepo := &repositories.StoreRepository{DB: config.DB}

	storeID, _ := strconv.Atoi(strings.TrimSpace(c.Param("store_id")))
	var (
		store models.Store
		err   error
	)

	if storeID > 0 {
		store, err = storeRepo.GetByID(storeID)
		if err != nil {
			if err == sql.ErrNoRows {
				c.String(http.StatusNotFound, "store tidak ditemukan")
				return
			}
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		store, err = storeRepo.GetFirstActive()
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}
		storeID = store.StoreID
	}

	repo := &repositories.GuestQueueRepository{DB: config.DB}
	today := time.Now()
	counters, err := repo.GetCountersForGuest(storeID, today)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	for i := range counters {
		counters[i].IndexLabel = formatCounterLabel(counters[i].CounterCode, i)
		if len(counters[i].StaffNames) == 0 {
			counters[i].StaffNames = []string{"-"}
		}
		counters[i].Icon = resolveCategoryIcon(counters[i].CategoryName)
	}

	c.HTML(http.StatusOK, "guest.html", gin.H{
		"Title":    "Cetak Antrian",
		"Store":    store,
		"StoreID":  store.StoreID,
		"Counters": counters,
		"Date":     today.Format("02 Jan 2006"),
	})
}

func GuestQueuePrint(c *gin.Context) {
	counterID, err := strconv.Atoi(strings.TrimSpace(c.PostForm("counter_id")))
	if err != nil || counterID <= 0 {
		c.String(http.StatusBadRequest, "counter tidak valid")
		return
	}

	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counter, err := counterRepo.GetByID(counterID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "counter tidak ditemukan")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !counter.IsActive {
		c.String(http.StatusBadRequest, "counter tidak aktif")
		return
	}

	ticketRepo := &repositories.QueueTicketRepository{DB: config.DB}
	today := time.Now()
	ticketPrefix := counter.TicketPrefix
	if strings.TrimSpace(ticketPrefix) == "" {
		ticketPrefix = "A"
	}
	ticket, err := ticketRepo.CreateTicket(counter.StoreID, counterID, ticketPrefix, today)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	if payload, payloadErr := buildQueueViewPayload(counter.StoreID, "new_ticket"); payloadErr == nil {
		realtime.QueueHub.Broadcast(counter.StoreID, payload)
	}

	c.Redirect(http.StatusSeeOther, "/guest/ticket/"+strconv.FormatInt(ticket.ID, 10))
}

func GuestQueueState(c *gin.Context) {
	storeRepo := &repositories.StoreRepository{DB: config.DB}

	storeID, _ := strconv.Atoi(strings.TrimSpace(c.Param("store_id")))
	if storeID <= 0 {
		store, err := storeRepo.GetFirstActive()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
		storeID = store.StoreID
	} else {
		if _, err := storeRepo.GetByID(storeID); err != nil {
			if err == sql.ErrNoRows {
				c.JSON(http.StatusNotFound, gin.H{"ok": false, "message": "store tidak ditemukan"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
			return
		}
	}

	repo := &repositories.GuestQueueRepository{DB: config.DB}
	today := time.Now()
	counters, err := repo.GetCountersForGuest(storeID, today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false, "message": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"ok":       true,
		"store_id": storeID,
		"counters": counters,
		"date":     today.Format("2006-01-02"),
		"time":     today.Format("15:04:05"),
	})
}

func GuestTicketShow(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "ticket tidak valid")
		return
	}

	ticketRepo := &repositories.QueueTicketRepository{DB: config.DB}
	ticket, err := ticketRepo.GetTicketByID(id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.String(http.StatusNotFound, "ticket tidak ditemukan")
			return
		}
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.HTML(http.StatusOK, "guest_ticket.html", gin.H{
		"Title":  "Cetak Tiket",
		"Ticket": ticket,
	})
}

func resolveCategoryIcon(categoryName string) string {
	name := strings.ToLower(categoryName)
	switch {
	case strings.Contains(name, "food"):
		return "food"
	case strings.Contains(name, "toilet"):
		return "toiletries"
	case strings.Contains(name, "fashion"):
		return "fashion"
	default:
		return "default"
	}
}
