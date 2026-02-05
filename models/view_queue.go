package models

// QueueViewCall menampung data nomor yang sedang dipanggil.
type QueueViewCall struct {
	TicketNo     string `json:"ticket_no"`
	CounterLabel string `json:"counter_label"`
	CategoryName string `json:"category_name"`
}

// QueueViewCounter menampung data status tiap loket pada layar antrean.
type QueueViewCounter struct {
	CounterID    int    `json:"counter_id"`
	CounterName  string `json:"counter_name"`
	CounterCode  string `json:"counter_code"`
	CategoryName string `json:"category_name"`
	TicketNo     string `json:"ticket_no"`
	TicketStatus string `json:"ticket_status"`
	CounterLabel string `json:"counter_label"`
	StatusLabel  string `json:"status_label"`
	StatusClass  string `json:"status_class"`
	IsCurrent    bool   `json:"is_current"`
}
