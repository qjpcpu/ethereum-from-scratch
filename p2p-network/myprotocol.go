package main

import (
	"crypto/ecdsa"
	"flag"
	"fmt"
	"github.com/ethereum/go-ethereum/crypto"
	//	ethlog "github.com/ethereum/go-ethereum/log"
	"github.com/ethereum/go-ethereum/p2p"
	"github.com/ethereum/go-ethereum/p2p/discover"
	"github.com/ethereum/go-ethereum/p2p/nat"
	"github.com/qjpcpu/ethereum/key"
	"github.com/qjpcpu/log"
	"io/ioutil"
	"os"
	"strings"
)

const messageId = 0

type Message string

func MyProtocol() p2p.Protocol {
	return p2p.Protocol{
		Name:    "MyProtocol",
		Version: 1,
		Length:  1,
		Run:     msgHandler,
	}
}

var (
	nodeName string
	port     string
	bootnode string
)

func init() {
	log.InitLog(log.LogOption{Format: log.CliFormat})
	flag.StringVar(&nodeName, "name", "", "node name")
	flag.StringVar(&port, "port", "", "listen port")
	flag.StringVar(&bootnode, "bootstrap", "", "bootstrap node")
}

func bootstrapNodes() []*discover.Node {
	var nodes []*discover.Node
	if bootnode != "" {
		log.Infof("bootstrap nodes:%+v", bootnode)
		nodes = append(nodes, discover.MustParseNode(bootnode))
	}
	return nodes
}

func parseArgs() {
	flag.Parse()
	if port == "" {
		log.Error("no port")
		os.Exit(1)
	}
	if nodeName == "" {
		log.Error("no node name")
		os.Exit(1)
	}
	if !strings.HasPrefix(port, ":") {
		port = ":" + port
	}
}

func getPrivateKey() *ecdsa.PrivateKey {
	os.MkdirAll(nodeName, 0777)
	filename := nodeName + "/private-key"
	var pk *ecdsa.PrivateKey
	for loop := true; loop; loop = false {
		data, err := ioutil.ReadFile(filename)
		if err != nil {
			break
		}
		pk = key.PrivateKeyFromBytes(data)
		log.Info("load private key from file")
		return pk
	}
	pk, _ = crypto.GenerateKey()
	log.Info("create new private key")
	ioutil.WriteFile(filename, key.PrivateKeyToBytes(pk), 0644)
	return pk
}

func main() {
	parseArgs()
	nodekey := getPrivateKey()
	//logger := ethlog.New()
	//logger.SetHandler(ethlog.StderrHandler)
	srv := p2p.Server{
		Config: p2p.Config{
			MaxPeers:   10,
			PrivateKey: nodekey,
			Name:       nodeName,
			ListenAddr: port,
			Protocols:  []p2p.Protocol{MyProtocol()},
			NAT:        nat.Any(),
			//		Logger:         logger,
			BootstrapNodes: bootstrapNodes(),
		},
	}
	if err := srv.Start(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	log.Info("started..", srv.NodeInfo().Enode)
	select {}
}

func msgHandler(peer *p2p.Peer, ws p2p.MsgReadWriter) error {
	// send msg
	log.Infof("new peer connected:%v", peer.String())
	p2p.SendItems(ws, messageId, "hello new peer "+peer.Name())
	for {
		msg, err := ws.ReadMsg()
		if err != nil {
			log.Warningf("peer %s disconnected", peer.Name())
			return err
		}

		var myMessage [1]Message
		err = msg.Decode(&myMessage)
		if err != nil {
			// handle decode error
			continue
		}

		log.Info("code:", msg.Code, "receiver at:", msg.ReceivedAt, "msg:", myMessage)
		switch myMessage[0] {
		case "foo":
			err := p2p.SendItems(ws, messageId, "bar")
			if err != nil {
				return err
			}
		default:
			log.Info("recv:", myMessage)
		}
	}
}
