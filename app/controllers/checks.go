package controllers

import (
	"fmt"
	"time"

	"aahframework.org/aah.v0"
	"aahframework.org/log.v0"
	"github.com/keiwi/api/app/models"
	"github.com/keiwi/utils"
	utilModels "github.com/keiwi/utils/models"
	utilNats "github.com/keiwi/utils/nats"
	"gopkg.in/mgo.v2/bson"
)

// ChecksController controller for check API related methods
type ChecksController struct {
	*aah.Context
}

// DeleteCheck removes a check from the database with a specific ID
func (a *ChecksController) DeleteCheck(delete models.ChecksID) {
	// Check if ID is provided and if it's a valid ObjectIdHex
	if delete.ID != "" || bson.IsObjectIdHex(delete.ID) {
		a.Reply().BadRequest().JSON(models.Response{Message: "ID is not a valid ObjectId"})
		return
	}

	n := models.Conn
	if n == nil {
		log.Error("nats is not initialized")
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Initialize delete data for nats
	del := utils.DeleteOptions{
		Filter: utils.Filter{"_id": bson.ObjectIdHex(delete.ID)},
	}

	// Marshal the delete data
	data, err := bson.MarshalJSON(del)
	if err != nil {
		log.Debugf("error marshaling data: %v", err)
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Send the delete data to nats
	err = utilNats.DeleteCheck(n, data)
	if err != nil {
		log.Debugf("error deleting check: %v", err)
		a.Reply().BadRequest().JSON(models.Response{Message: "Internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully deleted the check"})
}

// GetChecks returns all existing checks in the database
func (a *ChecksController) GetChecks() {
	n := models.Conn
	if n == nil {
		log.Error("nats is not initialized")
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Initialize data for finding all existing checks
	find := utils.FindOptions{
		Sort: utils.Sort{"-created_at"},
	}

	// Marshal finding data
	data, err := bson.MarshalJSON(find)
	if err != nil {
		log.Debugf("error marshaling data: %v", err)
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Send the data to nats
	checks, err := utilNats.FindCheck(n, data)
	if err != nil {
		log.Debugf("error finding check: %v", err)
		a.Reply().BadRequest().JSON(models.Response{Message: "Internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found all checks in database", Data: checks})
}

// GetCheckWithID returns a check if ID exists
func (a *ChecksController) GetCheckWithID(check models.ChecksID) {
	// Check if ID is provided and if it's a valid ObjectIdHex
	if check.ID != "" || bson.IsObjectIdHex(check.ID) {
		a.Reply().BadRequest().JSON(models.Response{Message: "ID is not a valid ObjectId"})
		return
	}

	n := models.Conn
	if n == nil {
		log.Error("nats is not initialized")
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Initialize data for finding all checks with a specific ID
	find := utils.FindOptions{
		Filter: utils.Filter{"_id": bson.ObjectIdHex(check.ID)},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	// Marshal the data
	data, err := bson.MarshalJSON(find)
	if err != nil {
		log.Debugf("error marshaling data: %v", err)
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Send the data to nats
	checks, err := utilNats.FindCheck(n, data)
	if err != nil {
		log.Debugf("error finding check: %v", err)
		a.Reply().BadRequest().JSON(models.Response{Message: "Internal error"})
		return
	}
	if len(checks) <= 0 {
		a.Reply().BadRequest().JSON(models.Response{Message: "Could not find any checks"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found the check", Data: checks[0]})
}

// GetWithClientIDAndCommandID tries to find checks with client id and command id
func (a *ChecksController) GetWithClientIDAndCommandID(c models.ChecksWithClientCommandID) {
	// Check if ClientID is provided and if it's a valid ObjectIdHex
	if c.ClientID != "" || bson.IsObjectIdHex(c.ClientID) {
		a.Reply().BadRequest().JSON(models.Response{Message: "Client ID is not a valid ObjectId"})
		return
	}

	// Loop through all CommandIDs and check if they are valid ObjectIdHex
	for _, id := range c.CommandID {
		if !bson.IsObjectIdHex(id) {
			a.Reply().BadRequest().JSON(models.Response{Message: fmt.Sprintf("The command id \"%s\" is not a valid ObjectId", id)})
			return
		}
	}

	n := models.Conn
	if n == nil {
		log.Error("nats is not initialized")
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Loop through all CommandIDs and attempt to find existing data from the database
	var checks []utilModels.Check
	for _, cmd := range c.CommandID {
		// Initialize data for finding all checks with a command ID and client ID
		find := utils.FindOptions{
			Filter: utils.Filter{"command_id": bson.ObjectIdHex(cmd), "client_id": bson.ObjectIdHex(c.ClientID)},
			Sort:   utils.Sort{"-created_at"},
			Limit:  1,
		}

		// Marshal the data
		data, err := bson.MarshalJSON(find)
		if err != nil {
			log.Debugf("error marshaling data: %v", err)
			a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
			return
		}

		// Send the data to nats
		cc, err := utilNats.FindCheck(n, data)
		if err != nil {
			log.Debugf("error finding check with client and command id: %v", err)
			a.Reply().BadRequest().JSON(models.Response{Message: "Internal error"})
			return
		}
		if len(cc) >= 1 {
			checks = append(checks, cc[0])
		}
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found the check", Data: checks})
}

// GetWithChecksBetweenDateClient tries to find checks between dates with client id
func (a *ChecksController) GetWithChecksBetweenDateClient(c models.ChecksBetweenDateClient) {
	// Check if CommandID is provided and if it's a valid ObjectIdHex
	if c.CommandID != "" || bson.IsObjectIdHex(c.CommandID) {
		a.Reply().BadRequest().JSON(models.Response{Message: "Command ID is not a valid ObjectId"})
		return
	}

	// Check if ClientID is provided and if it's a valid ObjectIdHex
	if c.ClientID != "" || bson.IsObjectIdHex(c.ClientID) {
		a.Reply().BadRequest().JSON(models.Response{Message: "Client ID is not a valid ObjectId"})
		return
	}

	// Attempt to parse the time
	from, err := time.Parse("2006-01-02 15:04:05", c.From)
	if err != nil {
		log.Debugf("error parsing time (from): %v", err)
		// TODO: Maybe consider returning the time format?
		a.Reply().InternalServerError().JSON(models.Response{Message: "Invalid time format (from)"})
		return
	}
	to, err := time.Parse("2006-01-02 15:04:05", c.To)
	if err != nil {
		log.Debugf("error parsing time (from): %v", err)
		// TODO: Maybe consider returning the time format?
		a.Reply().InternalServerError().JSON(models.Response{Message: "Invalid time format (to)"})
		return
	}

	n := models.Conn
	if n == nil {
		log.Error("nats is not initialized")
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Initialize data for finding the checks between dates
	find := utils.FindOptions{
		Filter: utils.Filter{"command_id": bson.ObjectIdHex(c.CommandID), "client_id": bson.ObjectIdHex(c.ClientID), "created_at": bson.M{"$gte": from, "$lte": to}},
		Sort:   utils.Sort{"created_at"},
		Max:    utils.Max(c.Max),
	}

	// Marshal the data
	data, err := bson.MarshalJSON(find)
	if err != nil {
		log.Debugf("error marshaling data: %v", err)
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Send the data to nats
	checks, err := utilNats.FindCheck(n, data)
	if err != nil {
		log.Debugf("error finding checks: %v", err)
		a.Reply().BadRequest().JSON(models.Response{Message: "Internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found checks", Data: checks})
}
