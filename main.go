package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"os/user"
	"strconv"
	"strings"
	"time"
)

func RunPingerWithCtx(ctx context.Context, quitCh chan error, pinger *ping.Pinger) {
	err := pinger.Run()
	if err != nil {
		quitCh <- errors.New(err.Error())
	}
	select {
	case <-ctx.Done():
		quitCh <- ctx.Err()
	case <-quitCh:
		return
	default:
		quitCh <- nil
	}
}

func SendPingByRouter(packegCount int, dhcp int) error {
	pinger, _ := ping.NewPinger("www.google.com")
	sourceaddr := fmt.Sprintf("192.168.%d.100", dhcp)

	pinger.Source = sourceaddr
	pinger.Timeout = 3 * time.Second
	pinger.SetPrivileged(true)
	pinger.Count = packegCount
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	// Can run a lot of time
	quitCh := make(chan error, 1)
	go RunPingerWithCtx(ctx, quitCh, pinger)

	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats

	log.Infof("Send %d packeges, recived %d for %s\n", pinger.Count, stats.PacketsRecv, sourceaddr)
	if stats.PacketsRecv == 0 {
		log.Error("Received 0 packages from dhcp: %d", dhcp)
		return errors.New("get 0 packages after ping")
	}

	return nil
}

func main() {
	currentUser, _ := user.Current()
	log.Infof("as %s execute pinger script", currentUser.Username)

	//f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	//log.SetOutput(f)
	//log.SetFormatter(&log.JSONFormatter{})
	//log.SetOutput(os.Stdout)

	log.Infof("Start script!")

	var count int
	var ignored string
	var serverUID string
	const PACKAGECOUNT = 8

	flag.IntVar(&count, "p", 11, "enter port to check, default 11")
	flag.StringVar(&ignored, "i", "", "ports to ignore by , ")
	flag.StringVar(&serverUID, "name", "some server", "server name to correctly detect problem")
	flag.Parse()

	s := strings.Split(ignored, ",")

	var ports []int
	for _, value := range s {
		val, err := strconv.Atoi(value)

		if err != nil {
			log.Errorf("Can't convert symbol to in")
			return
		}

		ports = append(ports, val)
	}

	for i := 11; i < count; i++ {
		if ok := slices.Contains(ports, i); !ok {
			err := SendPingByRouter(PACKAGECOUNT, i)
			if err != nil {
				log.Warnf("router: %d have problems with ping, ", err.Error())
			}
		}
	}

}

//go run main.go -p 30 -name po1
