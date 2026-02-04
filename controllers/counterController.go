package controllers

import (
	"net/http"
	"strconv"
	"strings"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"
	"stok-hadiah/services"

	"github.com/gin-gonic/gin"
)

func CounterIndex(c *gin.Context) {
	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counterService := &services.CounterService{Repo: counterRepo}

	renderCounterPage(c, counterService, "")
}

func CounterStore(c *gin.Context) {
	type counterForm struct {
		StoreID      int    `form:"store_id" binding:"required"`
		CounterCode  string `form:"counter_code" binding:"required"`
		CounterName  string `form:"counter_name" binding:"required"`
		TicketPrefix string `form:"ticket_prefix" binding:"required"`
		IsActive     string `form:"is_active"`
	}

	var (
		form          counterForm
		counterRepo   = &repositories.CounterRepository{DB: config.DB}
		counterService = &services.CounterService{Repo: counterRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderCounterPage(c, counterService, "Form tidak lengkap")
		return
	}

	input := models.CounterCreateInput{
		StoreID:      form.StoreID,
		CounterCode:  strings.TrimSpace(form.CounterCode),
		CounterName:  strings.TrimSpace(form.CounterName),
		TicketPrefix: strings.TrimSpace(form.TicketPrefix),
		IsActive:     parseActive(form.IsActive),
	}

	if err := counterService.CreateCounter(input); err != nil {
		renderCounterPage(c, counterService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counters")
}

func CounterUpdate(c *gin.Context) {
	type counterUpdateForm struct {
		ID           int    `form:"counter_id" binding:"required"`
		StoreID      int    `form:"store_id" binding:"required"`
		CounterCode  string `form:"counter_code" binding:"required"`
		CounterName  string `form:"counter_name" binding:"required"`
		TicketPrefix string `form:"ticket_prefix" binding:"required"`
		IsActive     string `form:"is_active"`
	}

	var (
		form          counterUpdateForm
		counterRepo   = &repositories.CounterRepository{DB: config.DB}
		counterService = &services.CounterService{Repo: counterRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderCounterPage(c, counterService, "Form tidak lengkap")
		return
	}

	input := models.CounterUpdateInput{
		ID:           form.ID,
		StoreID:      form.StoreID,
		CounterCode:  strings.TrimSpace(form.CounterCode),
		CounterName:  strings.TrimSpace(form.CounterName),
		TicketPrefix: strings.TrimSpace(form.TicketPrefix),
		IsActive:     parseActive(form.IsActive),
	}

	if err := counterService.UpdateCounter(input); err != nil {
		renderCounterPage(c, counterService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counters")
}

func CounterDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid counter id")
		return
	}

	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counterService := &services.CounterService{Repo: counterRepo}

	if err := counterService.DeleteCounter(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counters")
}

func renderCounterPage(c *gin.Context, counterService *services.CounterService, message string) {
	counters, err := counterService.GetCounters()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	stores, err := storeRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	Render(c, "counter.html", gin.H{
		"Title":    "Setting Counters",
		"Page":     "counters",
		"counters": counters,
		"stores":   stores,
		"Error":    message,
	})
}

func parseActive(val string) bool {
	val = strings.TrimSpace(strings.ToLower(val))
	switch val {
	case "1", "true", "aktif", "active", "yes", "on":
		return true
	case "0", "false", "non_active", "non aktif", "inactive", "off":
		return false
	default:
		return true
	}
}
