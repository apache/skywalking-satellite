package http

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/apache/skywalking-satellite/internal/pkg/log"
	http_server "github.com/apache/skywalking-satellite/plugins/server/http"
	"github.com/apache/skywalking-satellite/protocol/gen-codes/satellite/protocol"
	logging "skywalking/network/logging/v3"
)

const (
	Name      = "http-log-receiver"
	eventName = "http-log-event"
	timeout   = 5 * time.Second
)

type Receiver struct {
	Server        *http_server.Server
	OutputChannel chan *protocol.Event
}

func (r *Receiver) Name() string {
	return Name
}

func (r *Receiver) Description() string {
	return "This is a receiver for SkyWalking http logging format, " +
		"which is defined at https://github.com/apache/skywalking-data-collect-protocol/blob/master/logging/Logging.proto."
}

func (r *Receiver) DefaultConfig() string {
	return ""
}

func (r *Receiver) RegisterHandler(server interface{}) {
	r.Server = server.(*http_server.Server)
	r.Server.Server.Handle(r.Server.Uri, httpHandler(r))
}

func httpHandler(r *Receiver) http.Handler {
	h := http.HandlerFunc(func(rsp http.ResponseWriter, req *http.Request) {
		b, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Logger.Errorf("get http body error: %v", err)
			http.Error(rsp, err.Error(), http.StatusBadRequest)
			return
		}
		var data *logging.LogData
		err = json.Unmarshal(b, &data)
		if err != nil {
			http.Error(rsp, err.Error(), http.StatusInternalServerError)
			return
		}
		e := &protocol.Event{
			Name:      eventName,
			Timestamp: time.Now().UnixNano() / 1e6,
			Meta:      nil,
			Type:      protocol.EventType_Logging,
			Remote:    true,
			Data: &protocol.Event_Log{
				Log: data,
			},
		}
		r.OutputChannel <- e
	})
	return http.TimeoutHandler(h, timeout, fmt.Sprintf("Exceeded configured timeout of %v \n", timeout))
}

func (r *Receiver) Channel() <-chan *protocol.Event {
	return r.OutputChannel
}
