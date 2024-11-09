package models

const TicketCachePrefix = "ticket:"

type Ticket struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Allocation  int    `json:"allocation"`
}

type PurchaseRequest struct {
	Quantity int `json:"quantity"`
}
