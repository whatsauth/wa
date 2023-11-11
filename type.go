package wa

import (
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/types"
	"go.mongodb.org/mongo-driver/mongo"
)

type TextMessage struct {
	To       string `json:"to"`
	IsGroup  bool   `json:"isgroup,omitempty"`
	Messages string `json:"messages"`
}

type QRStatus struct {
	PhoneNumber string `json:"phonenumber"`
	Status      bool   `json:"status"`
	Code        string `json:"code"`
	Message     string `json:"message"`
}

type WaClient struct {
	PhoneNumber    string
	WAClient       *whatsmeow.Client
	eventHandlerID uint32
	Mongoconn      *mongo.Database
	ID             *types.JID
}

type WebHook struct {
	URL    string `bson:"url" json:"url"`
	Secret string `bson:"secret" json:"secret"`
}

type User struct {
	PhoneNumber string  `bson:"phonenumber" json:"phonenumber"`
	WebHook     WebHook `bson:"webhook" json:"webhook"`
	Token       string  `bson:"token" json:"token"`
}
