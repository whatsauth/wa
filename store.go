package wa

import (
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetWaClient(phonenumber string, client []*WaClient, mongoconn *mongo.Database, container *sqlstore.Container) (waclient *WaClient, err error) {
	id := WithPhoneNumber(phonenumber, client)
	if id >= 0 {
		waclient = client[id]
	} else {
		waclient, err = CreateClientfromContainer(phonenumber, mongoconn, container)
		client = append(client, waclient)
	}
	return
}

func WithPhoneNumber(phonenumber string, clients []*WaClient) int {
	for i, client := range clients {
		if client.PhoneNumber == phonenumber {
			return i
		}
	}
	return -1
}
