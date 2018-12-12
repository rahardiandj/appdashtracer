package tracer

import (
	"fmt"
	"log"
	"net"
	"time"

	"sourcegraph.com/sourcegraph/appdash"
)

func InitAppdash(serverName string, port int) {
	log.Printf("starting tracer on  %+v", port)

	go setupTracer(port, 3600, serverName)
}

func setupTracer(appdashPort int, ttl int, server string) {
	memStore := appdash.NewMemoryStore()

	store := &appdash.RecentStore{
		MinEvictAge: time.Duration(ttl) * time.Second,
		DeleteStore: memStore,
	}

	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})

	if err != nil {
		log.Fatal("appdash", err)
	}

	collectorPort := l.Addr().String()
	log.Printf("collector listening on ", collectorPort)

	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))

	go cs.Start()

	if server == "" {
		server = fmt.Sprintf("http://localhost:%d", appdashPort)
	}
}
