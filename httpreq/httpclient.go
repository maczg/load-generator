package httpreq

import (
	"crypto/tls"
	"net/http"
)

func NewHttpClient() *http.Client {
	return &http.Client{
		Transport: NewTransport(),
	}
}

func NewTransport() *http.Transport {
	return &http.Transport{
		DisableKeepAlives: true,
		TLSClientConfig:   NewTlsConfig(),
	}
}

func NewTlsConfig() *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: true,
	}
}
