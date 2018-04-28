package controllers

import (
	"strings"

	"aahframework.org/aah.v0"
	"github.com/keiwi/api/app/models"
	"github.com/keiwi/utils"
	storageModel "github.com/keiwi/utils/models"
	utilNats "github.com/keiwi/utils/nats"
	"github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2/bson"
)

type CommandsController struct {
	*aah.Context
}

func (a *CommandsController) CreateCommand(create models.CommandCreate) {
	if create.Command == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Command is missing"})
		return
	}
	if create.Name == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Name is missing"})
		return
	}
	if create.Description == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Description is missing"})
		return
	}

	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	cmd := storageModel.Command{
		Command:     create.Command,
		Name:        create.Name,
		Description: create.Description,
		Format:      create.Format,
	}

	data, err := bson.MarshalJSON(cmd)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	err = utilNats.CreateClient(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully created the command", Data: cmd})
}

// EditCommand modifies an existing client in the database
func (a *CommandsController) EditCommand(edit models.EditRequest) {
	// retrieve nats instance
	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// retrieve existing command
	find := utils.FindOptions{
		Filter: utils.Filter{"_id": bson.ObjectIdHex(edit.ID)},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	data, err := bson.MarshalJSON(find)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// check if command actually exists
	commands, err := utilNats.FindCommand(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}
	if len(commands) <= 0 {
		a.Reply().BadRequest().JSON(models.Response{Message: "Can't find a client with this ID"})
		return
	}

	cmd := commands[0]
	v, ok := edit.Value.(string)
	if !ok {
		a.Reply().BadRequest().JSON(models.Response{Message: "Value is not a string"})
		return
	}

	// start parsing the update
	updates := bson.M{}
	switch strings.ToLower(edit.Option) {
	case "command":
		updates["command"] = v
		cmd.Command = v
	case "name", "namn":
		updates["name"] = v
		cmd.Name = v
	case "description":
		updates["description"] = v
		cmd.Description = v
	case "format":
		updates["format"] = v
		cmd.Format = v
	default:
		a.Reply().BadRequest().JSON(models.Response{Message: "Please provide a correct column"})
		return
	}

	// send the updates to nats
	update := utils.UpdateOptions{
		Updates: utils.Updates{"$set": updates},
	}

	updateData, err := bson.MarshalJSON(update)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	err = utilNats.UpdateCommand(n, updateData)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// if everything went well, respond with success
	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully saved the changes for the command", Data: cmd})
}

// DeleteCommand deletes a specific client from the database
func (a *CommandsController) DeleteCommand(delete models.CommandID) {
	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	del := utils.DeleteOptions{
		Filter: utils.Filter{"_id": bson.ObjectIdHex(delete.ID)},
	}

	data, err := bson.MarshalJSON(del)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	err = utilNats.DeleteCommand(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully deleted the command"})
}

// GetCommands returns an array of all the clients in the database
func (a *CommandsController) GetCommands() {
	inter := a.Get("nats")
	if inter == nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	n, ok := inter.(*nats.Conn)
	if !ok {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	find := utils.FindOptions{
		Sort: utils.Sort{"-created_at"},
	}

	data, err := bson.MarshalJSON(find)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	commands, err := utilNats.FindClient(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found all commands in database", Data: commands})
}
