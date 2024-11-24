package main

import (
	"context"
	"fmt"
	"log"

	"os"

	"github.com/cloudflare/cloudflare-go"
)

func main() {
	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_KEY"))
	if err != nil {
		log.Fatalf("Error connecting to Cloudflare: %s", err)
	}

	if len(os.Args) != 4 {
		log.Fatalf("Usage: %s <domain> <subdomain> <ip>", os.Args[0])
	}
	domain := os.Args[1]
	subdomain := os.Args[2]
	ip := os.Args[3]

	recordName := fmt.Sprintf("%s.%s", subdomain, domain)
	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		log.Fatalf("Error getting DNS records: %s", err)
	}

	records, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})

	//records, err := api.DNSRecords(zone, cloudflare.DNSRecord{Type: "A", Name: recordName})
	if err != nil {
		log.Fatalf("Error finding DNS records: %s", err)
	}

	var oldIP string
	var record cloudflare.DNSRecord
	for _, r := range records {
		if r.Type == "A" && r.Name == recordName {
			record = r
			oldIP = r.Content
			break
		}
	}

	if oldIP == ip {
		log.Println("No change required")
		return
	}

	record.Content = ip
	recordParams := cloudflare.UpdateDNSRecordParams{
		ID:      record.ID,
		Type:    record.Type,
		Name:    record.Name,
		Content: record.Content,
		TTL:     record.TTL,
		Proxied: record.Proxied,
	}

	_, err = api.UpdateDNSRecord(context.Background(), cloudflare.ZoneIdentifier(zoneID), recordParams)

	if err != nil {
		log.Fatalf("Error updating DNS record: %s", err)
	}

	fmt.Printf("Updated IP address from %s to %s\n", oldIP, ip)
}
