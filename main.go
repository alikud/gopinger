package main

import (
	"flag"
	"fmt"
	"os/user"
	"time"

	"sync"

	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
)

type PingerConfig struct {
	Timeout time.Duration
	Send    int
	PingUrl string
}

func SendPingByRouter(cfg PingerConfig, dhcp int, wg *sync.WaitGroup) {

	defer wg.Done()

	pinger, _ := ping.NewPinger(cfg.PingUrl)
	sourceaddr := fmt.Sprintf("192.168.%d.100", dhcp)

	pinger.Source = sourceaddr
	pinger.Timeout = cfg.Timeout
	pinger.SetPrivileged(true)
	pinger.Count = cfg.Send

	err := pinger.Run()

	if err != nil {
		log.Warningf("error with ping %d dhcp", dhcp)

	}

	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats

	log.Infof("Send %d packeges, recived %d for %s\n", pinger.Count, stats.PacketsRecv, sourceaddr)
	if stats.PacketsRecv == 0 {
		log.Warningf("Received 0 packages from dhcp: %d", dhcp)
	}

}

func main() {

	lastDHCP := flag.Int("stop", 11, "A number which represent last dhcp addr. of network interface")
	flag.Parse()

	currentUser, _ := user.Current()
	log.Infof("as %s execute pinger script", currentUser.Username)
	log.Infof("Start script!")

	cfg := PingerConfig{
		Timeout: 5 * time.Second,
		Send:    8,
		PingUrl: "www.google.com",
	}

	var wg sync.WaitGroup

	for i := 11; i < *lastDHCP; i++ {
		wg.Add(1)
		go SendPingByRouter(cfg, i, &wg)
		wg.Wait()
	}

}

// go run main.go -stop 55
