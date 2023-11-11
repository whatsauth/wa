package wa

import (
	"context"
	"fmt"

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

func ResetDeviceStore(mongoconn *mongo.Database, client *WaClient, container *sqlstore.Container) (err error) {
	if client.WAClient.Store.ID != nil {
		var id uint16
		id, err = GetDeviceIDFromContainer(client.PhoneNumber, container)
		filter := bson.M{"phonenumber": client.PhoneNumber}
		var user User
		user, err = atdb.GetOneLatestDoc[User](mongoconn, "user", filter)
		user.DeviceID = id
		client.WAClient.Store.ID.Device = id
		atdb.ReplaceOneDoc(mongoconn, "user", bson.M{"phonenumber": user.PhoneNumber}, user)

	}
	return
}

func CreateClientfromContainer(phonenumber string, mongoconn *mongo.Database, container *sqlstore.Container) (client *WaClient, err error) {
	var user User
	user, err = atdb.GetOneLatestDoc[User](mongoconn, "user", bson.M{"phonenumber": phonenumber})
	if err != nil {
		user.PhoneNumber = phonenumber
	}
	var deviceStore *store.Device
	if user.DeviceID == 0 {
		var deviceid uint16
		deviceid, err = GetDeviceIDFromContainer(phonenumber, container)
		deviceStore, err = container.GetDevice(types.JID{User: user.PhoneNumber, Device: deviceid, Server: "s.whatsapp.net"})
	} else {
		deviceStore, err = container.GetDevice(types.JID{User: user.PhoneNumber, Device: user.DeviceID, Server: "s.whatsapp.net"})
	}
	if deviceStore == nil {
		deviceStore = container.NewDevice()
	}
	var wc WaClient
	wc.PhoneNumber = phonenumber
	wc.WAClient = whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "ERROR", true))
	wc.Mongoconn = mongoconn
	wc.register()
	if wc.WAClient.Store.ID != nil { //tanda belum terkoneksi
		user.DeviceID = deviceStore.ID.Device
	}
	atdb.ReplaceOneDoc(mongoconn, "user", bson.M{"phonenumber": phonenumber}, user)
	client = &wc
	return

}

func GetDeviceIDFromContainer(phonenumber string, container *sqlstore.Container) (deviceid uint16, err error) {
	deviceStores, err := container.GetAllDevices()
	fmt.Println(err)
	for _, dv := range deviceStores {
		if dv.ID.User == phonenumber {
			deviceid = dv.ID.Device
		}
	}
	return
}

func GetDeviceStoreFromContainer(phonenumber string, container *sqlstore.Container) (device *store.Device, err error) {
	deviceStores, err := container.GetAllDevices()
	fmt.Println(err)
	for _, dv := range deviceStores {
		if dv.ID.User == phonenumber {
			device = dv
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

func ConnectClient(client *whatsmeow.Client) {
	if !client.IsConnected() {
		client.Connect()
	}
}

func ConnectAllClient(mongoconn *mongo.Database, container *sqlstore.Container) (clients []*WaClient, err error) {
	deviceStores, err := container.GetAllDevices()
	for _, deviceStore := range deviceStores {
		client := whatsmeow.NewClient(deviceStore, waLog.Stdout("Client", "ERROR", true))
		filter := bson.M{"phonenumber": deviceStore.ID.User}
		user, err := atdb.GetOneLatestDoc[User](mongoconn, "user", filter)
		if (client.Store.ID != nil) && (err == nil) {
			var mycli WaClient
			mycli.WAClient = client
			mycli.PhoneNumber = deviceStore.ID.User
			mycli.Mongoconn = mongoconn
			mycli.register()
			client.Connect()
			clients = append(clients, &mycli)
			user.DeviceID = deviceStore.ID.Device
			atdb.ReplaceOneDoc(mongoconn, "user", bson.M{"phonenumber": user.PhoneNumber}, user)
		}

	}

	return

}
