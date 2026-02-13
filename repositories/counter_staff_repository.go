package repositories

import (
	"database/sql"
	"stok-hadiah/models"
	"strings"
)

type CounterStaffRepository struct {
	DB *sql.DB
}

// GetByStoreIDs mengambil data counter staff berdasarkan store_id tertentu.
func (r *CounterStaffRepository) GetByStoreIDs(storeIDs []int) ([]models.CounterStaff, error) {
	if len(storeIDs) == 0 {
		return []models.CounterStaff{}, nil
	}

	placeholders := make([]string, len(storeIDs))
	args := make([]interface{}, len(storeIDs))
	for i, id := range storeIDs {
		placeholders[i] = "?"
		args[i] = id
	}

	query := `
		SELECT
			cs.id,
			cs.counter_id,
			c.store_id,
			COALESCE(s.store_name, '') AS store_name,
			COALESCE(c.counter_code, '') AS counter_code,
			COALESCE(c.counter_name, '') AS counter_name,
			cs.user_id,
			COALESCE(u.name, '') AS user_name,
			COALESCE(u.username, '') AS username,
			cs.status,
			cs.created_at
		FROM counter_staffs cs
		LEFT JOIN counters c ON c.id = cs.counter_id
		LEFT JOIN stores s ON s.store_id = c.store_id
		LEFT JOIN users u ON u.id = cs.user_id
		WHERE c.store_id IN (` + strings.Join(placeholders, ",") + `)
		ORDER BY cs.created_at DESC, cs.id DESC
	`

	rows, err := r.DB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var staffs []models.CounterStaff

	for rows.Next() {
		var (
			staff     models.CounterStaff
			createdAt sql.NullTime
		)

		if err := rows.Scan(
			&staff.ID,
			&staff.CounterID,
			&staff.StoreID,
			&staff.StoreName,
			&staff.CounterCode,
			&staff.CounterName,
			&staff.UserID,
			&staff.UserName,
			&staff.Username,
			&staff.Status,
			&createdAt,
		); err != nil {
			return nil, err
		}

		staff.StatusLabel = resolveCounterStaffStatusLabel(staff.Status)

		if createdAt.Valid {
			staff.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
			staff.CreatedAtDisplay = createdAt.Time.Format("02 Jan 2006 15:04:05")
		} else {
			staff.CreatedAt = "-"
			staff.CreatedAtDisplay = "-"
		}

		if staff.StoreName == "" {
			staff.StoreName = "-"
		}
		if staff.CounterCode == "" {
			staff.CounterCode = "-"
		}
		if staff.CounterName == "" {
			staff.CounterName = "-"
		}
		if staff.UserName == "" {
			staff.UserName = "-"
		}
		if staff.Username == "" {
			staff.Username = "-"
		}

		staffs = append(staffs, staff)
	}

	return staffs, rows.Err()
}

// Create menyimpan data counter staff baru.
func (r *CounterStaffRepository) Create(params models.CounterStaffCreateInput) error {
	_, err := r.DB.Exec(`
		INSERT INTO counter_staffs (counter_id, user_id, status)
		VALUES (?, ?, ?)
	`, params.CounterID, params.UserID, params.Status)
	return err
}

// Update memperbarui data counter staff.
func (r *CounterStaffRepository) Update(params models.CounterStaffUpdateInput) error {
	_, err := r.DB.Exec(`
		UPDATE counter_staffs
		SET counter_id = ?, user_id = ?, status = ?
		WHERE id = ?
	`, params.CounterID, params.UserID, params.Status, params.ID)
	return err
}

// GetStatusByCounterAndUser mengambil status counter staff berdasarkan counter dan user.
func (r *CounterStaffRepository) GetStatusByCounterAndUser(counterID, userID int) (string, bool, error) {
	if counterID <= 0 || userID <= 0 {
		return "", false, nil
	}

	var status string
	err := r.DB.QueryRow(`
		SELECT status
		FROM counter_staffs
		WHERE counter_id = ? AND user_id = ?
		LIMIT 1
	`, counterID, userID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", false, nil
		}
		return "", false, err
	}

	return status, true, nil
}

