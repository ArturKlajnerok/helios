package main

import (
	"./storage"
	"fmt"
	"github.com/RangelReale/osin"
	"net/http"
)

var server *osin.Server

func tokenHandler(w http.ResponseWriter, r *http.Request) {
	resp := server.NewResponse()
	defer resp.Close()
	if ar := server.HandleAccessRequest(resp, r); ar != nil {
		switch ar.Type {
		case osin.PASSWORD:
			if ar.Username == "test" && ar.Password == "test" {
				ar.Authorized = true
			}
		}
		server.FinishAccessRequest(resp, r, ar)
	}
	if resp.IsError && resp.InternalError != nil {
		fmt.Printf("ERROR: %s\n", resp.InternalError)
	}
	osin.OutputJSON(resp, w, r)
}

func main() {
	cfg := osin.NewServerConfig()
	cfg.AllowedAccessTypes = osin.AllowedAccessType{osin.PASSWORD, osin.REFRESH_TOKEN}
	cfg.AllowGetAccessRequest = true
	cfg.AllowClientSecretInParams = true

	server = osin.NewServer(cfg, storage.NewRedisStorage("localhost:6379", "auth"))

	http.HandleFunc("/token", tokenHandler)
	http.ListenAndServe(":8080", nil)
}
