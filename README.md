# Update DNS

This Go program updates DNS records using the Cloudflare API. It supports both single domain updates and batch processing from a file.

## Prerequisites

- Go installed on your system.
- A Cloudflare account with an API token. The token should have permissions to edit DNS records.
- Set the environment variable `CLOUDFLARE_API_KEY` with your Cloudflare API token.

## Usage

The program supports three ways to invoke updates. The preferred form is a single domain/FQDN plus IP. The first argument may also be a path to a domains file for batch updates. Legacy three-argument form is still supported.

1) Preferred single FQDN form (recommended): resolve zone automatically

```bash
update-dns <domain-or-fqdn> <new-ip>
```

- `<domain-or-fqdn>`: A fully-qualified domain name such as `www.example.com`, `example.com`, or a path to a domains file (see batch mode below). If a file with this name exists, it will be treated as a domains file.
- `<new-ip>`: The new IP address for the DNS record.

Behavior:
- If the first argument is a file path, the program runs in batch mode and treats the file as a list of domains to update.
- If the first argument is a FQDN, the program will automatically determine the Cloudflare zone by trying candidates from left-to-right (e.g., `www.sub.example.com` → tries `www.sub.example.com`, `sub.example.com`, `example.com`), using the first matching zone. The portion left of the found zone becomes the record name (or `@` for the apex/root).

2) Batch mode (explicit file):

```bash
update-dns <domains-file> <new-ip>
```

- `<domains-file>`: Path to a text file containing domain/subdomain entries (one per line).
- `<new-ip>`: The new IP address for all DNS records.

3) Legacy form (still supported):

```bash
update-dns <domain> <subdomain> <new-ip>
```

- `<domain>`: The zone (e.g., `example.com`).
- `<subdomain>`: The record name to update (`www` or `@` for the root).
- `<new-ip>`: The new IP address.

#### Domains File Format

The domains file should contain one domain/subdomain per line. Examples:

```
www.example.com
api.example.com
mail.subdomain.example.com
.example.com
```

- Regular entries like `subdomain.domain.com` update that subdomain's A record.
- Entries starting with `.` (like `.example.com`) explicitly indicate the base/apex domain record should be updated.

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
