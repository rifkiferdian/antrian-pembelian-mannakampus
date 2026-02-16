package controllers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"stok-hadiah/config"
	"stok-hadiah/models"
	"stok-hadiah/repositories"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const reportDateLayout = "2006-01-02"

func ReportsIndex(c *gin.Context) {
	storeIDs := getSessionStoreIDs(c)
	role := getSessionRole(c)

	counterRepo := &repositories.CounterRepository{DB: config.DB}
	counters := make([]models.Counter, 0)
	var err error

	if hasGlobalReportAccess(role) {
		counters, err = counterRepo.GetByStoreIDs(storeIDs)
	} else {
		userID := getSessionUserID(c)
		counters, err = counterRepo.GetByStoreIDsAndUserID(storeIDs, userID)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	allowedCounterIDs := make([]int, 0, len(counters))
	for _, counter := range counters {
		if counter.ID > 0 {
			allowedCounterIDs = append(allowedCounterIDs, counter.ID)
		}
	}

	storeRepo := &repositories.StoreRepository{DB: config.DB}
	stores, err := storeRepo.GetByIDs(storeIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	selectedStoreID := parseQueryInt(c.Query("store_id"))
	if selectedStoreID > 0 && !containsInt(storeIDs, selectedStoreID) {
		selectedStoreID = 0
	}

	selectedCounterID := parseQueryInt(c.Query("counter_id"))
	if selectedCounterID > 0 && !counterAllowed(counters, selectedCounterID) {
		selectedCounterID = 0
	}

	filteredCounters := filterCountersByStore(counters, selectedStoreID)
	filteredCounterIDs := make([]int, 0, len(filteredCounters))
	for _, counter := range filteredCounters {
		filteredCounterIDs = append(filteredCounterIDs, counter.ID)
	}
	if selectedCounterID > 0 && !containsInt(filteredCounterIDs, selectedCounterID) {
		selectedCounterID = 0
	}

	reportRepo := &repositories.ReportRepository{DB: config.DB}
	categories, err := reportRepo.GetCategoriesByCounterIDs(filteredCounterIDs)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	selectedCategoryID := parseQueryInt(c.Query("category_id"))
	if selectedCategoryID > 0 && !categoryAllowed(categories, selectedCategoryID) {
		selectedCategoryID = 0
	}

	now := time.Now()
	defaultStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	defaultEnd := defaultStart.AddDate(0, 1, -1)

	startDate := parseQueryDate(c.Query("start_date"), defaultStart)
	endDate := parseQueryDate(c.Query("end_date"), defaultEnd)
	if endDate.Before(startDate) {
		endDate = startDate
	}

	page := parseQueryInt(c.Query("page"))
	if page <= 0 {
		page = 1
	}
	perPage := 10

	summary := models.ReportSummary{}
	logItems := make([]models.ReportQueueItem, 0)
	pagination := models.ReportPagination{Page: page, PerPage: perPage}
	hourlyCounts := make([]int, 24)

	if len(storeIDs) > 0 && len(allowedCounterIDs) > 0 {
		params := repositories.ReportQuery{
			StoreIDs:   storeIDs,
			CounterIDs: allowedCounterIDs,
			StoreID:    selectedStoreID,
			CounterID:  selectedCounterID,
			CategoryID: selectedCategoryID,
			StartDate:  startDate,
			EndDate:    endDate,
		}

		summary, err = reportRepo.GetSummary(params)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if peakHour, ok, err := reportRepo.GetPeakHour(params); err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		} else if ok {
			summary.PeakHourLabel = fmt.Sprintf("%02d:00", peakHour)
		} else {
			summary.PeakHourLabel = "-"
		}

		offset := (page - 1) * perPage
		rows, total, err := reportRepo.GetQueueLogs(params, perPage, offset)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			return
		}

		if total > 0 && offset >= total && page > 1 {
			lastPage := total / perPage
			if total%perPage != 0 {
				lastPage++
			}
			page = lastPage
			offset = (page - 1) * perPage
			rows, total, err = reportRepo.GetQueueLogs(params, perPage, offset)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				return
			}
		}

		logItems = buildReportLogItems(rows)
		pagination = buildReportPagination(page, perPage, total, selectedStoreID, selectedCounterID, selectedCategoryID, startDate, endDate)

		if counts, err := reportRepo.GetHourlyCounts(params); err == nil && len(counts) == 24 {
			hourlyCounts = counts
		}
	} else {
		summary.PeakHourLabel = "-"
	}

	summary.AvgWaitingLabel = formatDurationLabel(summary.AvgWaitingSeconds)
	summary.AvgServiceLabel = formatDurationLabel(summary.AvgServiceSeconds)
	if summary.TotalTickets > 0 {
		summary.SuccessRateLabel = fmt.Sprintf("%.0f%% Success", summary.SuccessRate)
	} else {
		summary.SuccessRateLabel = "0% Success"
	}

	labels := make([]string, 24)
	for i := 0; i < 24; i++ {
		labels[i] = fmt.Sprintf("%02d:00", i)
	}
	labelsJSON, _ := json.Marshal(labels)
	seriesJSON, _ := json.Marshal(hourlyCounts)

	Render(c, "reports.html", gin.H{
		"Title":              "Reports",
		"Page":               "reports",
		"Summary":            summary,
		"Logs":               logItems,
		"Pagination":         pagination,
		"Stores":             stores,
		"Counters":           filteredCounters,
		"Categories":         categories,
		"SelectedStoreID":    selectedStoreID,
		"SelectedCounterID":  selectedCounterID,
		"SelectedCategoryID": selectedCategoryID,
		"StartDate":          startDate.Format(reportDateLayout),
		"EndDate":            endDate.Format(reportDateLayout),
		"BusyLabels":         template.JS(labelsJSON),
		"BusySeries":         template.JS(seriesJSON),
	})
}

