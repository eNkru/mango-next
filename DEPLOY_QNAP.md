# 🚀 Deploying Mango to a QNAP NAS

This guide covers two ways to deploy Mango (with your local changes) on a QNAP NAS using **Container Station** (Docker):

| Method | Best for | NAS requirements |
|---|---|---|
| **A: Prebuild locally, export image** ⭐ | NAS with limited CPU/RAM, or fastest workflow | Docker daemon only |
| **B: Build on the NAS from source** | NAS with decent CPU/RAM (≥2 GB), want single-command deploys | Docker + source code on NAS |

---

## 📋 Prerequisites

| Requirement | Details |
|---|---|
| **QNAP NAS** | Any model running QTS 4.3+ or QuTS hero |
| **Container Station** | Install from the QNAP App Center (QTS → App Center → search "Container Station") |
| **SSH access** | Control Panel → Network & File Services → Telnet/SSH → Enable SSH |
| **Your manga files** | `.cbz`, `.cbr`, `.zip`, or `.rar` archives stored on the NAS |
| **Docker on dev machine** | Required for Method A only |

---

## Step 1 — Prepare Share Folders on the NAS

SSH into your QNAP and create the directories Mango needs:

```bash
ssh admin@<NAS-IP>

# Create directories for config, data, and compose file
mkdir -p /share/Container/mango/config
mkdir -p /share/Container/mango/data
mkdir -p /share/Container/mango
```

> If your manga files are already in a share like `/share/Multimedia/Manga`, you can mount that directly — no need to copy files.

---

## Step 2 — Deploy (Choose Method A or B)

---

### ⭐ Method A: Prebuild Locally, Export Image to NAS (Recommended)

Build the Docker image on your dev machine (which has plenty of RAM/CPU), then transfer it. **No build tools or source code needed on the NAS.**

#### 2A-1. Build the image locally

```bash
# On your dev machine (macOS/Linux):
cd /path/to/mango-next

# Build for the NAS architecture.
# Most QNAP NAS are linux/amd64.
# If your Mac is Apple Silicon (M1/M2/M3/M4), you MUST specify the platform:
docker build --platform linux/amd64 -t mango_cn:local .

# If your dev machine is already x86_64 Linux, platform flag is optional:
# docker build -t mango_cn:local .
```

> ⏱️ The build compiles the Go binary — usually under a few minutes on a modern machine.

#### 2A-2. Export the image to a compressed tar

```bash
# On your dev machine:
docker save mango_cn:local | gzip > mango_cn-local.tar.gz
```

Typical size: ~50–100 MB compressed.

#### 2A-3. Transfer the image to the NAS

```bash
# On your dev machine:
scp mango_cn-local.tar.gz admin@<NAS-IP>:/share/Container/
```

#### 2A-4. Load the image on the NAS

```bash
# SSH into your QNAP:
ssh admin@<NAS-IP>
docker load < /share/Container/mango_cn-local.tar.gz

# Verify the image loaded:
docker images | grep mango_cn
```

#### 2A-5. Create the compose file and start

Copy `docker-compose.qnap-prebuilt.yml` from this repo to `/share/Container/mango/` on the NAS, **or** create it inline:

```bash
cat > /share/Container/mango/docker-compose.yml << 'EOF'
version: '3.7'

services:
  mango:
    image: mango_cn:local
    container_name: mango
    restart: unless-stopped
    ports:
      - "9000:9000"
    volumes:
      - /share/Multimedia/Manga:/root/mango/library
      - /share/Container/mango/data:/root/mango
      - /share/Container/mango/config:/root/.config/mango
    environment:
      - PORT=9000
      - DB_PATH=/root/mango/mango.db
EOF

cd /share/Container/mango
docker compose up -d
```

#### 2A-6. Cleanup the tar file (optional)

```bash
rm /share/Container/mango_cn-local.tar.gz
```

---

### Method B: Build on the NAS from Source

Transfer source code to the NAS and let Docker build it there. Requires decent CPU/RAM on the NAS.

#### 2B-1. Transfer source code to the NAS

Choose **one** method:

**SCP (simple):**
```bash
# From your dev machine:
scp -r /path/to/mango-next admin@<NAS-IP>:/share/Container/mango-src
```

**Rsync (faster for re-deploys — only transfers changed files):**
```bash
# From your dev machine:
rsync -avz --exclude '.git' --exclude 'mango' \
  /path/to/mango-next/ admin@<NAS-IP>:/share/Container/mango-src/
```

