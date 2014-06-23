package localapi

import (
	"fmt"
	"github.com/gorilla/rpc"
	"github.com/gorilla/rpc/json"
	"net"
	"net/http"
	"strongmessage/api"
	"strongmessage/local/localdb"
	"strongmessage/objects"
)

type StrongService struct {
	Config *api.ApiConfig
	Log    chan string
}

type NilParam struct{}

func (s *StrongService) Version(r *http.Request, args *NilParam, reply *objects.Version) error {
	*reply = *s.Config.LocalVersion
	return nil
}

func Initialize(log chan string, config *api.ApiConfig, port uint16) error {

	e := localdb.Initialize(log, config.LocalDB)

	if e != nil {
		return e
	}

	s := rpc.NewServer()
	s.RegisterCodec(json.NewCodec(), "application/json")
	service := new(StrongService)
	service.Config = config
	service.Log = log
	s.RegisterService(service, "StrongService")

	http.Handle("/", s)

	l, e := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if e != nil {
		log <- fmt.Sprintf("RPC Listen Error: %s", e)
		return e
	}

	go http.Serve(l, nil)

	log <- fmt.Sprintf("Started RPC Server on: %s", fmt.Sprintf(":%d", port))
	return nil
}
