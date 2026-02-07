package models

// ServiceCategory merepresentasikan kategori layanan.
type ServiceCategory struct {
	ID           int
	CategoryName string
	TicketPrefix string
}

// ReportSummary menampung ringkasan metrik laporan.
type ReportSummary struct {
	TotalTickets      int
	DoneCount         int
	SuccessRate       float64
	AvgWaitingSeconds int
	AvgServiceSeconds int
	AvgWaitingLabel   string
	AvgServiceLabel   string
	SuccessRateLabel  string
	PeakHourLabel     string
}

// ReportQueueItem menampung data log antrian untuk laporan.
type ReportQueueItem struct {
	TicketNo      string
	CategoryName  string
	CounterName   string
	StoreName     string
	ArrivalLabel  string
	CallLabel     string
	DoneLabel     string
	DurationLabel string
	Status        string
}

// ReportPagination menampung data pagination untuk laporan.
type ReportPagination struct {
	Page       int
	PerPage    int
	Total      int
	TotalPages int
	From       int
	To         int
	PrevURL    string
	NextURL    string
	HasPrev    bool
	HasNext    bool
}
