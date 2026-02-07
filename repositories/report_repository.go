package repositories

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"stok-hadiah/models"
)

type ReportRepository struct {
	DB *sql.DB
}

// ReportQuery menampung parameter filter laporan.
type ReportQuery struct {
	StoreIDs   []int
	CounterIDs []int
	StoreID    int
	CounterID  int
	CategoryID int
	StartDate  time.Time
	EndDate    time.Time
}

// ReportQueueRow menampung data log raw dari database.
type ReportQueueRow struct {
	TicketNo        string
	CategoryName    string
	CounterName     string
	StoreName       string
	CreatedAt       time.Time
	CalledAt        *time.Time
	DoneAt          *time.Time
	DurationSeconds *int
	Status          string
}

// GetCategoriesByCounterIDs mengambil kategori berdasarkan daftar counter.
func (r *ReportRepository) GetCategoriesByCounterIDs(counterIDs []int) ([]models.ServiceCategory, error) {
	if len(counterIDs) == 0 {
		return []models.ServiceCategory{}, nil
	}

	placeholders := make([]string, len(counterIDs))
	args := make([]interface{}, len(counterIDs))
	for i, id := range counterIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT DISTINCT
			sc.id,
			sc.category_name,
			sc.ticket_prefix
		FROM counters c
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE c.id IN (` + strings.Join(placeholders, ",") + `)
		  AND sc.id IS NOT NULL
		ORDER BY sc.category_name ASC
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make([]models.ServiceCategory, 0)
	for rows.Next() {
		var cat models.ServiceCategory
		if err := rows.Scan(&cat.ID, &cat.CategoryName, &cat.TicketPrefix); err != nil {
			return nil, err
		}
		categories = append(categories, cat)
	}

	return categories, rows.Err()
}

// GetSummary mengambil ringkasan metrik laporan.
func (r *ReportRepository) GetSummary(params ReportQuery) (models.ReportSummary, error) {
	where, args := buildReportFilters(params)
	query := `
		SELECT
			COUNT(*) AS total_tickets,
			COALESCE(SUM(CASE WHEN qt.status = 'DONE' THEN 1 ELSE 0 END), 0) AS done_count,
			AVG(CASE
				WHEN qt.called_at IS NULL THEN NULL
				ELSE GREATEST(TIMESTAMPDIFF(SECOND, qt.created_at, qt.called_at), 0)
			END) AS avg_waiting_seconds,
			AVG(CASE
				WHEN qt.service_duration_seconds IS NULL THEN NULL
				ELSE qt.service_duration_seconds
			END) AS avg_service_seconds
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE ` + where

	var (
		summary    models.ReportSummary
		avgWaiting sql.NullFloat64
		avgService sql.NullFloat64
	)

	err := r.DB.QueryRow(query, args...).Scan(
		&summary.TotalTickets,
		&summary.DoneCount,
		&avgWaiting,
		&avgService,
	)
	if err != nil {
		return models.ReportSummary{}, err
	}

	if avgWaiting.Valid {
		summary.AvgWaitingSeconds = int(math.Round(avgWaiting.Float64))
	}
	if avgService.Valid {
		summary.AvgServiceSeconds = int(math.Round(avgService.Float64))
	}

	if summary.TotalTickets > 0 {
		summary.SuccessRate = float64(summary.DoneCount) / float64(summary.TotalTickets) * 100
	}

	return summary, nil
}

// GetPeakHour mengambil jam tersibuk berdasarkan jumlah tiket.
func (r *ReportRepository) GetPeakHour(params ReportQuery) (int, bool, error) {
	where, args := buildReportFilters(params)
	query := `
		SELECT
			HOUR(qt.created_at) AS peak_hour,
			COUNT(*) AS total_count
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE ` + where + `
		GROUP BY HOUR(qt.created_at)
		ORDER BY total_count DESC, peak_hour ASC
		LIMIT 1
	`

	var (
		peakHour sql.NullInt64
		total    sql.NullInt64
	)

	err := r.DB.QueryRow(query, args...).Scan(&peakHour, &total)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}

	if !peakHour.Valid {
		return 0, false, nil
	}

	return int(peakHour.Int64), true, nil
}

