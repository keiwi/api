package models

// ChecksID
type ChecksID struct {
	ID string `json:"id"`
}

type ChecksWithClientCommandID struct {
	ClientID  string   `json:"client_id"`
	CommandID []string `json:"command_id"`
}

type ChecksBetweenDateClient struct {
	ClientID  string `json:"client_id"`
	CommandID string `json:"command_id"`
	From      string `json:"from"`
	To        string `json:"to"`
	Max       int    `json:"max"`
}
