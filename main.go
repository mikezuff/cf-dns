package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net/http/httptrace"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
)

func usage() {
	log.Printf("Usage: %s [options] -zone <zone_name>\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var (
		trace    bool
		testZone string
	)
	flag.BoolVar(&trace, "trace", false, "enable HTTP tracing")
	flag.StringVar(&testZone, "zone", "", "test zone")
	flag.Parse()

	if testZone == "" {
		usage()
		log.Fatal("missing zone")
	}

	api, err := cloudflare.New(os.Getenv("CLOUDFLARE_API_KEY"), os.Getenv("CLOUDFLARE_API_EMAIL"))
	if err != nil {
		log.Fatal("error creating API:", err)
	}

	ctx := context.Background()
	if trace {
		ht := &httptrace.ClientTrace{
			GotConn: func(connInfo httptrace.GotConnInfo) {
				log.Printf("Got Conn: %+v\n", connInfo)
			},
			DNSDone: func(dnsInfo httptrace.DNSDoneInfo) {
				log.Printf("DNS Info: %+v\n", dnsInfo)
			},
			WroteHeaderField: func(key string, value []string) {
				log.Printf("Wrote Header: %s: %v\n", key, value)
			},
		}
		ctx = httptrace.WithClientTrace(ctx, ht)
	}

	u, err := api.UserDetails(ctx)
	if err != nil {
		log.Fatal("error: UserDetails:", err)
	}

	zid, err := api.ZoneIDByName(testZone)
	if err != nil {
		log.Fatal("error: ZoneIDByName:", err)
	}

	zone, err := api.ZoneDetails(ctx, zid)
	if err != nil {
		log.Fatal("error: ZoneDetails:", err)
	}

	recs, err := api.DNSRecords(ctx, zid, cloudflare.DNSRecord{})
	if err != nil {
		log.Fatal("error: DNSRecords:", err)
	}

	testNames := []string{"cf-dns-test-1", "cf-dns-test-2"}
	found := false
	for _, testName := range testNames {
		for _, rec := range recs {
			if strings.HasPrefix(rec.Name, testName) {
				found = true
				break
			}
		}
	}

	var (
		rrIDs       []string // DNSRecord IDs for test names
		respRecords []*cloudflare.DNSRecord
	)
	if found {
		log.Printf("Zone %s contains one of the test names %v, not modifying the zone.\n",
			testZone, testNames)
	} else {
		resp1, err := api.CreateDNSRecord(ctx, zid, cloudflare.DNSRecord{
			Type:    "CNAME",
			Name:    testNames[0],
			Content: "ns1.zuffs.net",
			TTL:     1,
		})
		if err != nil {
			log.Fatal("error: CreateDNSRecord", err)
		}
		if !resp1.Success {
			log.Fatalf("creation of %s failed: %v\n", testNames[0], resp1.Errors)
		}
		rrIDs = append(rrIDs, resp1.Result.ID)
		respRecords = append(respRecords, &resp1.Result)

		testRR2 := resp1.Result
		testRR2.Name = testNames[1]
		err = api.UpdateDNSRecord(ctx, zid, rrIDs[0], testRR2)
		if err != nil {
			log.Fatal("update failed", err)
		}
		rec2, err := api.DNSRecord(ctx, zid, rrIDs[0])
		if err != nil {
			log.Fatal("get2 failed:", err)
		}
		rrIDs = append(rrIDs, rec2.ID)
		respRecords = append(respRecords, &rec2)

		err = api.DeleteDNSRecord(ctx, zid, rrIDs[1])
		if err != nil {
			log.Fatal("deletion failed", err)
		}

		rec3, err := api.DNSRecord(ctx, zid, rrIDs[0])
		if err != nil {
			log.Printf("Get after Delete returned error: %s", err)
		} else {
			log.Printf("After deletion, query for %s returned a result\n", rrIDs[0])
			respRecords = append(respRecords, &rec3)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	type out struct {
		User      cloudflare.User         `json:"user"`
		Zone      cloudflare.Zone         `json:"zone"`
		IDs       []string                `json:"ids"`
		Responses []*cloudflare.DNSRecord `json:"responses"`
	}
	o := out{
		User:      u,
		Zone:      zone,
		IDs:       rrIDs,
		Responses: respRecords,
	}
	enc.Encode(o)
}
