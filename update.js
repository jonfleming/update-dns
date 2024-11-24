const Cloudflare = require('cloudflare');
const { program } = require('commander');

program
  .version('1.0.0')
  .description('Update the IP address for a Cloudflare subdomain A record.')
  .option('-d, --domain <domain>', 'Domain name (e.g. example.com)')
  .option('-s, --sub-domain <sub-domain>', 'Sub domain (e.g. my-sub-domain)')
  .option('-i, --ip <ip>', 'IP address to update')
  .parse(process.argv);

const { domain, subDomain, ip } = program.opts();

if (!domain || !subDomain || !ip) {
  console.error('Error: Domain, sub domain and IP are required.');
  process.exit(1);
}

const cloudflare = new Cloudflare({
  apiEmail: process.env.CLOUDFLARE_EMAIL,
  apiKey: process.env.CLOUDFLARE_API_KEY
});

cloudflare.zones.list().then(zones => {
  const zone = zones.filter(z => z.name === domain)[0];

  if (!zone) {
    console.error(`Error: Zone not found for domain ${domain}`);
    process.exit(1);
  }

  return cloudflare.dnsRecords.list(zone.id).then(records => {
    const record = records.filter(r => r.type === 'A' && r.name === `${subDomain}.${domain}`)[0];

    if (!record) {
      console.error(`Error: A record not found for sub domain ${subDomain}`);
      process.exit(1);
    }

    return cloudflare.dnsRecords.update(zone.id, record.id, { type: 'A', name: `${subDomain}.${domain}`, content: ip });
  });
}).then(record => {
  console.log(`Updated A record for ${record.name} to IP address ${record.content}`);
}).catch(err => {
  console.error('Error:', err);
  process.exit(1);
});