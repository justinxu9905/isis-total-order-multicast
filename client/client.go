package client

import (
	"bufio"
	"github.com/spf13/viper"
	"isis-total-order-multicast/config"
	log "isis-total-order-multicast/logger"
	"isis-total-order-multicast/server"
	"os"
	"regexp"
	"strings"
)

type Client struct {
	server *server.Server
}

func NewClient(server *server.Server) *Client {
	return &Client{
		server: server,
	}
}

func (c *Client) Start() {
	regex := regexp.MustCompile(" +")
	reader := bufio.NewReader(os.Stdin)
	for {
		cmdString, err := reader.ReadString('\n')
		if err != nil {
			log.Errorf(err.Error())
			continue
		}
		cmdString = strings.TrimSuffix(cmdString, "\n")
		cmdString = strings.Trim(cmdString, " ")
		if cmdString == "" {
			continue
		}
		cmd := regex.Split(cmdString, -1)
		if cmd[0] == "quit" || cmd[0] == "exit" {
			break
		} else if cmd[0] == "send" {
			if len(cmd) < 2 {
				log.Errorf("Command error: missing required argument [message]")
				continue
			}
			msg := strings.Join(cmd[1:], " ")
			c.server.SendMsg(msg)
			log.Infof("Send message %v successfully", msg)
		} else if cmd[0] == "add" {
			if len(cmd) != 2 {
				log.Errorf("Command error: missing required argument [peer id]")
				continue
			}
			peerId := cmd[1]
			err := c.server.AddPeer(config.NodeConfig{
				Id: peerId,
				Addr: viper.GetString(peerId+".ip"),
				Port: viper.GetInt(peerId+".port"),
				StorageFilename: viper.GetString(peerId+".filename"),
			})
			if err != nil {
				log.Errorf(err.Error())
				continue
			}
			log.Infof("Add peer %v successfully", peerId)
		} else if cmd[0] == "help" {
			log.Infof(`Supported commands:
			send [message]
			add [peer id]
			exit/quit`)
		} else {
			log.Errorf("Command error: invalid command")
		}
	}
}
