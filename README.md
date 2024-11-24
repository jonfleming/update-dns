# Update DNS

This Go program updates a DNS record using the Cloudflare API.

## Prerequisites

- Go installed on your system.
- A Cloudflare account with an API token. The token should have permissions to edit DNS records.
- Set the environment variable `CLOUDFLARE_API_KEY` with your Cloudflare API token.

## Usage

To run the program, use the following command:

```bash
go run update-dns.go <domain> <subdomain> <new-ip>
```

- `<domain>`: The domain name (e.g., example.com).
- `<subdomain>`: The subdomain to update (e.g., www).
- `<new-ip>`: The new IP address for the DNS record.

## How It Works

1. The program initializes a connection to the Cloudflare API using the API token.
2. It retrieves the zone ID for the specified domain.
3. It lists all DNS records for the domain and finds the A record for the specified subdomain.
4. If the current IP address matches the new IP address, no changes are made.
5. If the IP addresses differ, the program updates the DNS record with the new IP address.
6. The program outputs a message indicating the IP address has been updated.

## Error Handling

The program logs errors and exits if it encounters issues connecting to Cloudflare, retrieving DNS records, or updating the DNS record.

## Example

```bash
update-dns example.com www 192.0.2.1
```

This command updates the A record for `www.example.com` to the IP address `192.0.2.1`.
