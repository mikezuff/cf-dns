package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http/httptrace"
	"os"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	var (
		trace bool
	)
	flag.BoolVar(&trace, "trace", false, "enable HTTP tracing")
	flag.Parse()

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	if trace {
		ht := &httptrace.ClientTrace{
			GotConn: func(connInfo httptrace.GotConnInfo) {
				fmt.Printf("Got Conn: %+v\n", connInfo)
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				fmt.Printf("DNS Info: %+v\n", dnsInfo)
			},
			WroteHeaderField: func(key string, value []string) {
				fmt.Printf("Wrote Header: %s: %v\n", key, value)
			},
		}
		ctx = httptrace.WithClientTrace(ctx, ht)
	}

	u, err := api.UserDetails(ctx)
	if err != nil {
		log.Fatal(err)
	}

	id, err := api.ZoneIDByName("zuff.dev")
	if err != nil {
		log.Fatal(err)
	}

	zone, err := api.ZoneDetails(ctx, id)
	if err != nil {
		log.Fatal(err)
	}

	enc := json.NewEncoder(os.Stdout)

	type out struct {
		User cloudflare.User `json:"user"`
		Zone cloudflare.Zone `json:"zone"`
	}
	o := out{
		User: u,
		Zone: zone,
	}
	enc.Encode(o)
}
