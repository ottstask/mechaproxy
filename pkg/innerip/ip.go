package innerip

import "net"

var ip string

func getInnerIP() string {
	conn, err := net.Dial("udp", "10.0.0.1:80")
	if err != nil {
		return "0.0.0.0"
	}
	defer conn.Close()

	addr := conn.LocalAddr().(*net.UDPAddr)
	if addr == nil {
		return "127.0.0.1"
	}
	return addr.IP.String()
}

// Get return inner ip
func Get() string {
	if ip != "" {
		return ip
	}
	ip = getInnerIP()
	return ip
}
