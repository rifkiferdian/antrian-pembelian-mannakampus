package controllers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"

	"github.com/gin-gonic/gin"
)

func ViewQueuePage(c *gin.Context) {
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

	repo := &repositories.ViewQueueRepository{DB: config.DB}
	today := time.Now()

	currentCall, err := repo.GetCurrentCall(storeID, today)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	counters, err := repo.GetCountersStatus(storeID, today)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	currentTicket := strings.TrimSpace(currentCall.TicketNo)
	for i := range counters {
		counters[i].CounterLabel = formatCounterLabel(counters[i].CounterCode, i)
		if counters[i].TicketNo == "" {
			counters[i].TicketNo = "-"
		}
		isCurrent := currentTicket != "" && counters[i].TicketNo == currentTicket
		counters[i].IsCurrent = isCurrent
		counters[i].StatusLabel, counters[i].StatusClass = resolveQueueStatus(counters[i].TicketStatus, isCurrent)
	}

	if currentTicket == "" {
		for _, counter := range counters {
			if counter.TicketStatus == "CALLED" && counter.TicketNo != "-" {
				currentCall.TicketNo = counter.TicketNo
				currentCall.CounterLabel = counter.CounterLabel
				currentCall.CategoryName = counter.CategoryName
				currentTicket = counter.TicketNo
				break
			}
		}
	} else {
		if strings.TrimSpace(currentCall.CounterLabel) == "" {
			for _, counter := range counters {
				if counter.TicketNo == currentTicket {
					currentCall.CounterLabel = counter.CounterLabel
					currentCall.CategoryName = counter.CategoryName
					break
				}
			}
		} else {
			currentCall.CounterLabel = formatCounterLabel(currentCall.CounterLabel, -1)
		}
	}

	c.HTML(http.StatusOK, "view_queue.html", gin.H{
		"Title":       "Tampilan Antrian",
		"Store":       store,
		"StoreID":     store.StoreID,
		"Date":        today.Format("02 Jan 2006"),
		"CurrentCall": currentCall,
		"Counters":    counters,
	})
}

func resolveQueueStatus(status string, isCurrent bool) (string, string) {
	switch status {
	case "CALLED":
		if isCurrent {
			return "SEDANG MEMANGGIL", "bg-orange-100 text-orange-600"
		}
		return "SEDANG DILAYANI", "bg-orange-100 text-orange-600"
	case "WAITING":
		return "MENUNGGU", "bg-slate-100 text-slate-500"
	case "DONE":
		return "SELESAI", "bg-emerald-100 text-emerald-600"
	case "SKIPPED":
		return "TERLEWAT", "bg-rose-100 text-rose-600"
	case "CANCELLED":
		return "BATAL", "bg-slate-200 text-slate-600"
	default:
		return "BELUM ADA", "bg-slate-100 text-slate-400"
	}
}

func formatCounterLabel(counterCode string, index int) string {
	code := normalizeCounterCode(counterCode)
	if code == "" {
		if index < 0 {
			return "Loket -"
		}
		code = strconv.Itoa(index + 1)
	}
	if number, err := strconv.Atoi(code); err == nil {
		code = fmt.Sprintf("%02d", number)
	}
	return "Loket " + code
}

func normalizeCounterCode(counterCode string) string {
	code := strings.TrimSpace(counterCode)
	if code == "" {
		return ""
	}
	lower := strings.ToLower(code)
	if strings.HasPrefix(lower, "loket") {
		code = strings.TrimSpace(code[5:])
	}
	if code == "" {
		return ""
	}
	if number, err := strconv.Atoi(code); err == nil {
		return strconv.Itoa(number)
	}
	return code
}
