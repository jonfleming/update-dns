package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/cloudflare/cloudflare-go"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %s", err)
	}

	api, err := cloudflare.NewWithAPIToken(os.Getenv("CLOUDFLARE_API_KEY"))
	if err != nil {
		log.Fatalf("Error connecting to Cloudflare: %s", err)
	}

	// Check for batch mode (file + IP) or single domain mode
	if len(os.Args) == 3 {
		// Batch mode: update-dns.exe <domains_file> <ip>
		domainsFile := os.Args[1]
		ip := os.Args[2]
		processBatch(api, domainsFile, ip)
	} else if len(os.Args) == 4 {
		// Single domain mode: update-dns.exe <domain> <subdomain> <ip>
		domain := os.Args[1]
		subdomain := os.Args[2]
		ip := os.Args[3]
		updateSingleDomain(api, domain, subdomain, ip)
	} else {
		fmt.Println("Usage:")
		fmt.Printf("  Batch mode:  %s <domains_file> <ip>\n", os.Args[0])
		fmt.Printf("  Single mode: %s <domain> <subdomain> <ip>\n", os.Args[0])
		fmt.Println("\nExamples:")
		fmt.Printf("  %s domains.txt 192.168.1.100\n", os.Args[0])
		fmt.Printf("  %s example.com www 192.168.1.100\n", os.Args[0])
		fmt.Printf("  %s example.com @ 192.168.1.100  (for base domain)\n", os.Args[0])
		os.Exit(1)
	}
}

func processBatch(api *cloudflare.API, domainsFile, ip string) {
	file, err := os.Open(domainsFile)
	if err != nil {
		log.Fatalf("Error opening domains file '%s': %s", domainsFile, err)
	}
	defer file.Close()

	fmt.Printf("Starting DNS updates for IP address: %s\n", ip)
	fmt.Printf("Using domains file: %s\n\n", domainsFile)

	scanner := bufio.NewScanner(file)
	successCount := 0
	errorCount := 0

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines
		if line == "" {
			continue
		}

		var domain, subdomain string

		if strings.HasPrefix(line, ".") {
			// Base domain - remove the leading dot
			domain = line[1:]
			subdomain = "@"
			fmt.Printf("Updating base domain: %s with IP %s\n", domain, ip)
		} else {
			// Subdomain - split at first dot
			parts := strings.SplitN(line, ".", 2)
			if len(parts) != 2 {
				fmt.Printf("Error: Invalid domain format '%s' - skipping\n", line)
				errorCount++
				continue
			}
			subdomain = parts[0]
			domain = parts[1]
			fmt.Printf("Updating subdomain: %s.%s with IP %s\n", subdomain, domain, ip)
		}

		err := updateSingleDomain(api, domain, subdomain, ip)
		if err != nil {
			fmt.Printf("Error updating %s: %s\n", line, err)
			errorCount++
		} else {
			successCount++
		}
		fmt.Println()
	}

	if err := scanner.Err(); err != nil {
		log.Fatalf("Error reading domains file: %s", err)
	}

	fmt.Printf("DNS update process completed.\n")
	fmt.Printf("Successfully updated: %d domains\n", successCount)
	if errorCount > 0 {
		fmt.Printf("Failed to update: %d domains\n", errorCount)
	}
}

func updateSingleDomain(api *cloudflare.API, domain, subdomain, ip string) error {
	var recordName string
	if subdomain == "@" {
		// Base domain record
		recordName = domain
	} else {
		// Subdomain record
		recordName = fmt.Sprintf("%s.%s", subdomain, domain)
	}

	zoneID, err := api.ZoneIDByName(domain)
	if err != nil {
		return fmt.Errorf("error getting zone ID for domain '%s': %s", domain, err)
	}

	records, _, err := api.ListDNSRecords(context.Background(), cloudflare.ZoneIdentifier(zoneID), cloudflare.ListDNSRecordsParams{})
	if err != nil {
		return fmt.Errorf("error listing DNS records: %s", err)
	}

	var oldIP string
	var record cloudflare.DNSRecord
	recordFound := false

	for _, r := range records {
		if r.Type == "A" && r.Name == recordName {
			record = r
			oldIP = r.Content
			recordFound = true
			break
		}
	}

	if !recordFound {
		return fmt.Errorf("A record for '%s' not found", recordName)
	}

	if oldIP == ip {
		fmt.Printf("No change required for %s (already %s)\n", recordName, ip)
		return nil
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
		return fmt.Errorf("error updating DNS record: %s", err)
	}

	fmt.Printf("Updated IP address from %s to %s\n", oldIP, ip)
	return nil
}
