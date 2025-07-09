# DERO Docker Usage Guide

This guide explains how to build and use the Docker images for all DERO project components:
- `derod` (daemon/node)
- `dero-wallet-cli` (wallet CLI)
- `dero-miner` (miner)
- `dero-explorer` (blockchain explorer)
- `dero-simulator` (test/simulation node)

---

## Prerequisites
- [Docker](https://docs.docker.com/get-docker/) installed
- (Optional) [WSL2](https://docs.microsoft.com/en-us/windows/wsl/) for Windows users

---

## Building Docker Images

### Script Tag and Docker Hub Arguments
The build script accepts an **optional tag argument** and an **optional Docker Hub username**. If omitted, the default tag is `local` and images are not pushed.

**Examples:**
```sh
# Build with default tag 'local'
./build_docker_images.sh

# Build with a custom tag (e.g., 'v1.2.3')
./build_docker_images.sh v1.2.3

# Build, tag, and push to Docker Hub (e.g., user 'myuser')
./build_docker_images.sh v1.2.3 myuser
```
This will produce images like:
- `derod:local` or `derod:v1.2.3` or `myuser/derod:v1.2.3`
- `dero-wallet-cli:local` or `dero-wallet-cli:v1.2.3` or `myuser/dero-wallet-cli:v1.2.3`
- etc.

If a Docker Hub username is provided, the script will tag and push each image to Docker Hub as `<dockerhub-username>/<image>:<tag>`.

**Note:**
If you plan to push images to Docker Hub, make sure you are logged in first:

```sh
docker login
```
You will be prompted for your Docker Hub username and password. This is required before you can push images to your Docker Hub account.

From the project root, run:
```sh
./build_docker_images.sh [tag] [dockerhub-username]
```

---

## Running the Containers

### 1. derod (Node/Daemon)
**Ephemeral:**
```sh
docker run --rm -it derod:local --rpc-bind=0.0.0.0:9999 --data-dir=/data
```
**With Persistent Data:**
```sh
# Linux/macOS
docker run --rm -it -v ~/dero_data:/data -p 9999:9999 -p 18089:18089 -p 10100:10100 derod:local --rpc-bind=0.0.0.0:9999 --data-dir=/data
# Windows (PowerShell)
docker run --rm -it -v C:\Users\YourName\dero_data:/data -p 9999:9999 -p 18089:18089 -p 10100:10100 derod:local --rpc-bind=0.0.0.0:9999 --data-dir=/data
```

### 2. dero-wallet-cli (Wallet CLI)
**Ephemeral Wallet:**
```sh
docker run --rm -it dero-wallet-cli:local --wallet-file /tmp/mywallet.db --password mypass
```
**With Persistent Wallet File:**
```sh
# Linux/macOS
docker run --rm -it -v ~/dero_wallets:/wallet dero-wallet-cli:local --wallet-file /wallet/mywallet.db --password mypass
# Windows (PowerShell)
docker run --rm -it -v C:\Users\YourName\dero_wallets:/wallet dero-wallet-cli:local --wallet-file /wallet/mywallet.db --password mypass
```

### 3. dero-miner (Miner)
```sh
docker run --rm -it dero-miner:local --wallet-address=YOUR_DERO_ADDRESS --daemon-rpc-address=NODE_IP:10100
```

### 4. dero-explorer (Explorer)
```sh
docker run --rm -it -p 8081:8081 dero-explorer:local --http-address=0.0.0.0:8081
```

### 5. dero-simulator (Test Node)
```sh
docker run --rm -it dero-simulator:local
```

---

## Volumes and Persistence
- Use `-v /host/path:/container/path` to persist data (blockchain, wallet, logs).
- For wallet files and node data, always use a volume for real use.

---

## Best Practices
- Use volumes for persistent data.
- Use strong passwords for wallet files.
- Back up your wallet files regularly.
- Expose only necessary ports.

---

## Troubleshooting
- **Data lost after restart?** Use volumes!
- **Can't connect to node?** Check port mappings and firewall.
- **Permission errors?** Ensure your user has access to the host directory.
- **Windows path issues?** Use double backslashes or forward slashes in PowerShell.

---

*This guide helps you get started with DERO Docker images quickly and safely.* 