func parseQueryInt(value string) int {
	if value == "" {
		return 0
	}
	id, err := strconv.Atoi(value)
	if err != nil || id < 0 {
		return 0
	}
	return id
}

func parseQueryDate(value string, fallback time.Time) time.Time {
	if value == "" {
		return fallback
	}
	parsed, err := time.Parse(reportDateLayout, value)
	if err != nil {
		return fallback
	}
	return parsed
}

func counterAllowed(counters []models.Counter, id int) bool {
	for _, counter := range counters {
		if counter.ID == id {
			return true
		}
	}
	return false
}

func categoryAllowed(categories []models.ServiceCategory, id int) bool {
	for _, cat := range categories {
		if cat.ID == id {
			return true
		}
	}
	return false
}

func filterCountersByStore(counters []models.Counter, storeID int) []models.Counter {
	if storeID <= 0 {
		return counters
	}
	filtered := make([]models.Counter, 0)
	for _, counter := range counters {
		if counter.StoreID == storeID {
			filtered = append(filtered, counter)
		}
	}
	return filtered
}

func buildReportLogItems(rows []repositories.ReportQueueRow) []models.ReportQueueItem {
	items := make([]models.ReportQueueItem, 0, len(rows))
	for _, row := range rows {
		item := models.ReportQueueItem{
			TicketNo:      row.TicketNo,
			CategoryName:  row.CategoryName,
			CounterName:   row.CounterName,
			StoreName:     row.StoreName,
			Status:        row.Status,
			ArrivalLabel:  row.CreatedAt.Format("15:04"),
			CallLabel:     "-",
			DoneLabel:     "-",
			DurationLabel: "-",
		}

		if row.CalledAt != nil {
			item.CallLabel = row.CalledAt.Format("15:04")
		}
		if row.DoneAt != nil {
			item.DoneLabel = row.DoneAt.Format("15:04")
		}
		if row.DurationSeconds != nil {
			item.DurationLabel = formatDurationLabel(*row.DurationSeconds)
		}

		items = append(items, item)
	}
	return items
}

func buildReportPagination(page, perPage, total, storeID, counterID, categoryID int, startDate, endDate time.Time) models.ReportPagination {
	pagination := models.ReportPagination{
		Page:    page,
		PerPage: perPage,
		Total:   total,
	}

	if total <= 0 {
		return pagination
	}

	totalPages := total / perPage
	if total%perPage != 0 {
		totalPages++
	}
	if page > totalPages {
		page = totalPages
	}
	if page < 1 {
		page = 1
	}

	pagination.Page = page
	pagination.TotalPages = totalPages

	from := (page-1)*perPage + 1
	to := from + perPage - 1
	if to > total {
		to = total
	}

	pagination.From = from
	pagination.To = to

	values := url.Values{}
	if storeID > 0 {
		values.Set("store_id", strconv.Itoa(storeID))
	}
	if counterID > 0 {
		values.Set("counter_id", strconv.Itoa(counterID))
	}
	if categoryID > 0 {
		values.Set("category_id", strconv.Itoa(categoryID))
	}
	values.Set("start_date", startDate.Format(reportDateLayout))
	values.Set("end_date", endDate.Format(reportDateLayout))

	pagination.HasPrev = page > 1
	if pagination.HasPrev {
		values.Set("page", strconv.Itoa(page-1))
		pagination.PrevURL = "/reports?" + values.Encode()
	}

	pagination.HasNext = page < totalPages
	if pagination.HasNext {
		values.Set("page", strconv.Itoa(page+1))
		pagination.NextURL = "/reports?" + values.Encode()
	}

	return pagination
}

func formatDurationLabel(seconds int) string {
	if seconds <= 0 {
		return "0m"
	}

	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	secs := seconds % 60

	if hours > 0 {
		if minutes > 0 {
			return fmt.Sprintf("%dh %dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}

	if minutes > 0 {
		if secs > 0 {
			return fmt.Sprintf("%dm %ds", minutes, secs)
		}
		return fmt.Sprintf("%dm", minutes)
	}

	return fmt.Sprintf("%ds", secs)
}

func hasGlobalReportAccess(role string) bool {
	role = strings.TrimSpace(strings.ToLower(role))
	if role == "" {
		return false
	}

	parts := strings.FieldsFunc(role, func(r rune) bool {
		return r == ',' || r == ';' || r == '|'
	})
	if len(parts) == 0 {
		parts = []string{role}
	}

	for _, part := range parts {
		normalized := strings.TrimSpace(part)
		normalized = strings.ReplaceAll(normalized, "_", "-")
		normalized = strings.ReplaceAll(normalized, " ", "-")
		switch normalized {
		case "super-admin", "admin", "manager":
			return true
		}
	}

	return false
}

func getSessionRole(c *gin.Context) string {
	sess := sessions.Default(c)
	user := sess.Get("user")

	switch val := user.(type) {
	case models.SessionUser:
		return strings.TrimSpace(val.Role)
	case map[string]interface{}:
		if role, ok := val["role"].(string); ok {
			return strings.TrimSpace(role)
		}
		if role, ok := val["Role"].(string); ok {
			return strings.TrimSpace(role)
		}
	case gin.H:
		if role, ok := val["role"].(string); ok {
			return strings.TrimSpace(role)
		}
		if role, ok := val["Role"].(string); ok {
			return strings.TrimSpace(role)
		}
	}

	return ""
}
