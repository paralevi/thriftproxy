package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"io/ioutil"
	"net/http"
)

type Admin struct {
	server   http.Server
	proxyMgr *ProxyMgr
}

type ProxyBackends struct {
	Proxies []struct {
		Name     string
		Backends []BackendInfo
	}
}

func NewAdmin(addr string, proxyMgr *ProxyMgr) *Admin {
	admin := &Admin{proxyMgr: proxyMgr}
	admin.server.Addr = addr
	router := mux.NewRouter()
	router.HandleFunc("/addbackend", admin.processAddBackend)
	router.HandleFunc("/removebackend", admin.processRemoveBackend)
	admin.server.Handler = router
	return admin
}

func (admin *Admin) Start() {
	go admin.server.ListenAndServe()
}

func (admin *Admin) processAddBackend(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	defer r.Body.Close()
	proxyBackends, err := admin.readProxyBackends(r)
	if err == nil {
		admin.addBackend(proxyBackends)
	}
}

func (admin *Admin) processRemoveBackend(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)

	defer r.Body.Close()
	proxyBackends, err := admin.readProxyBackends(r)
	if err == nil {
		admin.removeBackend(proxyBackends)
	}
}

func (admin *Admin) readProxyBackends(r *http.Request) (*ProxyBackends, error) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	info := ProxyBackends{}
	err = json.Unmarshal(b, &info)
	return &info, err

}

func (admin *Admin) processBackend(proxyBackends *ProxyBackends, proxyProcFunc func(proxy *Proxy, backend *BackendInfo)) {
	for _, proxyInfo := range proxyBackends.Proxies {
		proxy, err := admin.proxyMgr.GetProxy(proxyInfo.Name)
		if err == nil {
			for _, backend := range proxyInfo.Backends {
				proxyProcFunc(proxy, &backend)
			}
		}
	}

}

func (admin *Admin) addBackend(proxyBackends *ProxyBackends) {
	admin.processBackend(proxyBackends, func(proxy *Proxy, backend *BackendInfo) {
		proxy.AddBackend(backend.Addr, backend.Readiness)
	})
}

func (admin *Admin) removeBackend(proxyBackends *ProxyBackends) {
	admin.processBackend(proxyBackends, func(proxy *Proxy, backend *BackendInfo) {
		proxy.RemoveBackend(backend.Addr)
	})
}