**File Station:** Zip the project, upload via File Station, extract on the NAS via SSH.

> **Tip:** Exclude `.git` and local binaries to save transfer time. Docker only needs the `go/` tree + root `Dockerfile`.

#### 2B-2. Create the compose file and build

Copy `docker-compose.qnap.yml` from this repo to `/share/Container/mango/` on the NAS, **or** create it inline:

```bash
cat > /share/Container/mango/docker-compose.yml << 'EOF'
version: '3.7'

services:
  mango:
    build:
      context: /share/Container/mango-src
      dockerfile: Dockerfile
    image: mango_cn:local
    container_name: mango
    restart: unless-stopped
    ports:
      - "9000:9000"
    volumes:
      - /share/Multimedia/Manga:/root/mango/library
      - /share/Container/mango/data:/root/mango
      - /share/Container/mango/config:/root/.config/mango
    environment:
      - PORT=9000
      - DB_PATH=/root/mango/mango.db
EOF

cd /share/Container/mango
docker compose build    # first run downloads Go toolchain image and compiles
docker compose up -d
```

---

## Step 3 — Verify It's Running

1. **Check the container status:**

   ```bash
   docker ps | grep mango
   ```

   You should see `mango` with status `Up`.

2. **Check logs** (if something seems wrong):

   ```bash
   docker logs mango
   ```

   On first run, Mango will:
   - Auto-generate `config.yml` in `/root/.config/mango/`
   - Scan your manga library at `/root/mango/library`
   - Create a default `admin` user and print the password to the log

3. **Open the web UI:**

   Navigate to `http://<NAS-IP>:9000` in your browser.

4. **Log in** with the default admin credentials (check the log output for the auto-generated password).

---

## 🔄 Updating After Local Changes

### Method A (prebuilt image) — rebuild and re-export

```bash
# === On your dev machine ===

# 1. Rebuild the image with your latest changes
cd /path/to/mango-next
#   Apple Silicon Macs must specify --platform linux/amd64:
docker build --platform linux/amd64 -t mango_cn:local .

# 2. Export and transfer
docker save mango_cn:local | gzip > mango_cn-local.tar.gz
scp mango_cn-local.tar.gz admin@<NAS-IP>:/share/Container/

# === On the QNAP (SSH) ===

# 3. Load the new image and restart
ssh admin@<NAS-IP>
docker load < /share/Container/mango_cn-local.tar.gz
cd /share/Container/mango
docker compose up -d    # Docker detects the updated image and recreates the container

# 4. Cleanup
rm /share/Container/mango_cn-local.tar.gz
```

> Your data (database, config, manga library) is preserved across updates because it lives in mounted volumes.

### Method B (build on NAS) — sync and rebuild

```bash
# 1. Sync updated source to the NAS (from your dev machine)
rsync -avz --exclude '.git' --exclude 'mango' \
  /path/to/mango-next/ admin@<NAS-IP>:/share/Container/mango-src/

# 2. Rebuild and restart on the NAS (SSH into the NAS)
ssh admin@<NAS-IP>
cd /share/Container/mango
docker compose build
docker compose up -d
```

---

## 🔧 Configuration

Mango's config is auto-generated at `/share/Container/mango/config/config.yml`. Edit it to customize:

```yaml
# Key options you may want to change:
host: 0.0.0.0
port: 9000
library_path: /root/mango/library
db_path: /root/mango/mango.db       # ⚠️ Must be inside /root/mango volume! Default is ~/mango.db (outside)
scan_interval_minutes: 5           # How often to re-scan the library
thumbnail_generation_interval_hours: 24
disable_login: false               # Set to true + set default_username to skip login
log_level: info                     # debug, info, warn, error
cache_enabled: true
cache_size_mbs: 50
```

> **After editing config, restart the container:**
> ```bash
> docker restart mango
> ```

All config options can also be overridden via **environment variables** (uppercase name), e.g. `SCAN_INTERVAL_MINUTES=10`. See `go/internal/config/config.go` for the full list.

---

## 🗂️ Adding Manga

Simply drop `.cbz`, `.cbr`, `.zip`, or `.rar` files into your manga share folder. Mango will automatically detect new files on the next library scan (default: every 5 minutes).

Organize with nested folders for titles:

