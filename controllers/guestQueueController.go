package controllers

import (
	"database/sql"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"

	"github.com/gin-gonic/gin"
)

func GuestQueuePage(c *gin.Context) {
	storeRepo := &repositories.StoreRepository{DB: config.DB}

	storeID, _ := strconv.Atoi(strings.TrimSpace(c.Query("store_id")))
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
		counters[i].IndexLabel = "Loket " + strconv.Itoa(i+1)
		counters[i].StaffName = "-"
		counters[i].Icon = resolveCategoryIcon(counters[i].CategoryName)
	}

	c.HTML(http.StatusOK, "guest.html", gin.H{
		"Title":    "Cetak Antrian",
		"Store":    store,
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

	c.Redirect(http.StatusSeeOther, "/guest/ticket/"+strconv.FormatInt(ticket.ID, 10))
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
