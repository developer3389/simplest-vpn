> [!CAUTION]
> **This is not an OpenVPN, WireGuard, or HTTP-proxy implementation.**  
> This code is provided as a proof of concept (PoC) for educational purposes.  
> **NO ENCRYPTION, AUTHENTICATION, OR SECURITY MEASURES** ARE IMPLEMENTED.  
> Use at your own risk.

## Project Overview: VPN Complexity Deconstructed
The goal of this repository is to demonstrate that a functional VPN is, at its core, a simple mechanism.  
The entire project consists of roughly 200 unique lines of code, structured symmetrically: the two halves of the system share the same core logic.  
The fundamental operation - shuttling packets between sockets - requires less than 20 lines.  
The rest is standard boilerplate for configuration and routing.

Modern AI tools have changed the game: any student or developer can now build a custom tunneling protocol. The old myth that you need a decade of C++ experience to create a robust VPN is simply false. Today, languages like Go combined with smart coding assistance are more than enough.

We have built a simple prototype to prove this works. We kept it minimal - without encryption or authentication - to focus purely on the core mechanism: routing traffic through a TUN interface. In this basic form, it’s fast: it hits over 500 Mbps on mid-range hardware and up to 1 Gbps on high-end systems.

> [!NOTE]  
> Regarding Performance Metrics:  
> We have intentionally omitted formal benchmark reports, charts, or screenshots.  
> Our objective is for the community to verify these findings independently rather than relying on our claims.   
> This ensures the data remains objective and allows users to measure performance within their own network environments, where results will naturally vary based on hardware, network path, and configuration.

# Deployment Guide

## 1. Setup Hardware
If you have two Linux-based VPS, that's perfect.  
If you have two Linux-based hardware devices, that's perfect.  
If you have one Linux-based hardware device and one Linux-based VPS, that's perfect.


> [!TIP]  
> If you are in a region with strict network policies, conduct experiments only in a clean, isolated environment (e.g., a separate VM).  **Avoid using the regular internet**, as raw, unmasked traffic patterns are immediately identifiable and may compromise your privacy.

Otherwise, use virtualization:

<details>
<summary>Setup Emulated Hardware</summary>

### 1.1 Setup Emulated Hardware  

