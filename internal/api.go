package internal

import (
	"fmt"
	"net/http"
	"strings"
)

func AuthMiddleware(gatewayToken string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("verify-token")
		if token != gatewayToken {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}

func HandleHttp(w http.ResponseWriter, r *http.Request) {

	// 从 URL 获取域名参数
	domainParam := r.URL.Query().Get("domain")
	if domainParam == "" {
		http.Error(w, "domain parameter is required", http.StatusBadRequest)
		return
	}
	domains := strings.Split(domainParam, ",")

	outputFormat := r.URL.Query().Get("format")

	// 生成证书
	certificates, err := GenerateCertificate(domains)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error generating certificate: %v", err), http.StatusInternalServerError)
		return
	}

	if outputFormat == "json" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(fmt.Sprintf(`{
		"certificate": "%s",
		"private_key": "%s"
	}`, certificates.Certificate, certificates.PrivateKey)))
	} else {
		// 设置下载响应头
		w.Header().Set("Content-Type", "application/x-pem-file")
		w.Header().Set("Content-Disposition", "attachment; filename=\"cert.pem\"")

		// 输出证书 + 私钥
		w.Write(certificates.Certificate) // 证书 + 中间证书
		w.Write([]byte("\n"))
		w.Write(certificates.PrivateKey) // 私钥
	}
}
