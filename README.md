# Update DNS

This Go program updates DNS records using the Cloudflare API. It supports both single domain updates and batch processing from a file.

## Prerequisites

- Go installed on your system.
- A Cloudflare account with an API token. The token should have permissions to edit DNS records.
- Set the environment variable `CLOUDFLARE_API_KEY` with your Cloudflare API token.

## Usage

The program supports two modes of operation:

### Single Domain Mode

Update a single DNS record:

```bash
update-dns <domain> <subdomain> <new-ip>
```

- `<domain>`: The domain name (e.g., example.com).
- `<subdomain>`: The subdomain to update (e.g., www). Use `@` for the base domain.
- `<new-ip>`: The new IP address for the DNS record.

### Batch Mode

Update multiple DNS records from a file:

```bash
update-dns <domains-file> <new-ip>
```

- `<domains-file>`: Path to a text file containing domain/subdomain entries (one per line).
- `<new-ip>`: The new IP address for all DNS records.

#### Domains File Format

The domains file should contain one domain/subdomain per line:

```
subdomain.example.com
www.example.com
api.example.com
.example.com
```

- Regular entries like `subdomain.domain.com` update the subdomain A record
- Entries starting with `.` (like `.example.com`) update the base domain A record

## How It Works

### Single Domain Mode
1. The program initializes a connection to the Cloudflare API using the API token.
2. It retrieves the zone ID for the specified domain.
3. It lists all DNS records for the domain and finds the A record for the specified subdomain.
4. If the current IP address matches the new IP address, no changes are made.
5. If the IP addresses differ, the program updates the DNS record with the new IP address.
6. The program outputs a message indicating the IP address has been updated.

### Batch Mode
1. The program reads the domains file line by line.
2. For each domain entry, it parses the domain and subdomain components.
3. It processes each domain using the same logic as single domain mode.
4. Progress is tracked and reported, showing successful updates and any errors.
5. Individual domain failures don't stop the entire batch process.

## Error Handling

- **Single Mode**: The program logs errors and exits if it encounters issues.
- **Batch Mode**: Individual domain errors are reported but don't stop processing of remaining domains. A summary is provided at the end showing success/failure counts.

## Install
```bash
go install github.com/jonfleming/update-dns@latest
```

## Examples

### Single Domain Updates

Update a subdomain:
```bash
update-dns example.com www 192.0.2.1
```

Update the base domain:
```bash
update-dns example.com @ 192.0.2.1
```

### Batch Updates

Update multiple domains from a file:
```bash
update-dns domains.txt 192.0.2.1
```

With a `domains.txt` file containing:
```
www.example.com
api.example.com
.example.com
mail.anotherdomain.com
```

This will update:
- `www.example.com` → 192.0.2.1
- `api.example.com` → 192.0.2.1  
- `example.com` (base domain) → 192.0.2.1
- `mail.anotherdomain.com` → 192.0.2.1
