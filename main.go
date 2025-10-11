package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/miekg/dns"
	"github.com/vhvy/acme-gateway-go/internal"
)

const HTTPPort = "48952"

func main() {
	gatewayToken := os.Getenv("GATEWAY_TOKEN")
	acmeEmail := os.Getenv("ACME_EMAIL")
	cfToken := os.Getenv("CLOUDFLARE_DNS_API_TOKEN")
	if gatewayToken == "" {
		log.Fatal("GATEWAY_TOKEN not set")
	}
	if acmeEmail == "" {
		log.Fatal("ACME_EMAIL not set")
	}
	if cfToken == "" {
		log.Fatal("CF API TOKEN not set")
	}

	http.HandleFunc("/renew", internal.AuthMiddleware(gatewayToken, internal.HandleHttp))

	srv := &http.Server{
		Addr:         ":" + HTTPPort,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 600 * time.Second, // DNS-01 验证可能比较慢
	}

	log.Println("Server listening on :" + HTTPPort)

	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Fatalf("Could not start HTTP server: %v", err)
		}
	}()

	dns.HandleFunc(".", internal.HandleDNS)

	server := &dns.Server{Addr: ":5353", Net: "udp"}
	log.Println("DNS server listening on :5353")
	log.Fatal(server.ListenAndServe())
}
