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

// Version is the current CLI version. Update when tagging releases.
const Version = "v1.3.3"

func main() {
	// If invoked without args or with -v/--version, show version + help and exit.
	if len(os.Args) == 1 || len(os.Args) == 2 || os.Args[1] == "-h" || os.Args[1] == "-v" || os.Args[1] == "--version" {
		printUsage()
		return
	}

	API_KEY := os.Getenv("CLOUDFLARE_API_KEY")
	// Load environment variables from .env file
	if API_KEY == "" {
		err := godotenv.Load()
		if err != nil {
			log.Fatalf("Error loading .env file and CLOUDFLARE_API_KEY is not set: %s", err)
		}
		API_KEY = os.Getenv("CLOUDFLARE_API_KEY")
	}

	if API_KEY == "" {
		log.Fatalf("CLOUDFLARE_API_KEY environment variable is not set")
	}

	api, err := cloudflare.NewWithAPIToken(API_KEY)
	if err != nil {
		log.Fatalf("Error connecting to Cloudflare: %s", err)
	}

	// Accept either: batch file or single domain in a simpler form.
	// New preferred single-domain usage: update-dns <domain-or-fqdn> <ip>
	if len(os.Args) == 3 {
		arg := os.Args[1]
		ip := os.Args[2]

		// If the first arg is an existing file, treat as batch mode.
		if fi, err := os.Stat(arg); err == nil && !fi.IsDir() {
			processBatch(api, arg, ip)
			return
		}

		// Single domain mode (preferred): resolve zone and subdomain from FQDN.
		if strings.HasPrefix(arg, ".") {
			// Leading dot indicates base domain (legacy style)
			domain := arg[1:]
			if err := updateSingleDomain(api, domain, "@", ip); err != nil {
				log.Fatalf("Error updating %s: %s", arg, err)
			}
			return
		}

		domain, subdomain, err := resolveZoneAndSubdomain(api, arg)
		if err != nil {
			log.Fatalf("Error resolving domain '%s': %s", arg, err)
		}
		if err := updateSingleDomain(api, domain, subdomain, ip); err != nil {
			log.Fatalf("Error updating %s: %s", arg, err)
		}
		return
	} else if len(os.Args) == 4 {
		// Legacy single domain mode: update-dns.exe <domain> <subdomain> <ip>
		domain := os.Args[1]
		subdomain := os.Args[2]
		ip := os.Args[3]
		if err := updateSingleDomain(api, domain, subdomain, ip); err != nil {
			log.Fatalf("Error updating %s.%s: %s", subdomain, domain, err)
		}
		return
	}

	printUsage()
	os.Exit(1)
}

func printUsage() {
	// Strip the path from the executable name for clarity.
	exename := os.Args[0]

	if idx := strings.LastIndex(exename, string(os.PathSeparator)); idx != -1 {
		exename = exename[idx+1:]
	}

	fmt.Printf("update-dns %s\n\n", Version)
	fmt.Println("Usage:")
	fmt.Printf("  Batch mode:  %s <domains_file> <ip>\n", exename)
	fmt.Printf("  Single mode: %s <domain-or-fqdn> <ip>\n", exename)
	fmt.Printf("  Legacy:      %s <domain> <subdomain> <ip>\n", exename)
	fmt.Println("\nExamples:")
	fmt.Printf("  %s domains.txt 192.168.1.100\n", exename)
	fmt.Printf("  %s example.com 192.168.1.100     (updates base domain)\n", exename)
	fmt.Printf("  %s www.example.com 192.168.1.100 (updates subdomain)\n", exename)
	fmt.Printf("  %s .example.com 192.168.1.100    (legacy: leading dot = base domain)\n", exename)
}

// resolveZoneAndSubdomain attempts to determine the Cloudflare zone (base domain)
// and the subdomain (or "@" for root) from a provided FQDN by probing
// `api.ZoneIDByName` from left-to-right. For example, given "www.sub.example.com"
// it will try "www.sub.example.com", then "sub.example.com", then
// "example.com" until a zone is found.
func resolveZoneAndSubdomain(api *cloudflare.API, fqdn string) (string, string, error) {
	fqdn = strings.TrimSpace(fqdn)
	if fqdn == "" {
		return "", "", fmt.Errorf("empty domain")
	}

	parts := strings.Split(fqdn, ".")
	for i := 0; i < len(parts); i++ {
		candidate := strings.Join(parts[i:], ".")
		if _, err := api.ZoneIDByName(candidate); err == nil {
			// candidate is the zone; subdomain is whatever is left of it
			if i == 0 {
				return candidate, "@", nil
			}
			sub := strings.Join(parts[:i], ".")
			return candidate, sub, nil
		}
	}
	return "", "", fmt.Errorf("no Cloudflare zone found for '%s'", fqdn)
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
