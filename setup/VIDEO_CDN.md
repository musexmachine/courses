# SETUP_CDN.md

# Bunny.net CDN Setup for bots.ac Course Video Delivery

## Goal

Front your self-hosted course media with Bunny CDN so `cdn.bots.ac` serves MP4s and other course assets globally while your Montreal origin stays protected.

## Decisions baked into this version

- **Pull Zone**, not Storage Zone
- **Volume tier** (`Type: 1`)
- **Hostname origin**, not raw IP
- **Smart Cache enabled**
- **Ignore query strings enabled by default**
- **WebP / AVIF vary removed**
- **SSL scripted**
- **Correct single-URL purge endpoint**
- **Origin Shield deferred until you explicitly choose the shield location**
- **Range-request and MP4 faststart verification included before cutover**

---

## 1. Prerequisites

You need:

- A Bunny account and API key
- A hostname for your origin server, for example `https://origin.bots.ac`
- DNS control for `bots.ac`
- `curl`
- `jq`
- `dig` or `nslookup`
- `python3`
- Optional: `ffmpeg` if you want to repair MP4 faststart in-place

Use a **dedicated video/static hostname or path** behind this zone. Do not treat this as your generic LMS app CDN. Keep it focused on MP4s, downloadable assets, subtitles, images, and course bundles.

---

## 2. Environment file

Create `bunny-env.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail

export BUNNY_API_KEY="YOUR_BUNNY_API_KEY"
export PULL_ZONE_NAME="bots-ac-courses"
export ORIGIN_URL="https://origin.bots.ac"   # change this to your real origin hostname
export CUSTOM_HOSTNAME="cdn.bots.ac"

# used for verification and purge examples
export TEST_VIDEO_PATH="/courses/example/lesson1.mp4"
```

Load it:

```bash
chmod +x bunny-env.sh
source ./bunny-env.sh
```

---

## 3. Create the Pull Zone

Create `create-pull-zone.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
source ./bunny-env.sh

command -v jq >/dev/null 2>&1 || {
  echo "jq is required" >&2
  exit 1
}

payload="$({
  jq -n \
    --arg name "$PULL_ZONE_NAME" \
    --arg origin "$ORIGIN_URL" \
    '{
      Name: $name,
      OriginUrl: $origin,
      Type: 1,
      EnableCacheSlice: true,
      EnableSmartCache: true,
      IgnoreQueryStrings: true,
      CacheControlMaxAgeOverride: 31536000,
      CacheControlPublicMaxAgeOverride: 31536000,
      UseStaleWhileUpdating: true,
      UseStaleWhileOffline: true,
      UseBackgroundUpdate: true,
      OriginRetries: 3,
      VerifyOriginSSL: true,
      EnableAutoSSL: true
    }'
})"

response="$(curl --fail-with-body -sS -X POST "https://api.bunny.net/pullzone" \
  -H "AccessKey: $BUNNY_API_KEY" \
  -H "Content-Type: application/json" \
  -d "$payload")"

echo "$response" | jq .

pull_zone_id="$(echo "$response" | jq -r '.Id')"
if [[ -z "$pull_zone_id" || "$pull_zone_id" == "null" ]]; then
  echo "failed to parse Pull Zone ID" >&2
  exit 1
fi

echo "$pull_zone_id" > .pull_zone_id

echo

echo "Saved Pull Zone ID to .pull_zone_id"

echo "Fetching CNAME target..."

zone="$(curl --fail-with-body -sS "https://api.bunny.net/pullzone/${pull_zone_id}" \
  -H "AccessKey: $BUNNY_API_KEY")"

echo "$zone" | jq '{Id, Name, OriginUrl, CnameDomain, Hostnames}'

echo "$(echo "$zone" | jq -r '.CnameDomain')" > .pull_zone_cname

echo

echo "Next: run add-hostname.sh"
```

Run it:

```bash
chmod +x create-pull-zone.sh
./create-pull-zone.sh
```

---

## 4. Add `cdn.bots.ac` as the custom hostname

