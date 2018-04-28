package models

type CommandCreate struct {
	Command     string `json:"command"`
	Name        string `json:"namn"`
	Description string `json:"description"`
	Format      string `json:"format"`
}

type CommandID struct {
	ID string `json:"id"`
}