1. Download the Debian virtual machine ISO-image from here for 64-bit AMD PC: [link](https://cdimage.debian.org/debian-cd/current/amd64/iso-cd/debian-13.5.0-amd64-netinst.iso)

2. Using a hypervisor like Hyper-V, VirtualBox, or VMware, create two Debian virtual machines.  
Allocate `at least 2 GB` of RAM to each and configure the network adapter to receive an IP address from your physical router.

3. Once both VMs are installed, they should obtain IP addresses from your router.  
Pin these addresses via DHCP by assigning them as static leases in your router's interface.

4. Take a real home network with a physical router.    
Let's assume the router gateway is `192.168.0.1`.  
Run `ip a` command.  
Look for the `eth0` or `enp0s3` interface; it should have an IP address within the `192.168.0.x` range.  
Check ping: `ping 8.8.8.8`  
You should see an output similar to this:
```bash
PING 8.8.8.8 (8.8.8.8) xx(xx) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=x ttl=xxx time=xx.x ms
64 bytes from 8.8.8.8: icmp_seq=x ttl=xxx time=xx.x ms
```
</details>

### 1.2. Environment Setup
```bash
# Install prerequisites for networking and building
sudo apt update && \
sudo apt install -y \
    sudo \
    iproute2 \
    iptables \
    ethtool \
    procps \
    git \
    nano \
    iperf3
```

```bash
# Setup Speedtest OOKLA
sudo apt-get install curl
curl -s https://packagecloud.io/install/repositories/ookla/speedtest-cli/script.deb.sh | sudo bash
sudo apt-get install -y speedtest
```

### 1.3. Install Golang:
- Install: ```sudo apt update && sudo apt install golang -y```
- Verify installation: `go version`

You should see an output similar to this: `go version go1.2x.x linux/amd64`

## 2. Clone this repository:
```git clone https://github.com/developer3389/simplest-vpn.git```
> [!TIP]   
> Run this command *right now* on both the client and the server.

## 3. Firstly, Setup Server:

```bash
# Create directory and go into it
mkdir server && cd server
```

```bash
# Copy server.go from the cloned repository to the current directory
cp ../simplest-vpn/server.go .
```

```bash
# Init golang environment
go mod init server && go mod tidy
```

```bash
# Configure `server.go` in editor.
# Set the gateway for your environment.
nano server.go
# Save: ctrl+X; Y
# Exit without saving: ctrl+X; N
```

```bash
# Build executable file
go build
```

```bash
# Ensure system binaries are found (vital for iptables/ip/ethtool)
export PATH=$PATH:/usr/local/sbin:/usr/sbin:/sbin
```

```bash
# Set executable rights
# TUN interface needs root
sudo chmod +x ./server && sudo ./server
```

You should see `VPN listening on port 8080` string.

## 4. Then, Setup Client:

```bash
# Create directory and go into it
mkdir client && cd client
```

```bash
# Copy client.go from the cloned repository to the current directory
cp ../simplest-vpn/client.go .
```

```bash
# Init golang environment
go mod init client && go mod tidy
```

```bash
# Configure `client.go` in editor.
# Set the IP, port and gateway for your environment.
nano client.go
# Save: ctrl+X; Y
# Exit without saving: ctrl+X; N
```

```bash
# Build executable file
go build
```

```bash
# Ensure system binaries are found (vital for iptables/ip/ethtool)
export PATH=$PATH:/usr/local/sbin:/usr/sbin:/sbin
```

```bash
# Set executable rights
# TUN interface needs root
sudo chmod +x ./client && sudo ./client
```

You should see `VPN started on port 8080` string.

## 5. Testing
- On server: `iperf3 -s`
- On client without VPN: `iperf3 -c 192.168.0.x -t 10`
- On client via VPN: `iperf3 -c 10.0.0.1 -t 10`

You should see an output similar to this:
```
Connecting to host 10.0.0.1, port 5201
[  5] local 10.0.0.2 port 58352 connected to 10.0.0.1 port 5201
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec   xxx MBytes  1.08 Gbits/sec  xxx    xxx KBytes       
[  5]   1.00-2.00   sec   xxx MBytes  1.05 Gbits/sec  xxx    xxx KBytes       
[  5]   2.00-3.00   sec   xxx MBytes  1.08 Gbits/sec  xxx    xxx KBytes       
```

Now, stop the client by pressing `CTRL-C`

Block internet access for the client in your router settings.

Ensure that pinging `8.8.8.8` produces no output.

Reconnect the VPN by running: `./client`  

Check ping: `ping 8.8.8.8`

You should see an output similar to this:
```bash
PING 8.8.8.8 (8.8.8.8) xx(xx) bytes of data.
64 bytes from 8.8.8.8: icmp_seq=x ttl=xxx time=xx.x ms
64 bytes from 8.8.8.8: icmp_seq=x ttl=xxx time=xx.x ms
```

Now, run the speedtest: `speedtest`  
You should see an output similar to this:
```bash
   Download:   52.33 Mbps
   Upload:    92.57 Mbps
   Packet Loss:     0.0%
```

🟢🟢🟢 Congratulations! Your custom VPN protocol is now up and running! 🟢🟢🟢

## Traffic Lifecycle
*This is optional and provided only for better understanding of the underlying mechanics.*
<details>
<summary>View Technical Details</summary>

Both client and server instances employ two dedicated goroutines to handle bidirectional traffic flow, ensuring non-blocking, concurrent packet processing.

**CLIENT:**
- **Outbound:** `User App` -> `Kernel (plain)` -> `Golang (receives plain & encapsulates to IP/UDP)` -> `Kernel (enc.)` -> `Internet (to server)`
- **Inbound:** `Internet (from server)` -> `Kernel (enc.)` -> `Golang (receives encapsulated IP/UDP & decapsulates to plain)` -> `Kernel (plain)` -> `User App`

**SERVER:**
- **To Internet:** `Internet (from client)` -> `Kernel (enc.)` -> `Golang (receives encapsulated IP/UDP & decapsulates to plain)` -> `Kernel (plain)` -> `Internet (to dest)`
- **To Client:** `Internet (from dest)` -> `Kernel (plain)` -> `Golang (receives plain & encapsulates to IP/UDP)` -> `Kernel (enc.)` -> `Internet (to client)`

> [!NOTE]
> - (plain) = original raw traffic  
> - (enc.) = encapsulated traffic (IP-in-UDP)  
</details>

## Resources & Legal
> [!NOTE]
> For a detailed analysis of modern infrastructure challenges and the design philosophy of this project, see [another repository](https://github.com/developer3389/network-censorship-analysis).

> [!NOTE]
> This project is in the public domain.  
> Feel free to do whatever you want.  
> Attribution is not required.
