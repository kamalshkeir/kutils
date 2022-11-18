package kutils

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"strings"
	"time"

	"github.com/kamalshkeir/kutils/kstring"
)

// Graceful Shutdown
func GracefulShutdown(f func() error) error {
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	<-s
	return f()
}

// Cronjob like
func RunEvery(t time.Duration, function any) {
	//Usage : go RunEvery(2 * time.Second,func(){})
	fn, ok := function.(func())
	if !ok {
		fmt.Println("ERROR : fn is not a function")
		return
	}

	fn()
	c := time.NewTicker(t)

	for range c.C {
		fn()
	}
}

func RetryEvery(t time.Duration, function func() error, maxRetry ...int) {
	i := 0
	err := function()
	for err != nil {
		time.Sleep(t)
		i++
		if len(maxRetry) > 0 {
			if i < maxRetry[0] {
				err = function()
			} else {
				fmt.Println("stoping retry after", maxRetry, "times")
				break
			}
		} else {
			err = function()
		}
	}
}


func GenerateUUID() string {
	return kstring.GenerateUUID()
}


func PrintMemUsage() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	fmt.Printf("Alloc = %v MiB", bToMb(m.Alloc))
	fmt.Printf("\tTotalAlloc = %v MiB", bToMb(m.TotalAlloc))
	fmt.Printf("\tSys = %v MiB", bToMb(m.Sys))
	fmt.Printf("\tNumGC = %v\n", m.NumGC)
}

func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

func OpenBrowser(url string) {
	var err error
	switch runtime.GOOS {
	case "linux":
		err = exec.Command("xdg-open", url).Start()
	case "windows":
		err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
	case "darwin":
		err = exec.Command("open", url).Start()
	default:
	}
	_ = err
}

func GetIpCountry(ip string) string {
	cmd := exec.Command("nmap", "-n", "-sn", "-Pn", "--script", "ip-geolocation-geoplugin", ip)
	d, err := cmd.Output()
	if err != nil {
		return ""
	}
	bReader := bytes.NewReader(d)
	scanner := bufio.NewScanner(bReader)
	var line string
	for scanner.Scan() {
		if strings.Contains(scanner.Text(), "_location") {
			line = scanner.Text()
			sp := strings.Split(line, "location:")
			line = strings.TrimSpace(sp[len(sp)-1])
			spl := strings.Split(line, ",")
			spl[1] = strings.TrimSpace(spl[1])
			return spl[1]
		}
	}
	return ""
}

func GetPrivateIp() string {
	pIp := getOutboundIP()
	if pIp == "" {
		pIp = resolveHostIp()
		if pIp == "" {
			pIp = getLocalPrivateIps()[0]
		}
	}
	return pIp
}

func resolveHostIp() string {
	netInterfaceAddresses, err := net.InterfaceAddrs()

	if err != nil {
		return ""
	}

	for _, netInterfaceAddress := range netInterfaceAddresses {
		networkIp, ok := netInterfaceAddress.(*net.IPNet)
		if ok && !networkIp.IP.IsLoopback() && networkIp.IP.To4() != nil {
			ip := networkIp.IP.String()
			return ip
		}
	}

	return ""
}

func getLocalPrivateIps() []string {
	ips := []string{}
	host, _ := os.Hostname()
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
		if ipv4 := addr.To4(); ipv4 != nil {
			ips = append(ips, ipv4.String())
		}
	}
	return ips
}

func getOutboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	if localAddr.IP.To4().IsPrivate() {
		return localAddr.IP.String()
	}
	return ""
}