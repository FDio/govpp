// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package tcp

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
)

func RESTHandler(rpc RPCService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/tcp_configure_src_addresses", func(w http.ResponseWriter, req *http.Request) {
		var request = new(TCPConfigureSrcAddresses)
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			http.Error(w, "read body failed", http.StatusBadRequest)
			return
		}
		if err := json.Unmarshal(b, request); err != nil {
			http.Error(w, "unmarshal data failed", http.StatusBadRequest)
			return
		}
		reply, err := rpc.TCPConfigureSrcAddresses(req.Context(), request)
		if err != nil {
			http.Error(w, "request failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		rep, err := json.MarshalIndent(reply, "", "  ")
		if err != nil {
			http.Error(w, "marshal failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Write(rep)
	})
	return http.HandlerFunc(mux.ServeHTTP)
}