Create `add-hostname.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
source ./bunny-env.sh

[[ -f .pull_zone_id ]] || { echo ".pull_zone_id not found" >&2; exit 1; }
pull_zone_id="$(cat .pull_zone_id)"

curl --fail-with-body -sS -X POST "https://api.bunny.net/pullzone/${pull_zone_id}/addHostname" \
  -H "AccessKey: $BUNNY_API_KEY" \
  -H "Content-Type: application/json" \
  -d "{\"Hostname\":\"${CUSTOM_HOSTNAME}\"}"

echo

echo "Hostname added: ${CUSTOM_HOSTNAME}"

zone="$(curl --fail-with-body -sS "https://api.bunny.net/pullzone/${pull_zone_id}" \
  -H "AccessKey: $BUNNY_API_KEY")"

cname_target="$(echo "$zone" | jq -r '.CnameDomain')"

echo "Create this DNS record:"
echo "  ${CUSTOM_HOSTNAME} CNAME ${cname_target}"
```

Run it:

```bash
chmod +x add-hostname.sh
./add-hostname.sh
```

---

## 5. Wait for DNS, then load free SSL and force HTTPS

Create `enable-ssl.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
source ./bunny-env.sh

[[ -f .pull_zone_id ]] || { echo ".pull_zone_id not found" >&2; exit 1; }
pull_zone_id="$(cat .pull_zone_id)"

resolved_cname="$(dig +short CNAME "$CUSTOM_HOSTNAME" | sed 's/\.$//')"
if [[ -z "$resolved_cname" ]]; then
  echo "No CNAME visible yet for ${CUSTOM_HOSTNAME}" >&2
  echo "Wait for DNS propagation, then rerun." >&2
  exit 1
fi

echo "DNS CNAME currently resolves to: $resolved_cname"

echo "Loading free certificate..."

curl --fail-with-body -sS --request GET \
  --url "https://api.bunny.net/pullzone/loadFreeCertificate?hostname=${CUSTOM_HOSTNAME}&useOnlyHttp01=true" \
  --header "AccessKey: $BUNNY_API_KEY"

echo

echo "Enabling Force SSL..."

curl --fail-with-body -sS --request POST \
  --url "https://api.bunny.net/pullzone/${pull_zone_id}/setForceSSL" \
  --header "AccessKey: $BUNNY_API_KEY" \
  --header "Content-Type: application/json" \
  --data "{\"Hostname\":\"${CUSTOM_HOSTNAME}\",\"ForceSSL\":true}"

echo

echo "SSL requested and Force SSL enabled for ${CUSTOM_HOSTNAME}"
```

Run it after the CNAME is live:

```bash
chmod +x enable-ssl.sh
./enable-ssl.sh
```

---

## 6. Purge commands

### Purge the whole Pull Zone

```bash
source ./bunny-env.sh
PULL_ZONE_ID="$(cat .pull_zone_id)"

curl --fail-with-body -sS -X POST "https://api.bunny.net/pullzone/${PULL_ZONE_ID}/purgeCache" \
  -H "AccessKey: $BUNNY_API_KEY"
```

### Purge by tag

If your origin emits a `CDN-Tag` header, you can purge groups of files without clearing the whole zone:

```bash
source ./bunny-env.sh
PULL_ZONE_ID="$(cat .pull_zone_id)"

curl --fail-with-body -sS -X POST "https://api.bunny.net/pullzone/${PULL_ZONE_ID}/purgeCache" \
  -H "AccessKey: $BUNNY_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"CacheTag":"course-docker-101"}'
```

### Purge a single URL

```bash
source ./bunny-env.sh
VIDEO_URL="https://${CUSTOM_HOSTNAME}${TEST_VIDEO_PATH}"
ENCODED_URL="$(printf '%s' "$VIDEO_URL" | jq -sRr @uri)"

curl --fail-with-body -sS --request POST \
  --url "https://api.bunny.net/purge?url=${ENCODED_URL}" \
  --header "AccessKey: $BUNNY_API_KEY"
```

---

## 7. Verify origin range support before sending production traffic

Create `verify-origin-video.sh`:

