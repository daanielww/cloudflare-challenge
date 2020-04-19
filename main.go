package main

import (
	"flag"
	"golang.org/x/net/icmp"
	"golang.org/x/net/ipv4"
	"golang.org/x/net/ipv6"
	"log"
	"net"
	"os"
	"time"
)

/*
DOCUMENTATION

For EXTRA CREDIT I've added support for both ipv6 and ipv4
Please only have 1 of either "-4" or "-6" flag, for the ipv4 and ipv6 configs respectively
Note: Sudo permissions may be necessary

Example Commands:

sudo go run main.go -6 -ip=2001:4860:4860::8888
sudo go run main.go -4 -ip=8.8.8.8
*/

// Retrieved from golang.org/x/net/internal/iana.ProtocolICMP
var ProtocolICMPv4 int = 1
var ProtocolICMPv6 int = 58

type arguments struct {
	targetIP string
	ipv4     bool
	ipv6     bool
}

type configuration struct {
	targetIP   string
	connection *icmp.PacketConn
	msgType    icmp.Type
	replyType  icmp.Type
	protocol   int
}

func getArgs(args *arguments) {
	flag.StringVar(&args.targetIP, "ip", "", "IP Address to Ping")
	flag.BoolVar(&args.ipv4, "4", false, "ipv4")
	flag.BoolVar(&args.ipv6, "6", false, "ipv6")

	flag.Parse()
	log.Println("Target IP: ", args.targetIP)
}

func configure() *configuration {
	args := &arguments{}
	getArgs(args)

	config := &configuration{targetIP: args.targetIP}
	if args.ipv6 {
		log.Println("Type: ipv6")
		c, err := icmp.ListenPacket("ip6:ipv6-icmp", "::")
		if err != nil {
			log.Fatalf("listen err, %s", err)
		}
		config.connection = c
		config.msgType, config.replyType = ipv6.ICMPTypeEchoRequest, ipv6.ICMPTypeEchoReply
		config.protocol = ProtocolICMPv6
	} else {
		log.Println("Type: ipv4")
		c, err := icmp.ListenPacket("ip4:icmp", "0.0.0.0")
		if err != nil {
			log.Fatalf("listen err, %s", err)
		}
		config.connection = c
		config.msgType, config.replyType = ipv4.ICMPTypeEcho, ipv4.ICMPTypeEchoReply
		config.protocol = ProtocolICMPv4
	}
	return config
}

func main() {

	config := configure()
	defer config.connection.Close()

	msgSent := 0
	msgReceived := 0
	SeqNumber := 1

	for {
		wm := icmp.Message{
			Type: config.msgType,
			Code: 0,
			Body: &icmp.Echo{
				ID: os.Getpid() & 0xffff, Seq: SeqNumber,
				Data: []byte("ping"),
			},
		}
		wb, err := wm.Marshal(nil)
		if err != nil {
			log.Fatal(err)
		}

		if _, err = config.connection.WriteTo(wb, &net.IPAddr{IP: net.ParseIP(config.targetIP)}); err != nil {
			log.Fatalf("WriteTo err, %s", err)
		}
		sentTime := time.Now()
		msgSent++

		rb := make([]byte, 2000)
		n, _, err := config.connection.ReadFrom(rb)
		if err != nil {
			log.Fatal(err)
		}
		receiveTime := time.Now()

		rm, err := icmp.ParseMessage(config.protocol, rb[:n])
		if err != nil {
			log.Fatal(err)
		}
		switch rm.Type {
		case config.replyType:
			msgReceived++
		}

		packetLoss := float64(msgSent-msgReceived) / float64(msgSent) * 100
		latency := receiveTime.Sub(sentTime)
		log.Println("Packet Loss: ", packetLoss, " %,  Latency: ", latency)

		SeqNumber++
		time.Sleep(2 * time.Second)
	}
}
