// Code generated by GoVPP's binapi-generator. DO NOT EDIT.

package vpe

import (
	"encoding/json"
	"net/http"
)

func HTTPHandler(rpc RPCService) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/show_version", func(w http.ResponseWriter, req *http.Request) {
		var request = new(ShowVersion)
		reply, err := rpc.ShowVersion(req.Context(), request)
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
	mux.HandleFunc("/show_vpe_system_time", func(w http.ResponseWriter, req *http.Request) {
		var request = new(ShowVpeSystemTime)
		reply, err := rpc.ShowVpeSystemTime(req.Context(), request)
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
