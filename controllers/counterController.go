package controllers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"
	"stok-hadiah/services"

	"github.com/gin-contrib/sessions"
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

	allowedStoreIDs := getSessionStoreIDs(c)
	if !containsInt(allowedStoreIDs, form.StoreID) {
		renderCounterPage(c, counterService, "Store tidak sesuai akses")
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

	allowedStoreIDs := getSessionStoreIDs(c)
	if !containsInt(allowedStoreIDs, form.StoreID) {
		renderCounterPage(c, counterService, "Store tidak sesuai akses")
		return
	}

	exists, err := counterRepo.ExistsInStores(form.ID, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		c.String(http.StatusForbidden, "akses counter ditolak")
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

	allowedStoreIDs := getSessionStoreIDs(c)
	exists, err := counterRepo.ExistsInStores(id, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		c.String(http.StatusForbidden, "akses counter ditolak")
		return
	}

	if err := counterService.DeleteCounter(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counters")
}

func renderCounterPage(c *gin.Context, counterService *services.CounterService, message string) {
	storeIDs := getSessionStoreIDs(c)
	counters, err := counterService.GetCountersByStoreIDs(storeIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	stores, err := storeRepo.GetByIDs(storeIDs)
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

func getSessionStoreIDs(c *gin.Context) []int {
	sess := sessions.Default(c)
	user := sess.Get("user")
	var rawStore string

	switch val := user.(type) {
	case models.SessionUser:
		rawStore = val.StoreID
	case map[string]interface{}:
		rawStore = extractStoreString(val)
	case gin.H:
		rawStore = extractStoreString(map[string]interface{}(val))
	}

	return parseStoreIDs(rawStore)
}

func extractStoreString(data map[string]interface{}) string {
	if v, ok := data["store_id"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	if v, ok := data["StoreID"]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func parseStoreIDs(raw string) []int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return []int{}
	}

	var ids []int
	if strings.HasPrefix(raw, "[") {
		if err := json.Unmarshal([]byte(raw), &ids); err == nil {
			return ids
		}
	}

	parts := strings.Split(raw, ",")
	for _, part := range parts {
		part = strings.TrimSpace(strings.Trim(part, "[]"))
		if part == "" {
			continue
		}
		if id, err := strconv.Atoi(part); err == nil {
			ids = append(ids, id)
		}
	}

	return ids
}

func containsInt(values []int, target int) bool {
	for _, v := range values {
		if v == target {
			return true
		}
	}
	return false
}
