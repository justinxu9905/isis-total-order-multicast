package server

import (
	"context"
	"fmt"
	"isis-total-order-multicast/config"
	"isis-total-order-multicast/handler"
	"isis-total-order-multicast/isis"
	"time"
)

type Server struct {
	nodeConfig config.NodeConfig
	is         *isis.Isis
}

func NewServer(nodeConfig config.NodeConfig) *Server {
	return &Server{
		nodeConfig: nodeConfig,
		is:         isis.NewIsis(nodeConfig),
	}
}

func (s *Server) Start() {
	// run isis server
	svr := handler.NewMulticastServiceSvr(s.nodeConfig, s.is)
	go func() {
		if err := svr.Serve(); err != nil {
			panic(err)
		}
	}()

	// run peer change task
	go s.is.RunDelTask()
}

func (s *Server) AddPeer(nodeConfig config.NodeConfig) {
	cli, swt := handler.NewMulticastServiceCli(nodeConfig)
	if swt.IsOpen() {
		fmt.Println("already open")
		return
	}
	retry := 5
	for retry > 0 {
		if err := swt.Open(); err != nil {
			retry--
		} else {
			break
		}
		if retry == 0 {
			fmt.Println("failed to open")
			return
		}
		time.Sleep(time.Second)
	}
	if swt.IsOpen() {
		s.is.AddPeer(nodeConfig.Id, cli)
	}
}

func (s *Server) SendMsg(msg string) {
	s.is.TOMulticast(context.Background(), msg)
}
