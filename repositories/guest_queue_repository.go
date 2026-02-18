package repositories

import (
	"database/sql"
	"stok-hadiah/models"
	"strings"
	"time"
)

type GuestQueueRepository struct {
	DB *sql.DB
}

func (r *GuestQueueRepository) activateExpiredInactiveByStore(storeID int) error {
	if storeID <= 0 {
		return nil
	}

	_, err := r.DB.Exec(`
		UPDATE counter_staffs cs
		INNER JOIN counters c ON c.id = cs.counter_id
		SET
			cs.status = 'ACTIVE',
			cs.inactive_started_at = NULL,
			cs.inactive_until = NULL,
			cs.inactive_announcement = NULL
		WHERE c.store_id = ?
		  AND cs.status = 'INACTIVE'
		  AND cs.inactive_until IS NOT NULL
		  AND cs.inactive_until <= NOW()
	`, storeID)
	return err
}

// GetCountersForGuest mengambil daftar counter aktif beserta kategori dan jumlah antrian menunggu.
func (r *GuestQueueRepository) GetCountersForGuest(storeID int, ticketDate time.Time) ([]models.GuestQueueCounter, error) {
	if err := r.activateExpiredInactiveByStore(storeID); err != nil {
		return nil, err
	}

	rows, err := r.DB.Query(`
		SELECT
			c.id,
			c.counter_name,
			c.counter_code,
			c.ticket_prefix,
			COALESCE(sc.category_name, '') AS category_name,
			COALESCE(GROUP_CONCAT(DISTINCT u.name ORDER BY u.name SEPARATOR '||'), '') AS staff_names,
			CASE
				WHEN EXISTS (
					SELECT 1 FROM counter_staffs cs_active
					WHERE cs_active.counter_id = c.id AND cs_active.status = 'ACTIVE'
				) THEN 'ACTIVE'
				WHEN EXISTS (
					SELECT 1 FROM counter_staffs cs_inactive
					WHERE cs_inactive.counter_id = c.id AND cs_inactive.status = 'INACTIVE'
				) THEN 'INACTIVE'
				WHEN EXISTS (
					SELECT 1 FROM counter_staffs cs_rest
					WHERE cs_rest.counter_id = c.id AND cs_rest.status = 'REST'
				) THEN 'REST'
				ELSE 'INACTIVE'
			END AS staff_status,
			COALESCE((
				SELECT DATE_FORMAT(cs_inactive.inactive_until, '%Y-%m-%dT%H:%i:%s')
				FROM counter_staffs cs_inactive
				WHERE cs_inactive.counter_id = c.id
				  AND cs_inactive.status = 'INACTIVE'
				  AND cs_inactive.inactive_until IS NOT NULL
				ORDER BY cs_inactive.inactive_until DESC, cs_inactive.id DESC
				LIMIT 1
			), '') AS inactive_until,
			COALESCE((
				SELECT cs_inactive.inactive_announcement
				FROM counter_staffs cs_inactive
				WHERE cs_inactive.counter_id = c.id
				  AND cs_inactive.status = 'INACTIVE'
				  AND COALESCE(cs_inactive.inactive_announcement, '') <> ''
				ORDER BY cs_inactive.id DESC
				LIMIT 1
			), '') AS inactive_announcement,
			(
				SELECT COUNT(*)
				FROM queue_tickets qt
				WHERE qt.store_id = c.store_id
				  AND qt.counter_id = c.id
				  AND qt.status = 'WAITING'
				  AND qt.ticket_date = ?
			) AS waiting_count
		FROM counters c
		LEFT JOIN service_categories sc ON sc.ticket_prefix = c.ticket_prefix
		LEFT JOIN counter_staffs cs ON cs.counter_id = c.id AND cs.status IN ('ACTIVE', 'REST')
		LEFT JOIN users u ON u.id = cs.user_id
		WHERE c.store_id = ? AND c.is_active = 1
		GROUP BY c.id, c.counter_name, c.counter_code, c.ticket_prefix, sc.category_name
		ORDER BY c.id ASC
	`, ticketDate.Format("2006-01-02"), storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counters []models.GuestQueueCounter
	for rows.Next() {
		var c models.GuestQueueCounter
		var staffNames string
		if err := rows.Scan(
			&c.CounterID,
			&c.CounterName,
			&c.CounterCode,
			&c.TicketPrefix,
			&c.CategoryName,
			&staffNames,
			&c.StaffStatus,
			&c.InactiveUntil,
			&c.InactiveAnnouncement,
			&c.WaitingCount,
		); err != nil {
			return nil, err
		}
		if c.CategoryName == "" {
			c.CategoryName = c.CounterName
		}
		c.StaffStatus = strings.ToUpper(strings.TrimSpace(c.StaffStatus))
		if staffNames != "" {
			c.StaffNames = strings.Split(staffNames, "||")
		}
		c.InactiveAnnouncement = strings.TrimSpace(c.InactiveAnnouncement)
		counters = append(counters, c)
	}

	return counters, rows.Err()
}
