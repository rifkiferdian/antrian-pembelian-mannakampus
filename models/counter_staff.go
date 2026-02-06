package models

// CounterStaff merepresentasikan data pada tabel counter_staffs.
type CounterStaff struct {
	ID               int
	CounterID        int
	CounterCode      string
	CounterName      string
	StoreID          int
	StoreName        string
	UserID           int
	UserName         string
	Username         string
	Status           string
	StatusLabel      string
	CreatedAt        string
	CreatedAtDisplay string
}

// CounterStaffCreateInput menampung data yang dikirimkan dari form create counter staff.
type CounterStaffCreateInput struct {
	CounterID int
	UserID    int
	Status    string
}

// CounterStaffUpdateInput menampung data yang dikirimkan dari form edit counter staff.
type CounterStaffUpdateInput struct {
	ID        int
	CounterID int
	UserID    int
	Status    string
}
