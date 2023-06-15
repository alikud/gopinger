package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
	"os/user"
	"time"
)

func RunPingerWithCtx(ctx context.Context, errChan chan error, pinger *ping.Pinger) {
	err := pinger.Run()
	if err != nil {
		errChan <- errors.New(err.Error())
	}
	select {
	case <-ctx.Done():
		errChan <- ctx.Err()
	case <-errChan:
		close(errChan)
	default:
		errChan <- nil
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
		log.Warningf("Received 0 packages from dhcp: %d", dhcp)
		//return errors.New("get 0 packages after ping")
	}

	return nil
}

func main() {
	currentUser, _ := user.Current()
	log.Infof("as %s execute pinger script", currentUser.Username)
	log.Infof("Start script!")

	const PACKAGECOUNT = 8

	for i := 11; i < 30; i++ {
		SendPingByRouter(PACKAGECOUNT, i)
	}

}

//go run main.go
