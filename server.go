// THIS CODE IS PROVIDED AS A PROOF OF CONCEPT (POC) FOR EDUCATIONAL PURPOSES.
// NO ENCRYPTION, AUTHENTICATION, OR SECURITY MEASURES ARE IMPLEMENTED. USE AT YOUR OWN RISK.
// THIS PROJECT IS IN THE PUBLIC DOMAIN; FEEL FREE TO DO WHATEVER YOU WANT.

// SERVER

package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"sync/atomic"
	"syscall"

	"github.com/songgao/water"
)

const (
	serverPort = 8080 // server port for listening to clients
)

// mainIface
const (
	mainIfaceName = "eth0" // Your main interface name, e.g., "eth0" or "wlan0"
	mainIfaceMtu  = 1500   // MTU of the main interface
)

// tun interface
const (
	tunName = "tun0"        // name of the tun interface, change if needed
	tunIp   = "10.0.0.1/24" // current client IP in tun network, change if needed
)

// special constants, do not chnge without understanding
const (
	tunMTU             = mainIfaceMtu - 40 // 40 bytes for IP and UDP headers, to avoid fragmentation
	tunMSS             = tunMTU - 40       // 40 bytes for IP and UDP headers from tunMTU to prevent fragmentation
	kernelBufferKbytes = 32 * 1024         // 32 MB for kernel buffers
	packetBufSize      = 2048              // buffer size for reading/writing packets, should be >= tunMTU + headers
	tunQueueSize       = 10_000            // tx queue length for tun interface, adjust based on expected RTT
)

// saving last client address for sending packets back to the last client
// Many clients support is not implemented in this POC
var lastClientAddr atomic.Pointer[net.UDPAddr]

func main() {

	preUp()

	device, err := createTun()
	if err != nil {
		log.Fatalf("Error creating TUN device: %v", err)
	}

	waitForExit := handleCtrlC()

	onUp()

	conn, _ := net.ListenUDP("udp", &net.UDPAddr{Port: serverPort})

	conn.SetReadBuffer(kernelBufferKbytes * 1024)
	conn.SetWriteBuffer(kernelBufferKbytes * 1024)

	log.Printf("VPN listening on port %d", serverPort)

	go fromTun(device, conn)

	go toTun(device, conn)

	<-waitForExit

	preDown()
	device.Close()

	log.Println("VPN stopped")
}

func fromTun(device *water.Interface, conn *net.UDPConn) {
	buf := make([]byte, packetBufSize)

	for {
		packetSize, err := device.Read(buf)
		if err != nil {
			log.Printf("TUN Read error: %v", err)
			continue
		}

		lastClientAddr := lastClientAddr.Load()

		if lastClientAddr != nil {
			_, err := conn.WriteToUDP(buf[0:packetSize], lastClientAddr)
			if err != nil {
				log.Printf("UDP Write error: %v", err)
			}
		}
	}
}

func toTun(device *water.Interface, conn *net.UDPConn) {
	buf := make([]byte, packetBufSize)

	for {
		packetSize, clientAddr, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Printf("UDP Read error: %v", err)
			continue
		}

		lastClientAddr.Store(clientAddr)

		_, err = device.Write(buf[0:packetSize])
		if err != nil {
			log.Printf("TUN Write error: %v", err)
		}
	}
}

func createTun() (*water.Interface, error) {
	config := water.Config{
		DeviceType: water.TUN,
		PlatformSpecificParams: water.PlatformSpecificParams{
			Name: tunName,
		},
	}

	device, err := water.New(config)
	if err != nil {
		log.Fatalf("Error creating TUN device: %v", err)
	}
	return device, nil
}

func preUp() {

	runCmd("sysctl -w net.ipv4.ip_forward=1")
	runCmd("sysctl -w net.core.rmem_max=%d", kernelBufferKbytes*1024)
	runCmd("sysctl -w net.core.wmem_max=%d", kernelBufferKbytes*1024)

	// Uncomment and modify the following line if you need to perform a factory reset
	// of the firewall rules and set default policies to ACCEPT.
	// WARNING: This will clear all existing rules, including those for other services!

	//runCmd("iptables -F")
	//runCmd("iptables -t nat -F")
	//runCmd("iptables -t mangle -F")
	//runCmd("iptables -X")
	//runCmd("iptables -P INPUT ACCEPT")
	//runCmd("iptables -P FORWARD ACCEPT")
	//runCmd("iptables -P OUTPUT ACCEPT")

	// Optionally, you can set TCP congestion control and qdisc for better performance

	//runCmd("sysctl -w net.ipv4.tcp_congestion_control=bbr")
	//runCmd("sysctl -w net.core.default_qdisc=cake")
}

func onUp() {
	runCmd("ip addr add %s dev %s", tunIp, tunName)
	runCmd("ip link set dev %s mtu %d up", tunName, tunMTU)
	runCmd("ip link set %s txqueuelen %d", tunName, tunQueueSize)

	// Uncomment the following line if you want to set the tx queue length of the main interface as well.

	//runCmd("ip link set %s txqueuelen %d", mainIfaceName, tunQueueSize)

	runCmd("iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", mainIfaceName)

	runCmd("iptables -A FORWARD -i %s -o %s -j ACCEPT", tunName, mainIfaceName)
	runCmd("iptables -A FORWARD -i %s -o %s -m state --state RELATED,ESTABLISHED -j ACCEPT", mainIfaceName, tunName)

	runCmd("ethtool -K %s gso off gro off tso off tx-udp-segmentation off", tunName)

	runCmd("iptables -t mangle -A FORWARD -p tcp --tcp-flags SYN,RST SYN -j TCPMSS --set-mss %d", tunMSS)
}

func preDown() {
	runCmd("iptables -t nat -D POSTROUTING -o %s -j MASQUERADE", mainIfaceName)
	runCmd("ip link set dev %s down", tunName)
}

func runCmd(cmdStr string, args ...any) {
	cmd := fmt.Sprintf(cmdStr, args...)
	if cmd == "" {
		return
	}

	log.Printf("[EXEC] %s", cmd)
	cmdResult := exec.Command("sh", "-c", cmd)
	if output, err := cmdResult.CombinedOutput(); err != nil {
		log.Printf("[WARN] Error: %v. Output: %s", err, string(output))
	}
}

func handleCtrlC() chan os.Signal {
	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGTERM)
	return exit
}
