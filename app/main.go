package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"time"
	"context"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"sourcegraph.com/sourcegraph/appdash"
	appdashot "sourcegraph.com/sourcegraph/appdash/opentracing"
	"sourcegraph.com/sourcegraph/appdash/traceapp"
)

var (
	port           = flag.Int("port", 8080, "Example app port.")
	appdashPort    = flag.Int("appdash.port", 8700, "Run appdash locally on this port.")
	lightstepToken = flag.String("lightstep.token", "", "Lightstep access token.")
)

func main(){
	flag.Parse()

	//var tracer opentracing.Tracer


	InitAppdash()

	router := mux.NewRouter()
	router.HandleFunc("/talk", talk).Methods("GET")
	log.Fatal(http.ListenAndServe(":8080", router))
	
}

// func main() {
// 	// ctx := context.Background()

// 	flag.Parse()

// 	var tracer opentracing.Tracer

// 	// Would it make sense to embed Appdash?
// 	addr := startAppdashServer(8700)
// 	tracer = appdashot.NewTracer(appdash.NewRemoteCollector(addr))

// 	opentracing.InitGlobalTracer(tracer)

// 	router := mux.NewRouter()
// 	router.HandleFunc("/talk", talk).Methods("GET")

// 	log.Fatal(http.ListenAndServe(":8080", router))

// 	// callA(ctx)
// }

// func callA(ctx context.Context) {
// 	span, ctx := opentracing.StartSpanFromContext(ctx, "GetSeatLayout")
// 	defer span.Finish()

// }

func talk(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	//span := opentracing.StartSpanFromContext(ctx, "talk")

	span, ctx := opentracing.StartSpanFromContext(ctx, "/talk")
	defer span.Finish()
	status := getStatus(ctx)
	getInfo(ctx)
	json.NewEncoder(w).Encode(status)
}

func getStatus(ctx context.Context) string {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get data")

	getDetail(ctx)
	

	defer span.Finish()
	return "ok"

}

func getDetail(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get detail")

	sum := 0
	text := "aaa"
	for i := 0; i < 10000; i++ {
		sum += i
		text +="bbb"
	}

	defer span.Finish()

}

func getInfo(ctx context.Context) {
	span, ctx := opentracing.StartSpanFromContext(ctx, "get info")
	getDetail(ctx)
	defer span.Finish()

}


func InitAppdash() {
	go setupTracer(8700, 3600, "")
}

// setupTracer must be called in a separate goroutine, as it's blocking
func setupTracer(appdashPort int, ttl int, server string) {

	// Tracer setup
	memStore := appdash.NewMemoryStore()

	// keep last hour of traces. In production, this will need to be modified.
	store := &appdash.RecentStore{
		MinEvictAge: time.Duration(ttl) * time.Second,
		DeleteStore: memStore,
	}
	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
	if err != nil {
		log.Fatalln("appdash", err)
	}

	collectorPort := l.Addr().String()
	//logging.Debug.Println("collector listening on", collectorPort)

	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
	go cs.Start()

	if server == "" {
		server = fmt.Sprintf("http://localhost:%d", appdashPort)
	}

	appdashURL, err := url.Parse(server)
	tapp, err := traceapp.New(nil, appdashURL)
	if err != nil {
		log.Fatal(err)
	}
	tapp.Store = store
	tapp.Queryer = memStore

	tracer := appdashot.NewTracer(appdash.NewRemoteCollector(collectorPort))
	opentracing.InitGlobalTracer(tracer)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp); err != nil {
		time.Sleep(15 * time.Second) // sleep 15 seconds so we don't run into port conflicts, but if we do again, exit
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp))
	}
}


// func startAppdashServer(appdashPort int) string {
// 	store := appdash.NewMemoryStore()

// 	// Listen on any available TCP port locally.
// 	l, err := net.ListenTCP("tcp", &net.TCPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 0})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	collectorPort := l.Addr().String()

// 	// Start an Appdash collection server that will listen for spans and
// 	// annotations and add them to the local collector (stored in-memory).
// 	cs := appdash.NewServer(l, appdash.NewLocalCollector(store))
// 	go cs.Start()

// 	// Print the URL at which the web UI will be running.
// 	appdashURLStr := fmt.Sprintf("http://localhost:%d", appdashPort)
// 	appdashURL, err := url.Parse(appdashURLStr)
// 	if err != nil {
// 		log.Fatalf("Error parsing %s: %s", appdashURLStr, err)
// 	}
// 	fmt.Printf("To see your traces, go to %s/traces\n", appdashURL)

// 	// Start the web UI in a separate goroutine.
// 	tapp, err := traceapp.New(nil, appdashURL)
// 	if err != nil {
// 		log.Fatalf("Error creating traceapp: %v", err)
// 	}
// 	tapp.Store = store
// 	tapp.Queryer = store
// 	go func() {
// 		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", appdashPort), tapp))
// 	}()
// 	return fmt.Sprintf(":%d", collectorPort)
// }
