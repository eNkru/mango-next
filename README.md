# Mango (Chinese Regional Fork)

![banner-paddings](https://user-images.githubusercontent.com/38988286/199423262-68f03906-5444-499c-8616-aa675039544e.png)

This repository is a customized fork of [getmango/Mango](https://github.com/getmango/Mango). It has been tailored for local deployment in regions with restricted access to public CDNs, optimized for Synology NAS environments, and translated into Chinese. 

Mango is a self-hosted manga server and web-based reader. 

## ✨ Features

- **Multi-user support**: Manage multiple users with customized permissions and preferences.
- **OPDS support**: Connect with external manga readers and applications easily.
- **Theme customization**: Built-in dark/light mode toggle.
- **Support for multiple format files**: Seamlessly handle `.cbz`, `.zip`, `.cbr`, and `.rar` formats.
- **Organized library**: Supports nested folders for structural organization within your library.
- **Smart resume**: Automatically preserves reading progress for each user.
- **Thumbnail generation**: Beautiful grid display with generated cover thumbnails.
- **Plugin system**: Support downloading from third-party websites with plugins.
- **Responsive Web Reader**: Enjoy responsive viewing across all devices (desktop, tablet, mobile), without the need for a dedicated app.
- **Pain-free deployment**: All static assets are embedded within a single binary. 

### 🌟 Fork-Specific Features
- Fully translated to Chinese.
- JS frontend dependencies are localized (removed external CDN reliance for mainland users).
- Fixed Docker file permission limitations in Synology NAS environments.
- Packed easily installable Synology SPK (install directly via Package Center).

---

## 🛠️ How to Develop

Mango is primarily built in [Crystal](https://crystal-lang.org/) and utilizes the Kemal web framework, backed by SQLite3. The frontend utilizes Node.js and Gulp for asset processing.

### Prerequisites
- **Crystal `1.0.0` or higher & Shards** (macOS: `brew install crystal`, Ubuntu/Debian: `sudo snap install crystal --classic`)
- **Node.js & Yarn** (macOS: `brew install node yarn`)
- **Required system libraries** (e.g., `libarchive`, `sqlite3`, and `wget`. macOS: `brew install libarchive sqlite3 wget`)

### Development Steps
1. Clone the repository and navigate into the directory.
2. Initialize and compile the frontend assets:
   ```bash
   make setup
   ```
   (This runs `yarn` install and the gulp development task).
3. To start the application in development mode:
   ```bash
   make run
   ```
4. Testing & linting:
   - Run tests: `make test` (Executes `crystal spec`)
   - Check formatting and linting: `make check` (Executes `crystal tool format` and `ameba`)

---

## 📦 How to Build

Building Mango is simplified by the provided `Makefile`.

1. **Install Shards (Crystal dependencies):**
   ```bash
   make libs
   ```
2. **Build frontend UI assets:**
   ```bash
   make uglify
   ```
3. **Build the standard binary:**
   ```bash
   make build
   ```
   *This compiles `src/mango.cr` and produces the `mango` binary executable.*

**Note:** For a fully standalone statically linked binary, you can run `make static`.
For ARM architectures:
- 32-bit ARM: `make arm32v7`
- 64-bit ARM: `make arm64v8`

---

## 🚀 How to Deploy

The simplest and recommended way to deploy Mango is using Docker or via Synology NAS Manager. All static assets (HTML/CSS/JS) are bundled inside the compiled binary. 

### Using Docker

1. **Build the image**:
   ```bash
   docker build -t mango .
   ```
2. **Run the container**:
   Ensure you mount your manga library containing your archive files to `/root/mango/library` within the container.
   ```bash
   docker run -d \\
     -p 9000:9000 \\
     -v /path/to/your/manga:/root/mango/library \\
     -v /path/to/mango/config:/root/mango/config \\
     --name manga-server \\
     mango
   ```

*Or use the included `docker-compose.yml` for convenience!*

### Synology NAS Direct Installation
Since this fork is heavily optimized for Synology devices, you can install the packaged `.spk` application directly via the Synology Package Center instead of using Docker manually.

---

## 🛠️ Troubleshooting

### 'ameba' failed to compile during `make setup` / `make libs`
If you see an error like `Error: type must be Ameba::Severity, not (Ameba::Severity | Nil)` during dependency installation:
- **Solution**: The `ameba` linter may be incompatible with newer versions of Crystal. This has now been moved to `development_dependencies` in `shard.yml` so that it is ignored during production builds. 
- Ensure you have the updated `shard.yml` from this repository.
- If you're encountering the error from a stale cache, run the following commands to clean the lock file and reinstall:
  ```bash
  shards install
  make setup
  ```
  
## Special thanks
[LINUX DO](https://linux.do/)
