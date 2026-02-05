package models

// QueueViewCall menampung data nomor yang sedang dipanggil.
type QueueViewCall struct {
	TicketNo     string
	CounterLabel string
	CategoryName string
}

// QueueViewCounter menampung data status tiap loket pada layar antrean.
type QueueViewCounter struct {
	CounterID    int
	CounterName  string
	CounterCode  string
	CategoryName string
	TicketNo     string
	TicketStatus string
	CounterLabel string
	StatusLabel  string
	StatusClass  string
	IsCurrent    bool
}
