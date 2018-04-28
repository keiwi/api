package models

type GroupRename struct {
	NewName string `json:"new_name"`
	OldName string `json:"old_name"`
}

type GroupCreate struct {
	GroupName string `json:"group_name"`
	CommandID string `json:"command_id"`
	Delay     int    `json:"delay"`
	StopError bool   `json:"stop_error"`
}

type GroupID struct {
	ID string `json:"id"`
}

type GroupName struct {
	Name string `json:"name"`
}
