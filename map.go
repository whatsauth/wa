package wa

import (
	"fmt"
	"github.com/puzpuzpuz/xsync/v3"
)

type MapClient struct {
	xsync.MapOf[string, *WaClient]
}

func (m *MapClient) NewMap() MapClient {
	return MapClient{}
}

func (m *MapClient) GetClient(id string) (client *WaClient, ok bool) {
	client, ok = m.Load(id)
	return
}

func (m *MapClient) StoreClient(id string, client *WaClient) {
	m.Store(id, client)
}

func (m *MapClient) StoreOnlineClient(id string, client *WaClient) (ok bool) {
	if client.WAClient.IsConnected() {
		goto STORECLIENT
	}
	if err := client.WAClient.Connect(); err != nil {
		return
	}

STORECLIENT:
	m.Store(id, client)
	return
}

func (m *MapClient) CheckClientOnline(id string) (ok bool) {
	client, ok := m.Load(id)
	if !ok {
		return
	}

	ok = client.WAClient.IsConnected()
	return
}

func (m *MapClient) GetAllClient() (listCli []*WaClient) {

	m.Range(func(k string, v *WaClient) bool {
		listCli = append(listCli, v)
		return true
	})
	return

}

func (m *MapClient) StatusAllClient() (res map[string]bool) {
	res = make(map[string]bool, m.Size())

	m.Range(func(k string, v *WaClient) bool {
		res[k] = v.WAClient.IsConnected()
		return true
	})

	return
}

func (m *MapClient) OfflineClient() (res []string) {

	m.Range(func(k string, v *WaClient) (ok bool) {
		ok = true

		if !v.WAClient.IsConnected() {
			res = append(res, k)
		}

		return
	})

	return
}

func (m *MapClient) StoreAllClient(listClient []*WaClient) (ok bool) {
	if len(listClient) < 1 {
		return
	}

	// (client.WAClient.Store.ID.User == phonenumber) && (client.WAClient.Store.ID.Device == user.DeviceID)
	for _, v := range listClient {
		phoneNum := v.WAClient.Store.ID
		devId := v.WAClient.Store.ID.Device
		m.Store(fmt.Sprintf("%s-%d", phoneNum, devId), v)
	}
	ok = true
	return
}

func (m *MapClient) SetOnlineClient(id string) (ok bool) {
	cli, ok := m.Load(id)
	if !ok {
		return
	}

	if cli.WAClient.IsConnected() {
		ok = true
		return
	}

	err := cli.WAClient.Connect()
	if err != nil {
		return
	}

	ok = true
	return
}