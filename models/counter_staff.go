package models

import "time"

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

// CounterStaffStatusUpdateInput menampung data perubahan status staff di dashboard.
type CounterStaffStatusUpdateInput struct {
	CounterID            int
	UserID               int
	Status               string
	InactiveStartedAt    *time.Time
	InactiveUntil        *time.Time
	InactiveAnnouncement string
}

// CounterStaffStatusDetail menampung detail status staff untuk kebutuhan dashboard.
type CounterStaffStatusDetail struct {
	Status               string
	InactiveStartedAt    string
	InactiveUntil        string
	InactiveAnnouncement string
}
