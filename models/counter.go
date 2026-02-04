package models

// Counter merepresentasikan data pada tabel counters.
type Counter struct {
	ID               int
	StoreID          int
	StoreName        string
	CounterCode      string
	CounterName      string
	TicketPrefix     string
	IsActive         bool
	StatusLabel      string
	CreatedAt        string
	CreatedAtDisplay string
}

// CounterCreateInput menampung data yang dikirimkan dari form create counter.
type CounterCreateInput struct {
	StoreID      int
	CounterCode  string
	CounterName  string
	TicketPrefix string
	IsActive     bool
}

// CounterUpdateInput menampung data yang dikirimkan dari form edit counter.
type CounterUpdateInput struct {
	ID           int
	StoreID      int
	CounterCode  string
	CounterName  string
	TicketPrefix string
	IsActive     bool
}
