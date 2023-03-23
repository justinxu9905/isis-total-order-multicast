package server

import (
	"context"
	"fmt"
	"isis-total-order-multicast/config"
	"isis-total-order-multicast/handler"
	"isis-total-order-multicast/isis"
	log "isis-total-order-multicast/logger"
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

	// add himself into peer list
	s.AddPeer(s.nodeConfig)

	// run peer change task
	s.is.RunDelTask()
}

func (s *Server) AddPeer(nodeConfig config.NodeConfig) error {
	cli, swt := handler.NewMulticastServiceCli(nodeConfig)
	if swt.IsOpen() {
		log.Infof("Switch with %v is already open", nodeConfig.Id)
		return nil
	}
	retry := 5
	for retry > 0 {
		if err := swt.Open(); err != nil {
			retry--
		} else {
			break
		}
		if retry == 0 {
			return fmt.Errorf("Failed to open switch with %v", nodeConfig.Id)
		}
		time.Sleep(time.Second)
	}
	if swt.IsOpen() {
		s.is.AddPeer(nodeConfig.Id, cli)
	}
	return nil
}

func (s *Server) SendMsg(msg string) {
	s.is.TOMulticast(context.Background(), msg)
}