// GetStatusDetailByCounterAndUser mengambil detail status staff berdasarkan counter dan user.
func (r *CounterStaffRepository) GetStatusDetailByCounterAndUser(counterID, userID int) (models.CounterStaffStatusDetail, bool, error) {
	if counterID <= 0 || userID <= 0 {
		return models.CounterStaffStatusDetail{}, false, nil
	}

	var (
		detail               models.CounterStaffStatusDetail
		inactiveStartedAt    sql.NullTime
		inactiveUntil        sql.NullTime
		inactiveAnnouncement sql.NullString
	)

	err := r.DB.QueryRow(`
		SELECT
			status,
			inactive_started_at,
			inactive_until,
			inactive_announcement
		FROM counter_staffs
		WHERE counter_id = ? AND user_id = ?
		LIMIT 1
	`, counterID, userID).Scan(
		&detail.Status,
		&inactiveStartedAt,
		&inactiveUntil,
		&inactiveAnnouncement,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.CounterStaffStatusDetail{}, false, nil
		}
		return models.CounterStaffStatusDetail{}, false, err
	}

	if inactiveStartedAt.Valid {
		detail.InactiveStartedAt = inactiveStartedAt.Time.Format("2006-01-02T15:04")
	}
	if inactiveUntil.Valid {
		detail.InactiveUntil = inactiveUntil.Time.Format("2006-01-02T15:04")
	}
	if inactiveAnnouncement.Valid {
		detail.InactiveAnnouncement = strings.TrimSpace(inactiveAnnouncement.String)
	}

	return detail, true, nil
}

// UpdateStatusByCounterAndUser memperbarui status staff berdasarkan counter dan user.
func (r *CounterStaffRepository) UpdateStatusByCounterAndUser(input models.CounterStaffStatusUpdateInput) (bool, error) {
	if input.CounterID <= 0 || input.UserID <= 0 {
		return false, nil
	}

	var inactiveStartedAt interface{}
	var inactiveUntil interface{}
	if input.InactiveStartedAt != nil {
		inactiveStartedAt = input.InactiveStartedAt.Format("2006-01-02 15:04:05")
	}
	if input.InactiveUntil != nil {
		inactiveUntil = input.InactiveUntil.Format("2006-01-02 15:04:05")
	}

	var inactiveAnnouncement interface{}
	if strings.TrimSpace(input.InactiveAnnouncement) != "" {
		inactiveAnnouncement = strings.TrimSpace(input.InactiveAnnouncement)
	}

	res, err := r.DB.Exec(`
		UPDATE counter_staffs
		SET
			status = ?,
			inactive_started_at = ?,
			inactive_until = ?,
			inactive_announcement = ?
		WHERE counter_id = ? AND user_id = ?
	`,
		input.Status,
		inactiveStartedAt,
		inactiveUntil,
		inactiveAnnouncement,
		input.CounterID,
		input.UserID,
	)
	if err != nil {
		return false, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return false, err
	}

	return affected > 0, nil
}

// Delete menghapus counter staff berdasarkan ID.
func (r *CounterStaffRepository) Delete(id int) error {
	_, err := r.DB.Exec(`DELETE FROM counter_staffs WHERE id = ?`, id)
	return err
}

// ExistsInStores mengecek apakah counter_staff id ada pada store yang diizinkan.
func (r *CounterStaffRepository) ExistsInStores(id int, storeIDs []int) (bool, error) {
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

	query := `
		SELECT 1
		FROM counter_staffs cs
		INNER JOIN counters c ON c.id = cs.counter_id
		WHERE cs.id = ? AND c.store_id IN (` + strings.Join(placeholders, ",") + `)
		LIMIT 1
	`
	var dummy int
	if err := r.DB.QueryRow(query, args...).Scan(&dummy); err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func resolveCounterStaffStatusLabel(status string) string {
	switch strings.ToUpper(strings.TrimSpace(status)) {
	case "ACTIVE":
		return "Aktif"
	case "REST":
		return "Istirahat"
	case "INACTIVE":
		return "Non Aktif"
	default:
		return "-"
	}
}
