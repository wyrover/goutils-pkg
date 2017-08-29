package main

import (
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"os"

	"github.com/kisom/goutils/die"
)

func main() {
	cfg := &tls.Config{}

	var sysRoot, listenAddr string
	var verify bool
	flag.StringVar(&sysRoot, "ca", "", "provide an alternate CA bundle")
	flag.StringVar(&listenAddr, "listen", ":443", "address to listen on")
	flag.BoolVar(&verify, "verify", false, "verify client certificates")
	flag.Parse()

	if verify {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}
	if sysRoot != "" {
		pemList, err := ioutil.ReadFile(sysRoot)
		die.If(err)

		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(pemList) {
			fmt.Printf("[!] no valid roots found")
			roots = nil
		}

		cfg.RootCAs = roots
	}

	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println(err.Error())
		}

		raddr := conn.RemoteAddr()
		tconn := tls.Server(conn, cfg)
		cs := tconn.ConnectionState()
		if len(cs.PeerCertificates) == 0 {
			fmt.Printf("[+] %v: no chain presented\n", raddr)
			continue
		}

		var chain []byte
		for _, cert := range cs.PeerCertificates {
			p := &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert.Raw,
			}
			chain = append(chain, pem.EncodeToMemory(p)...)
		}

		var nonce [16]byte
		_, err = rand.Read(nonce[:])
		if err != nil {
			panic(err)
		}
		fname := fmt.Sprintf("%v-%v.pem", raddr, hex.EncodeToString(nonce[:]))
		err = ioutil.WriteFile(fname, chain, 0644)
		die.If(err)
		fmt.Printf("%v: [+] wrote %v.\n", raddr, fname)
	}
}
