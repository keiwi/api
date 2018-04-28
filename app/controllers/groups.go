package controllers

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"aahframework.org/aah.v0"
	"github.com/keiwi/api/app/models"
	"github.com/keiwi/utils"
	storageModel "github.com/keiwi/utils/models"
	utilNats "github.com/keiwi/utils/nats"
	"github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2/bson"
)

type GroupsController struct {
	*aah.Context
}

func (a *GroupsController) RenameGroup(rename models.GroupRename) {
	if rename.NewName == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Please provide a new name for the group"})
		return
	}
	if rename.OldName == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Please provide the name of the group you want to rename"})
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

	existsOptions := utils.HasOptions{
		Filter: utils.Filter{"name": rename.NewName},
	}

	existsData, err := bson.MarshalJSON(existsOptions)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	exists, err := utilNats.HasGroup(n, existsData)
	if err != nil || exists {
		a.Reply().BadRequest().JSON(models.Response{Message: "There is already an existing group with this name"})
		return
	}

	update := utils.UpdateOptions{
		Filter: utils.Filter{"name": rename.OldName},
		Updates: utils.Updates{
			"name":       rename.NewName,
			"updated_at": time.Now(),
		},
	}

	data, err := bson.MarshalJSON(update)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	err = utilNats.UpdateGroup(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: fmt.Sprintf("Renamed %d group instances in the database", -1), Data: -1})
}

// CreateGroup - Handler for creating a new client
func (a *GroupsController) CreateGroup(create models.GroupCreate) {
	if create.CommandID == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Command ID is missing"})
		return
	}
	if create.GroupName == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Group name is missing"})
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

	findGroup := utils.FindOptions{
		Filter: utils.Filter{"name": create.GroupName},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	findData, err := bson.MarshalJSON(findGroup)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	groups, err := utilNats.FindGroup(n, findData)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	var group storageModel.Group

	if len(groups) <= 0 {
		group = storageModel.Group{
			Name: create.GroupName,
			Commands: []storageModel.GroupCommand{
				{
					ID:        bson.NewObjectId(),
					CommandID: bson.ObjectIdHex(create.CommandID),
				},
			},
		}

		data, err := bson.MarshalJSON(group)
		if err != nil {
			a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
			return
		}

		err = utilNats.CreateGroup(n, data)
		if err != nil {
			a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
			return
		}
	} else {
		group = groups[0]

		group.Commands = append(group.Commands, storageModel.GroupCommand{
			ID:        bson.NewObjectId(),
			CommandID: bson.ObjectIdHex(create.CommandID),
		})

		update := utils.UpdateOptions{
			Filter:  utils.Filter{"name": group.Name},
			Updates: utils.Updates{"$set": bson.M{"commands": group.Commands}},
		}

		data, err := bson.MarshalJSON(update)
		if err != nil {
			a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
			return
		}

		err = utilNats.UpdateGroup(n, data)
		if err != nil {
			a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
			return
		}
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully created added the command to the group", Data: group})
}

// EditGroup modifies an existing client in the database
func (a *GroupsController) EditGroup(edit models.EditRequest) {
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

	// retrieve existing client
	find := utils.FindOptions{
		Filter: utils.Filter{"commands.id": bson.ObjectIdHex(edit.ID)},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	data, err := bson.MarshalJSON(find)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// check if client actually exists
	groups, err := utilNats.FindGroup(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}
	if len(groups) <= 0 {
		a.Reply().BadRequest().JSON(models.Response{Message: "Can't find a client with this ID"})
		return
	}

	group := groups[0]

	// start parsing the update
	updates := bson.M{}
	switch strings.ToLower(edit.Option) {
	case "command_id", "commandid":
		v, ok := edit.Value.(string)
		if !ok {
			a.Reply().BadRequest().JSON(models.Response{Message: "Value is not a string"})
			return
		}

		for i, c := range group.Commands {
			if c.ID == bson.ObjectIdHex(edit.ID) {
				group.Commands[i].CommandID = bson.ObjectIdHex(v)
				updates["commands.$.command_id"] = bson.ObjectIdHex(v)
				break
			}
		}
	case "next_check", "nextcheck":
		next, err := convertToInt(edit.Value)
		if err != nil {
			a.Reply().BadRequest().JSON(models.Response{Message: "Value is not a number"})
			return
		}
		if next > 2147483647 || next < 0 {
			a.Reply().BadRequest().JSON(models.Response{Message: "Value is not a valid number"})
			return
		}

		for i, c := range group.Commands {
			if c.ID == bson.ObjectIdHex(edit.ID) {
				group.Commands[i].NextCheck = int(next)
				updates["commands.$.next_check"] = int(next)
				break
			}
		}
	case "stop_error", "stoperror":
		stop, ok := edit.Value.(bool)
		if !ok {
			a.Reply().BadRequest().JSON(models.Response{Message: "Value is not a boolean"})
			return
		}

		for i, c := range group.Commands {
			if c.ID == bson.ObjectIdHex(edit.ID) {
				group.Commands[i].StopError = stop
				updates["commands.$.stop_error"] = stop
				break
			}
		}
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

	err = utilNats.UpdateClient(n, updateData)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// if everything went well, respond with success
	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully saved the changes for the group", Data: group})
}

// DeleteGroup deletes a specific client from the database
func (a *GroupsController) DeleteGroup(delete models.GroupID) {
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

	err = utilNats.DeleteGroup(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully deleted the client"})
}

// DeleteGroupWithName deletes a specific client from the database
func (a *GroupsController) DeleteGroupWithName(delete models.GroupName) {
	if delete.Name == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Name is missing"})
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

	del := utils.DeleteOptions{
		Filter: utils.Filter{"name": delete.Name},
	}

	data, err := bson.MarshalJSON(del)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	err = utilNats.DeleteGroup(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully deleted the client"})
}

// GetGroups returns an array of all the clients in the database
func (a *GroupsController) GetGroups() {
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

	groups, err := utilNats.FindGroup(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found all groups in database", Data: groups})
}

// ExistsGroup returns an array of all the clients in the database
func (a *GroupsController) ExistsGroup(group models.GroupName) {
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

	hasOptions := utils.HasOptions{
		Filter: utils.Filter{"name": group.Name},
	}

	data, err := bson.MarshalJSON(hasOptions)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	has, err := utilNats.HasGroup(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully retrieved data", Data: has})
}

func convertToInt(i interface{}) (int64, error) {
	switch ci := i.(type) {
	case int64:
		return ci, nil
	case float64:
		return int64(ci), nil
	case string:
		pi, err := strconv.ParseInt(ci, 10, 64)
		if err != nil {
			return 0, err
		}
		return pi, nil
	default:
		return 0, errors.New("invalid number type")
	}
}
