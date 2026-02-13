package models

// GuestQueueCounter menampung data counter untuk tampilan halaman tamu.
type GuestQueueCounter struct {
	CounterID            int      `json:"counter_id"`
	CounterName          string   `json:"counter_name"`
	CounterCode          string   `json:"counter_code"`
	TicketPrefix         string   `json:"ticket_prefix"`
	CategoryName         string   `json:"category_name"`
	WaitingCount         int      `json:"waiting_count"`
	StaffNames           []string `json:"staff_names"`
	StaffStatus          string   `json:"staff_status"`
	InactiveUntil        string   `json:"inactive_until"`
	InactiveAnnouncement string   `json:"inactive_announcement"`
	Icon                 string   `json:"icon"`
	IndexLabel           string   `json:"index_label"`
}

// GuestTicket merepresentasikan tiket antrian yang dicetak oleh tamu.
type GuestTicket struct {
	ID               int64
	StoreID          int
	StoreName        string
	CounterID        int
	CounterName      string
	CounterCode      string
	CategoryName     string
	TicketNo         string
	QueueNumber      int
	TicketDate       string
	CreatedAt        string
	CreatedAtDisplay string
}
