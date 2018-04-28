package models

// ClientCreate - json data expected for creating a new client
type ClientCreate struct {
	IP   string `json:"ip"`
	Name string `json:"name"`
}

// ClientID
type ClientID struct {
	ID string `json:"id"`
}
