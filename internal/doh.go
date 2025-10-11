package internal

import (
	"bytes"
	"io"
	"log"
	"net/http"

	"github.com/miekg/dns"
)

const dohURL = "https://1.1.1.1/dns-query"

func HandleDNS(w dns.ResponseWriter, req *dns.Msg) {
	// 将 DNS 请求序列化为 wire format
	reqBytes, err := req.Pack()
	if err != nil {
		log.Println("Pack error:", err)
		return
	}

	// 发送到 DoH 上游
	client := &http.Client{
		Transport: &http.Transport{
			DisableKeepAlives: true,
		},
	}
	dohReq, _ := http.NewRequest("POST", dohURL, bytes.NewReader(reqBytes))
	dohReq.Header.Set("Content-Type", "application/dns-message")
	dohReq.Header.Set("Cache-Control", "no-cache")

	resp, err := client.Do(dohReq)
	if err != nil {
		log.Println("DoH request error:", err)
		return
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)

	// 将 DoH 响应反序列化为 DNS Msg
	dnsResp := new(dns.Msg)
	err = dnsResp.Unpack(respBytes)
	if err != nil {
		log.Println("Unpack error:", err)
		return
	}

	// 返回给客户端
	w.WriteMsg(dnsResp)
}
