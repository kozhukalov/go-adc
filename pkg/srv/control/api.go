/*
 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

     https://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
*/

package control

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"jinr.ru/greenlab/go-adc/pkg/config"
	"jinr.ru/greenlab/go-adc/pkg/layers"
	"jinr.ru/greenlab/go-adc/pkg/log"
	"jinr.ru/greenlab/go-adc/pkg/srv"
	"jinr.ru/greenlab/go-adc/pkg/srv/control/ifc"
	"net/http"
	"strconv"
)

const (
	ApiPort = 8000
)

// RegHex ...
type RegHex struct {
	Addr string // hexadecimal
	Value string // hexadecimal
}

type ApiServer struct {
	context.Context
	*config.Config
	*mux.Router
	ctrl ifc.ControlServer
}

var _ ifc.ApiServer = &ApiServer{}

func NewApiServer(ctx context.Context, cfg *config.Config, ctrl ifc.ControlServer) (ifc.ApiServer, error) {
	log.Debug("Initializing API server with address: %s port: %d", cfg.IP, ApiPort)

	s := &ApiServer{
		Context: ctx,
		Config: cfg,
		ctrl: ctrl,
	}
	return s, nil
}

func (s *ApiServer) regReadHex(addr uint16, device string) (*RegHex, error) {
	d, err := s.ctrl.GetDeviceByName(device)
	if err != nil {
		return nil, err
	}
	reg, err := d.RegRead(addr)
	if err != nil {
		return nil, err
	}
	hexAddr, hexValue := reg.Hex()
	return &RegHex{
		Addr:  hexAddr,
		Value: hexValue,
	}, nil
}

func (s *ApiServer) regReadAllHex(device string) ([]*RegHex, error) {
	d, err := s.ctrl.GetDeviceByName(device)
	if err != nil {
		return nil, err
	}
	regs, err := d.RegReadAll()
	if err != nil {
		return nil, err
	}
	var regsHex []*RegHex
	for _, reg := range regs {
		hexAddr, hexValue := reg.Hex()
		regsHex = append(regsHex, &RegHex{ Addr: hexAddr, Value: hexValue })
	}
	return regsHex, nil
}

// Start
func (s *ApiServer) Run() error {
	log.Debug("Starting API server: address: %s port: %d", s.Config.IP, ApiPort)
	s.configureRouter()
	httpServer := &http.Server{
		Handler: s.Router,
		Addr:    fmt.Sprintf("%s:%d", s.Config.IP, ApiPort),
	}
	return httpServer.ListenAndServe()
}

func (s *ApiServer) configureRouter() {
	s.Router = mux.NewRouter()
	subRouter := s.Router.PathPrefix("/api").Subrouter()
	subRouter.HandleFunc("/reg/r/{device}/{addr:0x[0-9abcdef]{4}}", s.handleRegRead()).Methods("GET")
	subRouter.HandleFunc("/reg/r/{device}", s.handleRegReadAll()).Methods("GET")
	subRouter.HandleFunc("/reg/w/{device}", s.handleRegWrite()).Methods("POST")
	subRouter.HandleFunc("/mstream/{action:start|stop}/{device}", s.handleMStreamAction()).Methods("GET")
	subRouter.HandleFunc("/mstream/{action:start|stop}", s.handleMStreamActionAll()).Methods("GET")
}

func (s *ApiServer) handleRegRead() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		log.Debug("Handling reg read request: device: %s, addr: %s", vars["device"], vars["addr"])

		addr, err := strconv.ParseUint(vars["addr"], 0, 16)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		regHex, err := s.regReadHex(uint16(addr), vars["device"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(regHex)
	}
}

func (s *ApiServer) handleRegReadAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		log.Debug("Handling reg read all request: device: %s", vars["device"])

		regsHex, err := s.regReadAllHex(vars["device"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(regsHex)
	}
}

func (s *ApiServer) handleRegWrite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)

		regHex := &RegHex{}
		err := json.NewDecoder(r.Body).Decode(regHex)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		log.Debug("Handling reg write request: device: %s addr: %s value: %s",
			vars["device"], regHex.Addr, regHex.Value)

		reg, err := layers.NewRegFromHex(regHex.Addr, regHex.Value)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		device, err := s.ctrl.GetDeviceByName(vars["device"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		err = device.RegWrite(reg)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
	}
}

func (s *ApiServer) handleMStreamAction() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		log.Debug("Handling MStream action request: device: %s action: %s", vars["device"], vars["action"])
		device, err := s.ctrl.GetDeviceByName(vars["device"])
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		switch vars["action"] {
		case "start":
			err = device.MStreamStart()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
		case "stop":
			err := device.MStreamStop()
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadGateway)
				return
			}
		default:
			err := srv.ErrUnknownOperation{
				What: "Wrong MStream action. Must be one of start/stop",
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}

func (s *ApiServer) handleMStreamActionAll() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		log.Debug("Handling MStream action request for all devices: action: %s", vars["action"])
		switch vars["action"] {
		case "start":
			for _, d := range s.ctrl.GetAllDevices() {
				err := d.MStreamStart()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadGateway)
					return
				}
			}
		case "stop":
			for _, d := range s.ctrl.GetAllDevices() {
				err := d.MStreamStop()
				if err != nil {
					http.Error(w, err.Error(), http.StatusBadGateway)
					return
				}
			}
		default:
			err := srv.ErrUnknownOperation{
				What: "Wrong MStream action. Must be one of start/stop",
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
}
