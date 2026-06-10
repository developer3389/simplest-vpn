> [!CAUTION]
> **This is not an OpenVPN, WireGuard, or HTTP-proxy implementation.**  
> This code is provided as a proof of concept (PoC) for educational purposes.  
> **NO ENCRYPTION, AUTHENTICATION, OR SECURITY MEASURES** ARE IMPLEMENTED.  
> Use at your own risk.

### Project Overview: VPN Complexity Deconstructed
The goal of this repository is to demonstrate that a functional VPN is, at its core, a simple mechanism.  
The entire project consists of roughly 200 unique lines of code, structured symmetrically: the two halves of the system share the same core logic.  
The fundamental operation—shuttling packets between sockets—requires less than 20 lines.  
The rest is standard boilerplate for configuration and routing.

Modern AI tools have changed the game: any student or developer can now build a custom tunneling protocol. The old myth that you need a decade of C++ experience to create a robust VPN is simply false. Today, languages like Go combined with smart coding assistance are more than enough.

We have built a simple prototype to prove this works. We kept it minimal—without encryption or authentication—to focus purely on the core mechanism: routing traffic through a TUN interface. In this basic form, it’s fast: it hits over 500 Mbps on mid-range hardware and up to 1 Gbps on high-end systems.

> [!NOTE]
> Regarding Performance Metrics:  
> We have intentionally omitted formal benchmark reports, charts, or screenshots.  
> Our objective is for the community to verify these findings independently rather than relying on our claims.   
> This ensures the data remains objective and allows users to measure performance within their own network environments, where results will naturally vary based on hardware, network path, and configuration.

### How to run on Linux:

**Install Golang:**
- Install: ```sudo apt update && sudo apt install golang -y```
- Verify installation: `go version`

**Configure:**
Open `server.go` and `client.go` and adjust the network constants (IP, Port) to match your environment.

**Setup Server:**
   ```bash
mkdir server && cd server
# Copy server.go into this directory
go mod init server && go mod tidy
go build server.go
sudo chmod +x ./server
sudo ./server
```

**Setup Client:**
   ```bash
mkdir client && cd client
# Copy client.go into this directory
go mod init client && go mod tidy
go build client.go
sudo chmod +x ./client
sudo ./client
```

> [!TIP]
> If you are in a region with strict network policies, conduct experiments only in a clean, isolated environment (e.g., a separate VM).  **Avoid using the regular internet**, as raw, unmasked traffic patterns are immediately identifiable and may compromise your anonymity.

### Traffic Lifecycle

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

### Resources & Legal
> [!NOTE]
> Background research is available in another pinned repository.

> [!NOTE]
> This project is in the public domain.  
> Feel free to do whatever you want.  
> Attribution is not required.
