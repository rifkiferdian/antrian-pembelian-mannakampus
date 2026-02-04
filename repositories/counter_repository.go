package repositories

import (
	"database/sql"
	"stok-hadiah/models"
	"strings"
)

type CounterRepository struct {
	DB *sql.DB
}

// GetAll mengambil seluruh data counter beserta nama store.
func (r *CounterRepository) GetAll() ([]models.Counter, error) {
	rows, err := r.DB.Query(`
		SELECT 
			c.id,
			c.store_id,
			COALESCE(s.store_name, '') AS store_name,
			c.counter_code,
			c.counter_name,
			c.ticket_prefix,
			c.is_active,
			c.created_at
		FROM counters c
		LEFT JOIN stores s ON s.store_id = c.store_id
		ORDER BY c.created_at DESC, c.id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []models.Counter

	for rows.Next() {
		var (
			counter   models.Counter
			isActive  int
			createdAt sql.NullTime
		)

		if err := rows.Scan(
			&counter.ID,
			&counter.StoreID,
			&counter.StoreName,
			&counter.CounterCode,
			&counter.CounterName,
			&counter.TicketPrefix,
			&isActive,
			&createdAt,
		); err != nil {
			return nil, err
		}

		counter.IsActive = isActive == 1
		if counter.IsActive {
			counter.StatusLabel = "Aktif"
		} else {
			counter.StatusLabel = "Non Aktif"
		}

		if createdAt.Valid {
			counter.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			counter.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			counter.CreatedAt = "-"
			counter.CreatedAtDisplay = "-"
		}

		if counter.StoreName == "" {
			counter.StoreName = "-"
		}

		counters = append(counters, counter)
	}

	return counters, rows.Err()
}

// GetByStoreIDs mengambil data counter berdasarkan store_id tertentu.
func (r *CounterRepository) GetByStoreIDs(storeIDs []int) ([]models.Counter, error) {
	if len(storeIDs) == 0 {
		return []models.Counter{}, nil
	}

	placeholders := make([]string, len(storeIDs))
	args := make([]interface{}, len(storeIDs))
	for i, id := range storeIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT 
			c.id,
			c.store_id,
			COALESCE(s.store_name, '') AS store_name,
			c.counter_code,
			c.counter_name,
			c.ticket_prefix,
			c.is_active,
			c.created_at
		FROM counters c
		LEFT JOIN stores s ON s.store_id = c.store_id
		WHERE c.store_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY c.created_at DESC, c.id DESC
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []models.Counter

	for rows.Next() {
		var (
			counter   models.Counter
			isActive  int
			createdAt sql.NullTime
		)

		if err := rows.Scan(
			&counter.ID,
			&counter.StoreID,
			&counter.StoreName,
			&counter.CounterCode,
			&counter.CounterName,
			&counter.TicketPrefix,
			&isActive,
			&createdAt,
		); err != nil {
			return nil, err
		}

		counter.IsActive = isActive == 1
		if counter.IsActive {
			counter.StatusLabel = "Aktif"
		} else {
			counter.StatusLabel = "Non Aktif"
		}

		if createdAt.Valid {
			counter.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			counter.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			counter.CreatedAt = "-"
			counter.CreatedAtDisplay = "-"
		}

		if counter.StoreName == "" {
			counter.StoreName = "-"
		}

		counters = append(counters, counter)
	}

	return counters, rows.Err()
}

// Create menyimpan data counter baru.
func (r *CounterRepository) Create(params models.CounterCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO counters (store_id, counter_code, counter_name, ticket_prefix, is_active)
		VALUES (?, ?, ?, ?, ?)
	`, params.StoreID, params.CounterCode, params.CounterName, params.TicketPrefix, boolToInt(params.IsActive))
	return err
}

// Update memperbarui data counter.
func (r *CounterRepository) Update(params models.CounterUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE counters
		SET store_id = ?, counter_code = ?, counter_name = ?, ticket_prefix = ?, is_active = ?
		WHERE id = ?
	`, params.StoreID, params.CounterCode, params.CounterName, params.TicketPrefix, boolToInt(params.IsActive), params.ID)
	return err
}

// Delete menghapus counter berdasarkan ID.
func (r *CounterRepository) Delete(id int) error {
	_, err := r.DB.Exec(`DELETE FROM counters WHERE id = ?`, id)
	return err
}

// ExistsInStores mengecek apakah counter id ada pada store yang diizinkan.
func (r *CounterRepository) ExistsInStores(id int, storeIDs []int) (bool, error) {
	if len(storeIDs) == 0 {
		return false, nil
	}

	placeholders := make([]string, len(storeIDs))
	args := make([]interface{}, 0, len(storeIDs)+1)
	args = append(args, id)
	for i, storeID := range storeIDs {
		placeholders[i] = "?"
		args = append(args, storeID)
	}

	query := `SELECT 1 FROM counters WHERE id = ? AND store_id IN (` + strings.Join(placeholders, ",") + `) LIMIT 1`
	var dummy int
	if err := r.DB.QueryRow(query, args...).Scan(&dummy); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func boolToInt(val bool) int {
	if val {
		return 1
	}
	return 0
}
