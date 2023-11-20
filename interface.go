package wa

type StoreClient interface {
	StoreClient(id string, client *WaClient)
	StoreOnlineClient(id string, client *WaClient) (ok bool)
}
