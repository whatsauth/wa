package wa

import (
	"context"

	"github.com/aiteung/atdb"
	"github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	waLog "go.mau.fi/whatsmeow/util/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func (mycli *WaClient) register() {
	mycli.eventHandlerID = mycli.WAClient.AddEventHandler(mycli.EventHandler)
}

func (mycli *WaClient) EventHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		go HandlingMessage(&v.Info, v.Message, mycli)
	}
}

func CreateContainerDB(pgstring string) (container *sqlstore.Container, err error) {
	dbLog := waLog.Stdout("Database", "ERROR", true)
	pgUrl, err := pq.ParseURL(pgstring)
	container, err = sqlstore.New("postgres", pgUrl, dbLog)
	return
}

func ResetDeviceStore(client *WaClient, container *sqlstore.Container) (err error) {
	err = container.DeleteDevice(client.WAClient.Store)
	deviceStore := container.NewDevice()
	client.WAClient = whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "ERROR", true))
	client.ID = deviceStore.ID
	return
}

func CreateClientfromContainer(phonenumber string, mongoconn *mongo.Database, container *sqlstore.Container) (client *WaClient, err error) {
	user, err := atdb.GetOneLatestDoc[User](mongoconn, "user", bson.M{"phonenumber": phonenumber})
	var deviceStore *store.Device
	if user.DeviceID == 0 {
		var deviceid uint16
		deviceid, err = GetDeviceIDFromContainer(phonenumber, mongoconn, container)
		deviceStore, err = container.GetDevice(types.JID{User: user.PhoneNumber, Device: deviceid, Server: "s.whatsapp.net"})
	} else {
		deviceStore, err = container.GetDevice(types.JID{User: user.PhoneNumber, Device: user.DeviceID, Server: "s.whatsapp.net"})
	}
	if deviceStore == nil {
		deviceStore = container.NewDevice()
	}
	client.PhoneNumber = user.PhoneNumber
	client.WAClient = whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "ERROR", true))
	client.Mongoconn = mongoconn
	client.ID = deviceStore.ID
	client.register()
	user.DeviceID = client.ID.Device
	atdb.ReplaceOneDoc(mongoconn, "user", bson.M{"phonenumber": phonenumber}, user)
	return

}

func GetDeviceIDFromContainer(phonenumber string, mongoconn *mongo.Database, container *sqlstore.Container) (deviceid uint16, err error) {
	deviceStores, err := container.GetAllDevices()
	for _, dv := range deviceStores {
		if dv.ID.User == phonenumber {
			deviceid = dv.ID.Device
		}
	}
	return
}

func QRConnect(client *WaClient, qr chan QRStatus) {
	if client.WAClient.Store.ID == nil {
		//client.PairPhone(PhoneNumber, true, whatsmeow.PairClientUnknown, "whatsauth.my.id")
		qrChan, _ := client.WAClient.GetQRChannel(context.Background())
		err := client.WAClient.Connect()
		if err != nil {
			panic(err)
		}
		// No ID stored, new login
		for evt := range qrChan {
			if evt.Event == "code" {
				qr <- QRStatus{client.PhoneNumber, true, evt.Code, evt.Event}
			} else {
				qr <- QRStatus{client.PhoneNumber, true, evt.Code, evt.Event}
			}
		}
	} else {
		message := "already login"
		err := client.WAClient.Connect()
		if err != nil {
			message = err.Error()
		}
		qr <- QRStatus{client.PhoneNumber, false, "", message}
	}

}

func PairConnect(client *WaClient, qr chan QRStatus) {
	if client.WAClient.Store.ID == nil {
		message := "Pair Code Device"
		err := client.WAClient.Connect()
		if err != nil {
			message = err.Error()
			qr <- QRStatus{client.PhoneNumber, false, "", message}
		}
		// No ID stored, new login
		code, err := client.WAClient.PairPhone(client.PhoneNumber, true, whatsmeow.PairClientUnknown, "Chrome (Mac OS)")
		if err != nil {
			message = err.Error()
			qr <- QRStatus{client.PhoneNumber, false, "", message}
		} else {
			qr <- QRStatus{client.PhoneNumber, true, code, message}
		}

	} else {
		message := "already login"
		if !client.WAClient.IsConnected() {
			message = "Melakukan Koneksi Ulang"
			err := client.WAClient.Connect()
			if err != nil {
				message = err.Error()
				qr <- QRStatus{client.PhoneNumber, false, "", message}
			}
		} else {
			qr <- QRStatus{client.PhoneNumber, false, "", message}
		}

	}

}

func ConnectAllClient(mongoconn *mongo.Database, container *sqlstore.Container) (clients []*WaClient, err error) {
	deviceStores, err := container.GetAllDevices()
	//deviceStore, err := container.GetDevice(jid)
	for _, deviceStore := range deviceStores {
		client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "ERROR", true))
		//client.AddEventHandler(EventHandler)
		filter := bson.M{"phonenumber": deviceStore.ID.User}
		user, err := atdb.GetOneLatestDoc[User](mongoconn, "user", filter)
		if (client.Store.ID != nil) && (err == nil) {
			var mycli WaClient
			mycli.WAClient = client
			mycli.PhoneNumber = deviceStore.ID.User
			mycli.Mongoconn = mongoconn
			mycli.ID = deviceStore.ID
			mycli.register()
			client.Connect()
			clients = append(clients, &mycli)
			user.DeviceID = mycli.ID.Device
			atdb.ReplaceOneDoc(mongoconn, "user", bson.M{"phonenumber": user.PhoneNumber}, user)
		}

	}

	return

}
