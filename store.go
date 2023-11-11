package wa

import (
	"github.com/aiteung/atdb"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetWaClient(phonenumber string, client []*WaClient, mongoconn *mongo.Database, container *sqlstore.Container) (waclient *WaClient, err error) {
	id, err := WithPhoneNumber(phonenumber, client, mongoconn)
	if id >= 0 {
		waclient = client[id]
	} else {
		client, err = ConnectAllClient(mongoconn, container)
		id, err = WithPhoneNumber(phonenumber, client, mongoconn)
		waclient = client[id]
		//waclient, err = CreateClientfromContainer(phonenumber, mongoconn, container)
		//client = append(client, waclient)
	}
	return
}

func WithPhoneNumber(phonenumber string, clients []*WaClient, mongoconn *mongo.Database) (idx int, err error) {
	user, err := atdb.GetOneLatestDoc[User](mongoconn, "user", bson.M{"phonenumber": phonenumber})
	idx = -1
	for i, client := range clients {
		if client.WAClient.Store.ID != nil {
			if (client.WAClient.Store.ID.User == phonenumber) && (client.WAClient.Store.ID.Device == user.DeviceID) {
				idx = i
			}
		}
	}
	return
}