// GetQueueLogs mengambil data log antrian sesuai filter.
func (r *ReportRepository) GetQueueLogs(params ReportQuery, limit, offset int) ([]ReportQueueRow, int, error) {
	where, args := buildReportFilters(params)

	countQuery := `
		SELECT COUNT(*)
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE ` + where

	var total int
	if err := r.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT
			qt.ticket_no,
			COALESCE(sc.category_name, '') AS category_name,
			COALESCE(c.counter_name, '') AS counter_name,
			COALESCE(s.store_name, '') AS store_name,
			qt.created_at,
			qt.called_at,
			qt.done_at,
			qt.service_duration_seconds,
			qt.status
		FROM queue_tickets qt
		LEFT JOIN counters c ON c.id = qt.counter_id
		LEFT JOIN stores s ON s.store_id = qt.store_id
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		WHERE ` + where + `
		ORDER BY qt.created_at DESC, qt.id DESC
		LIMIT ? OFFSET ?
	`

	argsWithPaging := append([]interface{}{}, args...)
	argsWithPaging = append(argsWithPaging, limit, offset)

	rows, err := r.DB.Query(query, argsWithPaging...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]ReportQueueRow, 0)
	for rows.Next() {
		var (
			row      ReportQueueRow
			calledAt sql.NullTime
			doneAt   sql.NullTime
			duration sql.NullInt64
		)

		if err := rows.Scan(
			&row.TicketNo,
			&row.CategoryName,
			&row.CounterName,
			&row.StoreName,
			&row.CreatedAt,
			&calledAt,
			&doneAt,
			&duration,
			&row.Status,
		); err != nil {
			return nil, 0, err
		}

		if row.CategoryName == "" {
			row.CategoryName = row.CounterName
		}
		if calledAt.Valid {
			t := calledAt.Time
			row.CalledAt = &t
		}
		if doneAt.Valid {
			t := doneAt.Time
			row.DoneAt = &t
		}
		if duration.Valid {
			val := int(duration.Int64)
			row.DurationSeconds = &val
		}

		items = append(items, row)
	}

	return items, total, rows.Err()
}

func buildReportFilters(params ReportQuery) (string, []interface{}) {
	clauses := make([]string, 0)
	args := make([]interface{}, 0)

	if len(params.StoreIDs) > 0 {
		if params.StoreID > 0 {
			clauses = append(clauses, "qt.store_id = ?")
			args = append(args, params.StoreID)
		} else {
			placeholders := make([]string, len(params.StoreIDs))
			for i, id := range params.StoreIDs {
				placeholders[i] = "?"
				args = append(args, id)
			}
			clauses = append(clauses, fmt.Sprintf("qt.store_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if len(params.CounterIDs) > 0 {
		if params.CounterID > 0 {
			clauses = append(clauses, "qt.counter_id = ?")
			args = append(args, params.CounterID)
		} else {
			placeholders := make([]string, len(params.CounterIDs))
			for i, id := range params.CounterIDs {
				placeholders[i] = "?"
				args = append(args, id)
			}
			clauses = append(clauses, fmt.Sprintf("qt.counter_id IN (%s)", strings.Join(placeholders, ",")))
		}
	}

	if params.CategoryID > 0 {
		clauses = append(clauses, "sc.id = ?")
		args = append(args, params.CategoryID)
	}

	if !params.StartDate.IsZero() {
		clauses = append(clauses, "qt.ticket_date >= ?")
		args = append(args, params.StartDate.Format("2006-01-02"))
	}
	if !params.EndDate.IsZero() {
		clauses = append(clauses, "qt.ticket_date <= ?")
		args = append(args, params.EndDate.Format("2006-01-02"))
	}

	if len(clauses) == 0 {
		return "1=1", args
	}

	return strings.Join(clauses, " AND "), args
}
