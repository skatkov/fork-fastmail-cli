# Email Tracking Worker

Cloudflare Worker that receives tracking pixels and records email opens.

## Deploy

```bash
cd internal/tracking/worker

# Install dependencies
pnpm install

# Create D1 database
wrangler d1 create email-tracker
# Copy the database_id from output

# Update wrangler.toml with the database_id
# database_id = "your-actual-id-here"

# Initialize database schema
wrangler d1 execute email-tracker --file=schema.sql

# Deploy worker
wrangler deploy

# Set secrets (use the same keys in fastmail-cli setup!)
wrangler secret put TRACKING_KEY
wrangler secret put ADMIN_KEY
```

## Configure fastmail-cli

After deploying, configure fastmail-cli to use your worker:

```bash
fastmail email track setup --worker-url https://email-tracker.<your-subdomain>.workers.dev
```

Enter the same TRACKING_KEY and ADMIN_KEY you set as wrangler secrets.

## Endpoints

- `GET /p/{blob}.gif` - Tracking pixel (records open, returns 1x1 GIF)
- `GET /q/{blob}` - Query opens for a specific tracking ID
- `GET /opens` - Admin endpoint to list all opens (requires Bearer token)
- `GET /health` - Health check

## Sharing with gogcli

This worker is designed to be shared between fastmail-cli and gogcli. Both CLIs store config at `~/.config/email-tracking/` so they use the same keys.
