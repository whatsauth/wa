package wa

import (
	"fmt"

	"go.mongodb.org/mongo-driver/mongo"
)

func GetWaClient(phonenumber string, client []*WaClient, mongoconn *mongo.Database) (waclient WaClient) {
	id := WithPhoneNumber(phonenumber, client)
	fmt.Println("id array:", id)
	if id >= 0 {
		fmt.Println("ada di array:", id)
		waclient = *client[id]
	} else {
		fmt.Println("masuk ke container whatsmeow")
		waclient = ClientDB(phonenumber, mongoconn)
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
