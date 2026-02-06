package models

// GuestQueueCounter menampung data counter untuk tampilan halaman tamu.
type GuestQueueCounter struct {
	CounterID    int
	CounterName  string
	CounterCode  string
	TicketPrefix string
	CategoryName string
	WaitingCount int
	StaffNames   []string
	Icon         string
	IndexLabel   string
}

// GuestTicket merepresentasikan tiket antrian yang dicetak oleh tamu.
type GuestTicket struct {
	ID           int64
	StoreID      int
	StoreName    string
	CounterID    int
	CounterName  string
	CounterCode  string
	CategoryName string
	TicketNo     string
	QueueNumber  int
	TicketDate   string
	CreatedAt    string
	CreatedAtDisplay string
}
