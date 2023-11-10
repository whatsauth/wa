package wa

import (
	"fmt"

	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetWaClient(phonenumber string, client []*WaClient, mongoconn *mongo.Database, container *sqlstore.Container) (waclient WaClient, err error) {
	id := WithPhoneNumber(phonenumber, client)
	fmt.Println("id array:", id)
	if id >= 0 {
		fmt.Println("ada di array:", id)
		waclient = *client[id]
	} else {
		fmt.Println("masuk ke container whatsmeow")
		waclient, err = ClientDB(phonenumber, mongoconn, container)
		client = append(client, &waclient)
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
