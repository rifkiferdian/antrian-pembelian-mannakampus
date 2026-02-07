package models

// DashboardServing berisi data tiket yang sedang dipanggil di dashboard.
type DashboardServing struct {
	TicketNo      string `json:"ticket_no"`
	CategoryName  string `json:"category_name"`
	CounterName   string `json:"counter_name"`
	CalledAt      string `json:"called_at"`
	CalledAtUnix  int64  `json:"called_at_unix"`
	DurationLabel string `json:"duration_label"`
}

// DashboardQueueItem menampung daftar antrian menunggu di sidebar dashboard.
type DashboardQueueItem struct {
	TicketID       int64  `json:"ticket_id"`
	TicketNo       string `json:"ticket_no"`
	CategoryName   string `json:"category_name"`
	QueueNumber    int    `json:"queue_number"`
	WaitingMinutes int    `json:"waiting_minutes"`
	Status         string `json:"status"`
	StatusLabel    string `json:"status_label"`
	StatusClass    string `json:"status_class"`
}
