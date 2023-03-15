package handler

import (
	"context"
	"fmt"
	"github.com/apache/thrift/lib/go/thrift"
	"isis-total-order-multicast/common"
	"isis-total-order-multicast/gen-go/multicast/rpc"
	"isis-total-order-multicast/isis"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

type Handler struct {
	isis *isis.Isis
}

func (h *Handler) Echo(ctx context.Context, req *rpc.EchoRequest) (_r *rpc.EchoResponse, _err error) {
	fmt.Printf("message from client: %v\n", req.GetMsg())

	resp := &rpc.EchoResponse{
		Msg: "Yeah.",
	}

	return resp, nil
}

func (h *Handler) SendData(ctx context.Context, req *rpc.SendDataRequest) (_r *rpc.SendDataResponse, _err error) {
	resp := &rpc.SendDataResponse{}
	msg := req.Msg
	seq := h.isis.Receive(msg, req.MsgId, req.Sender)
	resp.MsgId = req.MsgId
	resp.ProposedSeq = seq
	return resp, nil
}

func (h *Handler) SendSeq(ctx context.Context, req *rpc.SendSeqRequest) (_r *rpc.SendSeqResponse, _err error) {
	resp := &rpc.SendSeqResponse{}
	if h.isis.IsDelivered(req.MsgId, req.Sender) {
		return resp, nil
	}
	h.isis.Deliver(req.MsgId, req.AgreedSeq, req.Sender, req.DecisionMaker)
	resp.DeliverTime = time.Now().Format("2006-01-02 15:04:05")
	go h.isis.BMulticast(ctx, req.MsgId, req.AgreedSeq, req.Sender, req.DecisionMaker)
	return resp, nil
}

func NewMulticastServiceSvr(nodeInfo common.NodeInfo, isis *isis.Isis) *thrift.TSimpleServer {
	port := strconv.Itoa(nodeInfo.Port)
	transport, err := thrift.NewTServerSocket(":" + port)
	if err != nil {
		panic(err)
	}

	handler := &Handler{
		isis: isis,
	}
	processor := rpc.NewMulticastServiceProcessor(handler)

	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTCompactProtocolFactory()
	server := thrift.NewTSimpleServer4(
		processor,
		transport,
		transportFactory,
		protocolFactory,
	)
	return server
}

func NewMulticastServiceCli(nodeInfo common.NodeInfo) (*rpc.MulticastServiceClient, *thrift.TSocket) {
	ip, err := net.LookupHost(nodeInfo.Addr)
	if err != nil {
		log.Println("DNS analysing for node addr went wrong")
		os.Exit(1)
	}
	port := strconv.Itoa(nodeInfo.Port)
	transport, err := thrift.NewTSocket(net.JoinHostPort(ip[0], port))
	if err != nil {
		panic(err)
	}

	transportFactory := thrift.NewTTransportFactory()
	protocolFactory := thrift.NewTCompactProtocolFactory()

	useTransport, err := transportFactory.GetTransport(transport)
	client := rpc.NewMulticastServiceClientFactory(useTransport, protocolFactory)
	return client, transport
}