```
/share/Multimedia/Manga/
├── One Piece/
│   ├── Vol 01.cbz
│   ├── Vol 02.cbz
│   └── ...
├── Attack on Titan/
│   ├── Vol 01.cbr
│   └── ...
└── Solo Title.zip
```

---

## 🔒 Permissions Troubleshooting

If Mango can't read your manga files or fails to write its database:

1. **Check share permissions** in QNAP:
   Control Panel → Privilege → Shared Folders → select folder → **Edit** → set Read/Write for the admin user (or "Allow guest access").

2. **Ensure the Docker user can access the path:**
   QNAP's Docker daemon typically runs as `admin`. The share must be accessible to that user.

3. **Avoid spaces in folder names** — they can cause issues with Docker bind-mounts.

4. **Use the correct QNAP volume path** — if your data is on volume 2, the path might be `/share/CACHEDEV2_DATA/...` instead of `/share/Multimedia/...`. Run `ls /share/` via SSH to see available mount points.

---

## 🐛 Common Issues

| Problem | Solution |
|---|---|
| Build fails with "out of memory" (Method B) | NAS has limited RAM. Use **Method A** (prebuild on dev machine) instead. |
| Container won't start | Check `docker logs mango` for errors. Verify volume paths exist. |
| "The config file does not exist" | This is normal on first run — Mango creates it automatically. |
| Library appears empty | Ensure manga files are in the mounted library path. Check file permissions. |
| Can't access web UI | Verify port 9000 isn't blocked by the QNAP firewall. Try `http://<NAS-IP>:9000`. |
| Thumbnails not generating | The container includes all image libraries; just wait for the scan cycle (24h default). |
| Forgot admin password | Delete `/share/Container/mango/data/mango.db` and restart — Mango will recreate it. ⚠️ This resets all users and reading progress. |
| Database lost after update | Ensure `DB_PATH=/root/mango/mango.db` is set — the default `~/mango.db` lives outside the persisted volume. |
| Image won't load on NAS (Method A) | You may have built for the wrong architecture. Rebuild with `--platform linux/amd64`. Verify with `docker inspect mango_cn:local \| grep Architecture`. |
| `docker save` too slow | Use `gzip -1` for faster compression (larger file but much quicker): `docker save mango_cn:local \| gzip -1 > mango_cn-local.tar.gz` |

---

## 📁 File Structure on Your NAS (After Deployment)

```
/share/
├── Container/
│   ├── mango/
│   │   ├── docker-compose.yml       # Your compose file
│   │   ├── config/
│   │   │   └── config.yml           # Mango config (auto-generated)
│   │   └── data/
│   │       ├── mango.db             # User database & reading progress
│   │       ├── queue.db             # Download queue
│   │       ├── library.yml.gz       # Library cache
│   │       ├── uploads/             # Uploaded files
│   │       └── plugins/             # Download plugins
│   └── mango-src/                   # Source code (Method B only — not needed for Method A)
│       ├── Dockerfile
│       ├── Makefile
│       └── go/ ...
└── Multimedia/Manga/               # Your manga library
    ├── Title A/
    │   └── Vol 01.cbz
    └── ...
```

---

## 🎯 Quick-Start Summary

### Method A: Prebuild locally (recommended)

```bash
# === On your dev machine ===
cd /path/to/mango-next
docker build --platform linux/amd64 -t mango_cn:local .
docker save mango_cn:local | gzip > mango_cn-local.tar.gz
scp mango_cn-local.tar.gz admin@<NAS-IP>:/share/Container/

# === On the QNAP (SSH) ===
ssh admin@<NAS-IP>
mkdir -p /share/Container/mango/{config,data}
docker load < /share/Container/mango_cn-local.tar.gz

# Create compose file (see Step 2A-5 above), then:
cd /share/Container/mango
docker compose up -d

# Get admin password from logs
docker logs mango | head -20

# Open in browser → http://<NAS-IP>:9000
```

### Method B: Build on NAS

```bash
# === On your dev machine ===
rsync -avz --exclude '.git' --exclude 'mango' \
  /path/to/mango-next/ admin@<NAS-IP>:/share/Container/mango-src/

# === On the QNAP (SSH) ===
ssh admin@<NAS-IP>
mkdir -p /share/Container/mango/{config,data}

# Create compose file (see Step 2B-2 above), then:
cd /share/Container/mango
docker compose build
docker compose up -d

# Get admin password from logs
docker logs mango | head -20

# Open in browser → http://<NAS-IP>:9000
```

Happy reading! 📚
