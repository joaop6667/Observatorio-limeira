//go:build !windows

package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
)

// Usado apenas para testar o servidor local fora do Windows.
func main() {
	dir := os.Args[1]
	ln, _ := net.Listen("tcp", "127.0.0.1:0")
	a := &app{dir: dir, token: newToken()}
	fmt.Printf("http://127.0.0.1:%d/?t=%s\n", ln.Addr().(*net.TCPAddr).Port, a.token)
	_ = http.Serve(ln, a)
}
