package scheduling

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/go-ping/ping"
	"github.com/grussorusso/serverledge/internal/registration"
)

type Latency struct {
	latencyMap map[string]int // <k, v> = <ip, latency>
}

var reg *registration.Registry
var lat *Latency

// NewLatency creates new Latency struct
func newLatencyMap() {
	serverMap, err := reg.GetAll(true)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	delete(serverMap, reg.Key) // delete myself
	l := make(map[string]int)
	//retrieve ip addresses from urls
	for _, url := range serverMap {
		ip := url[7 : len(url)-5]
		l[ip] = 0
	}

	lat = &Latency{
		latencyMap: l,
	}
}

func monitoring() error {

	//Create new latency map to retrieve any nodes that are no longer reachable or recently added
	newLatencyMap()

	pinger, err := ping.NewPinger("0.0.0.0")
	if err != nil {
		log.Println(err)
		return err
	}
	pinger.SetPrivileged(true)
	for k := range lat.latencyMap {
		pinger.SetAddr(k)
		log.Println("Target host: ", pinger.IPAddr())
		pinger.Count = 1
		pinger.Run()                                        // blocks until finished
		avgrtt := pinger.Statistics().AvgRtt.Milliseconds() // get send/receive/rtt stats
		log.Println("avg rtt: ", avgrtt)
		lat.latencyMap[k] = int(avgrtt)
	}
	log.Println("new latency map: ", lat.latencyMap)
	return nil
}

func runTicker() {
	latencyTicker := time.NewTicker(time.Duration(60 * time.Second))

	for _ = range latencyTicker.C {
		log.Println("latency monitoring..")
		err := monitoring()
		if err != nil {
			log.Fatal(err)
			os.Exit(1)
		}
	}
}

// Init latency monitoring of nodes under a specific local Area
func InitLatencyMonitoring(r *registration.Registry) {
	reg = r
	newLatencyMap()

	log.Println("Init latenciesMap: ", lat.latencyMap)

	log.Println("Begin latency monitoring..")
	err := monitoring()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	go runTicker()
}

func minLatency(latMapClient map[string]interface{}, targetHost string) (bool, error) {

	ip := targetHost[7 : len(targetHost)-5]
	latClient, err := strconv.Atoi(latMapClient[ip].(string))
	if err != nil {
		log.Println(err)
		return false, err
	}
	return latClient <= lat.latencyMap[ip], nil

}
