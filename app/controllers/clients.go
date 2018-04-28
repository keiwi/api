package controllers

import (
	"strings"

	"aahframework.org/aah.v0"
	"github.com/apex/log"
	"github.com/keiwi/api/app/models"
	"github.com/keiwi/utils"
	storageModel "github.com/keiwi/utils/models"
	utilNats "github.com/keiwi/utils/nats"
	"github.com/nats-io/go-nats"
	"gopkg.in/mgo.v2/bson"
)

// ClientsController struct application controller
type ClientsController struct {
	*aah.Context
}

// CreateClient - Handler for creating a new client
func (a *ClientsController) CreateClient(create models.ClientCreate) {
	// Check if IP and name has been provided in the request
	if create.IP == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "IP is missing"})
		return
	}
	if create.Name == "" {
		a.Reply().BadRequest().JSON(models.Response{Message: "Name is missing"})
		return
	}

	n := models.Conn
	if n == nil {
		log.Error("nats is not initialized")
		a.Reply().InternalServerError().JSON(models.Response{Message: "Internal error"})
		return
	}

	// Initialize data for creating a new client
	client := storageModel.Client{
		Name: create.Name,
		IP:   create.IP,
	}

	// Marshal the data
	data, err := bson.MarshalJSON(client)
	if err != nil {
		log.Debugf("error marshaling data: %v", err)
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// Send the data to nats
	err = utilNats.CreateClient(n, data)
	if err != nil {
		log.Debugf("error creating the client: %v", err)
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully created the client", Data: client})
}

// DeleteClient deletes a specific client from the database
func (a *ClientsController) DeleteClient(delete models.ClientID) {
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

	err = utilNats.DeleteClient(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully deleted the client"})
}

// GetClients returns an array of all the clients in the database
func (a *ClientsController) GetClients() {
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

	clients, err := utilNats.FindClient(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found all clients in database", Data: clients})
}

// GetClientWithID returns a client if ID exists
func (a *ClientsController) GetClientWithID(client models.ClientID) {
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
		Filter: utils.Filter{"_id": bson.ObjectIdHex(client.ID)},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	data, err := bson.MarshalJSON(find)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	clients, err := utilNats.FindClient(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}
	if len(clients) <= 0 {
		a.Reply().BadRequest().JSON(models.Response{Message: "Could not find any clients"})
		return
	}

	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully found the client", Data: clients[0]})
}

// EditClient modifies an existing client in the database
func (a *ClientsController) EditClient(edit models.EditRequest) {
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
		Filter: utils.Filter{"_id": bson.ObjectIdHex(edit.ID)},
		Sort:   utils.Sort{"-created_at"},
		Limit:  1,
	}

	data, err := bson.MarshalJSON(find)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}

	// check if client actually exists
	clients, err := utilNats.FindClient(n, data)
	if err != nil {
		a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
		return
	}
	if len(clients) <= 0 {
		a.Reply().BadRequest().JSON(models.Response{Message: "Can't find a client with this ID"})
		return
	}

	client := clients[0]
	v, ok := edit.Value.(string)
	if !ok {
		a.Reply().BadRequest().JSON(models.Response{Message: "Value is not a string"})
		return
	}

	// start parsing the update
	updates := bson.M{}
	switch strings.ToLower(edit.Option) {
	case "name", "namn":
		updates["name"] = v
		client.Name = v
	case "ip":
		updates["ip"] = v
		client.IP = v
	case "group", "groups":
		// seperate which groups has been added and which groups has been deleted
		del, add := seperateGroups(objectIDArrayToString(client.GroupIDs), v)

		// loop through all groups that has been deleted and remove from the client instance
		for _, d := range del {
			if d == "" {
				continue
			}
			for i, g := range client.GroupIDs {
				id := g.Hex()
				if d == id {
					client.GroupIDs = append(client.GroupIDs[:i], client.GroupIDs[i+1:]...)
					break
				}
			}
		}

		// loop through all groups that has been added and add them to the client
		for _, ad := range add {
			if ad == "" {
				continue
			}

			// check if the added group is a real group in the database
			hasGroup := utils.HasOptions{
				Filter: utils.Filter{"_id": bson.ObjectIdHex(ad)},
			}

			hasData, err := bson.MarshalJSON(hasGroup)
			if err != nil {
				a.Reply().InternalServerError().JSON(models.Response{Message: "internal error"})
				return
			}

			// if the group exists add it to the client, otherwise error
			has, err := utilNats.HasGroup(n, hasData)
			if err != nil || !has {
				log.WithError(err).Error("error finding a group")
				a.Reply().BadRequest().JSON(models.Response{Message: "Can't find a group with the id " + ad})
				return
			}
			client.GroupIDs = append(client.GroupIDs, bson.ObjectIdHex(ad))
		}
		updates["group_ids"] = client.GroupIDs
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
	a.Reply().Ok().JSON(models.Response{Success: true, Message: "Successfully saved the changes for the client", Data: client})
}

func objectIDArrayToString(list []bson.ObjectId) string {
	length := len(list)
	out := ""
	for i, v := range list {
		if i >= length-1 {
			out += v.Hex()
		} else {
			out += v.Hex() + ", "
		}
	}
	return out
}

func seperateGroups(old, new string) (deleted, added []string) {
	splitOld := strings.Split(old, ",")
	splitNew := strings.Split(new, ",")

	deleted = findDifference(splitOld, splitNew)
	added = findDifference(splitNew, splitOld)
	return
}

func findDifference(a1, a2 []string) []string {
	var out []string
	for _, i1 := range a1 {
		m := false
		for _, i2 := range a2 {
			if i1 == i2 {
				m = true
				break
			}
		}
		if m {
			continue
		}
		out = append(out, i1)
	}
	return out
}
