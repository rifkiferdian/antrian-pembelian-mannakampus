package models

// QueueViewPayload dikirim ke websocket view_queue untuk update realtime.
type QueueViewPayload struct {
	Type        string             `json:"type"`
	Action      string             `json:"action"`
	StoreID     int                `json:"store_id"`
	Date        string             `json:"date"`
	CurrentCall QueueViewCall      `json:"current_call"`
	Counters    []QueueViewCounter `json:"counters"`
}
