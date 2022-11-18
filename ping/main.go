package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/go-ping/ping"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const PACKAGECOUNT = 3

func RebootRouter(key string, dhcpID int) {
	url := fmt.Sprintf("http://localhost:8881/reboot?key=%s&id=%d", key, dhcpID)
	_, err := http.Get(url)
	if err != nil {
		log.Error(err)
	}
}

func SendPingByRouter(packegCount int, dhcp int) error {
	pinger, _ := ping.NewPinger("www.google.com")
	sourceaddr := fmt.Sprintf("192.168.%d.100", dhcp)

	pinger.Source = sourceaddr
	pinger.Timeout = 20 * time.Second
	pinger.SetPrivileged(true)
	pinger.Count = packegCount

	err := pinger.Run() // Blocks until finished.
	if err != nil {
		log.Error(err)
		return err
	}

	//fmt.Printf("Check %s\n", sourceaddr)
	stats := pinger.Statistics() // get send/receive/duplicate/rtt stats

	log.Infof("Send %d packeges, recived %d for %s\n", pinger.Count, stats.PacketsRecv, sourceaddr)
	if stats.PacketsRecv == 0 {
		log.Error("Received 0 packages from dhcp: %d", dhcp)
		return errors.New("get 0 packages after ping")
	}

	return nil
}

func SendTelegramAllert(dhcp int, serverID string) {
	token := "5659413160:AAFoiIsvluEfluRExg29CNLZFZpzBcwr-vs"
	//chatID := "951131561"

	//my
	chatID := "170308082"
	q := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage?chat_id=%s&text=%s , %d", token, chatID, serverID, dhcp)
	_, err := http.Get(q)
	if err != nil {
		log.Errorf("Error to sending telegram")
	}
}

func main() {
	//currentUser, _ := user.Current()
	//path := fmt.Sprintf("/%s/pinger.log", currentUser.Username)
	//f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	log.SetOutput(os.Stdout)
	//log.SetFormatter(&log.JSONFormatter{})
	//log.SetOutput(os.Stdout)
	log.Infof("Start script!")

	var count int
	var ignored string
	var apikey string
	var serverName string

	flag.IntVar(&count, "p", 11, "enter port to check, default 11")
	flag.StringVar(&ignored, "i", "", "ports to ignore by , ")
	flag.StringVar(&apikey, "key", "", "key which access to web api")
	flag.StringVar(&serverName, "name", "some server", "server name to correctly detect problem")
	flag.Parse()

	s := strings.Split(ignored, ",")

	var ports []int
	for _, value := range s {
		val, _ := strconv.Atoi(value)
		ports = append(ports, val)
	}
	//i := count
	for i := 11; i < count; i++ {
		if ok := slices.Contains(ports, i); !ok {
			err := SendPingByRouter(PACKAGECOUNT, i)
			if err != nil {
				log.Errorf("Start reboot router %d", i)
				RebootRouter(apikey, i)
				time.Sleep(80 * time.Second)
				log.Infof("Start ping router after reboot %d", i)
				err := SendPingByRouter(PACKAGECOUNT, i)
				if err != nil {
					log.Error(err)
					log.Infof("Send telegram allert to %s %d", serverName, i)
					SendTelegramAllert(i, serverName)
				}
			}
		}
	}

}

//sudo go run /home/mac/code/ping/main.go -p 30 -key shir -name po1
