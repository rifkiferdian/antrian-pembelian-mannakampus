package controllers

import (
	"net/http"
	"strconv"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"
	"stok-hadiah/services"

	"github.com/gin-gonic/gin"
)

func CounterStaffIndex(c *gin.Context) {
	counterStaffRepo := &repositories.CounterStaffRepository{DB: config.DB}
	counterStaffService := &services.CounterStaffService{Repo: counterStaffRepo}

	renderCounterStaffPage(c, counterStaffService, "")
}

func CounterStaffStore(c *gin.Context) {
	type counterStaffForm struct {
		CounterID int    `form:"counter_id" binding:"required"`
		UserID    int    `form:"user_id" binding:"required"`
		Status    string `form:"status"`
	}

	var (
		form                counterStaffForm
		counterRepo         = &repositories.CounterRepository{DB: config.DB}
		counterStaffRepo    = &repositories.CounterStaffRepository{DB: config.DB}
		counterStaffService = &services.CounterStaffService{Repo: counterStaffRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderCounterStaffPage(c, counterStaffService, "Form tidak lengkap")
		return
	}

	allowedStoreIDs := getSessionStoreIDs(c)
	exists, err := counterRepo.ExistsInStores(form.CounterID, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		renderCounterStaffPage(c, counterStaffService, "Counter tidak sesuai akses")
		return
	}

	input := models.CounterStaffCreateInput{
		CounterID: form.CounterID,
		UserID:    form.UserID,
		Status:    form.Status,
	}

	if err := counterStaffService.CreateCounterStaff(input); err != nil {
		renderCounterStaffPage(c, counterStaffService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counter_staff")
}

func CounterStaffUpdate(c *gin.Context) {
	type counterStaffUpdateForm struct {
		ID        int    `form:"counter_staff_id" binding:"required"`
		CounterID int    `form:"counter_id" binding:"required"`
		UserID    int    `form:"user_id" binding:"required"`
		Status    string `form:"status"`
	}

	var (
		form                counterStaffUpdateForm
		counterRepo         = &repositories.CounterRepository{DB: config.DB}
		counterStaffRepo    = &repositories.CounterStaffRepository{DB: config.DB}
		counterStaffService = &services.CounterStaffService{Repo: counterStaffRepo}
	)

	if err := c.ShouldBind(&form); err != nil {
		renderCounterStaffPage(c, counterStaffService, "Form tidak lengkap")
		return
	}

	allowedStoreIDs := getSessionStoreIDs(c)

	exists, err := counterStaffRepo.ExistsInStores(form.ID, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		c.String(http.StatusForbidden, "akses counter staff ditolak")
		return
	}

	counterExists, err := counterRepo.ExistsInStores(form.CounterID, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !counterExists {
		renderCounterStaffPage(c, counterStaffService, "Counter tidak sesuai akses")
		return
	}

	input := models.CounterStaffUpdateInput{
		ID:        form.ID,
		CounterID: form.CounterID,
		UserID:    form.UserID,
		Status:    form.Status,
	}

	if err := counterStaffService.UpdateCounterStaff(input); err != nil {
		renderCounterStaffPage(c, counterStaffService, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counter_staff")
}

func CounterStaffDelete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil || id <= 0 {
		c.String(http.StatusBadRequest, "invalid counter staff id")
		return
	}

	counterStaffRepo := &repositories.CounterStaffRepository{DB: config.DB}
	counterStaffService := &services.CounterStaffService{Repo: counterStaffRepo}

	allowedStoreIDs := getSessionStoreIDs(c)
	exists, err := counterStaffRepo.ExistsInStores(id, allowedStoreIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	if !exists {
		c.String(http.StatusForbidden, "akses counter staff ditolak")
		return
	}

	if err := counterStaffService.DeleteCounterStaff(id); err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	c.Redirect(http.StatusSeeOther, "/counter_staff")
}

func renderCounterStaffPage(c *gin.Context, counterStaffService *services.CounterStaffService, message string) {
	storeIDs := getSessionStoreIDs(c)
	counterStaffs, err := counterStaffService.GetByStoreIDs(storeIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counters, err := counterRepo.GetByStoreIDs(storeIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	userRepo := &repositories.UserRepository{DB: config.DB}
	users, err := userRepo.GetAll()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	filteredUsers := filterUsersByStoreIDs(users, storeIDs)

	Render(c, "counter_staff.html", gin.H{
		"Title":         "Setting Counters Staff",
		"Page":          "counter_staff",
		"counterStaffs": counterStaffs,
		"counters":      counters,
		"users":         filteredUsers,
		"Error":         message,
	})
}

func filterUsersByStoreIDs(users []models.User, storeIDs []int) []models.User {
	if len(storeIDs) == 0 {
		return []models.User{}
	}

	filtered := make([]models.User, 0, len(users))
	for _, user := range users {
		if len(user.StoreIDs) == 0 {
			continue
		}
		for _, storeID := range user.StoreIDs {
			if containsInt(storeIDs, storeID) {
				filtered = append(filtered, user)
				break
			}
		}
	}

	return filtered
}