```bash
#!/usr/bin/env bash
set -euo pipefail
source ./bunny-env.sh

origin_video_url="${ORIGIN_URL%/}${TEST_VIDEO_PATH}"
cdn_video_url="https://${CUSTOM_HOSTNAME}${TEST_VIDEO_PATH}"

echo "== ORIGIN HEAD =="
curl -sSI "$origin_video_url"

echo
echo "== ORIGIN RANGE REQUEST =="
curl -sSI -H 'Range: bytes=0-1023' "$origin_video_url"

echo
echo "Expected on origin:"
echo "- Accept-Ranges: bytes"
echo "- HTTP 206 Partial Content for the range request"

echo
echo "== CDN HEAD =="
curl -sSI "$cdn_video_url" || true

echo
echo "== CDN RANGE REQUEST =="
curl -sSI -H 'Range: bytes=0-1023' "$cdn_video_url" || true
```

Run it:

```bash
chmod +x verify-origin-video.sh
./verify-origin-video.sh
```

If the origin does not return `Accept-Ranges: bytes` and `206 Partial Content`, fix that before cutover.

---

## 8. Verify MP4 faststart / web optimization

Bunny cache slicing solves skip-ahead behavior for uncached content, but badly packaged MP4s can still fail random access if the MP4 metadata lives at the end of the file. Test your source files before cutover.

Create `check-mp4-faststart.py`:

```python
#!/usr/bin/env python3
import mmap
import os
import sys

if len(sys.argv) != 2:
    print("usage: check-mp4-faststart.py /path/to/video.mp4", file=sys.stderr)
    sys.exit(2)

path = sys.argv[1]
size = os.path.getsize(path)

with open(path, "rb") as f, mmap.mmap(f.fileno(), 0, access=mmap.ACCESS_READ) as mm:
    idx = mm.find(b"moov")

if idx == -1:
    print("FAIL: moov atom not found")
    sys.exit(1)

pct = (idx / size) * 100 if size else 0
print(f"moov atom byte offset: {idx}")
print(f"file size: {size}")
print(f"moov position: {pct:.2f}% into file")

if idx <= 2 * 1024 * 1024:
    print("OK: moov atom is near the front of the file")
    sys.exit(0)

print("WARN: moov atom is not near the front; file may not be web optimized")
sys.exit(3)
```

Run it on a representative MP4 file:

```bash
python3 ./check-mp4-faststart.py /path/to/lesson1.mp4
```

If it warns, repair the file without re-encoding:

```bash
ffmpeg -i lesson1.mp4 -c copy -movflags +faststart lesson1.faststart.mp4
```

---

## 9. Optional: enable Origin Shield after you explicitly choose the location

For a Montreal origin, **Chicago** is the sensible first choice because Bunny recommends selecting the Origin Shield location closest to the origin. Bunny’s support docs currently list **Chicago** and **Paris** as the available Origin Shield locations.

However, the official docs pages reviewed here expose the `OriginShieldZoneCode` field without publishing the code strings for those locations, and I did not verify a safe minimal API update body for changing this setting after zone creation. Do **not** guess here.

Recommended approach:

1. Create the zone, hostname, SSL, and video verification first.
2. In Bunny, explicitly pick **Chicago** as the shield location for the Montreal origin.
3. After that is confirmed in the dashboard or official API schema, add a dedicated update script for `EnableOriginShield` and `OriginShieldZoneCode`.

This keeps the main setup path fully verified and avoids a brittle post-create update.

---

## 10. Post-setup checklist

- `cdn.bots.ac` resolves via CNAME to Bunny’s target
- Free SSL loads successfully
- `https://cdn.bots.ac/...` returns valid content
- Origin responds to range requests
- Skip-ahead works on uncached and cached MP4s
- Representative MP4 files are faststart/web-optimized
- Purge full-zone works
- Single-URL purge works
- LMS links now point to `https://cdn.bots.ac/...`
- Only video/static paths are routed through this zone

---

## 11. What this does not change

- It does **not** move your storage off the Montreal server
- It does **not** convert MP4 delivery to HLS/DASH
- It does **not** require Bunny Storage or Bunny Stream

If you later want object storage offload, signed delivery, or HLS packaging, that is a separate architecture step.
