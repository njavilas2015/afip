package internal

import (
	"crypto/tls"
	"errors"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type AFIPAdapter struct {
	client *http.Client
}

var (
	TZ_AR, _ = time.LoadLocation("America/Argentina/Buenos_Aires")

	WSDLS = map[string]map[string]string{
		"production": {
			"wsaa":                         "https://wsaa.afip.gov.ar/ws/services/LoginCms?wsdl",
			"wsfe":                         "https://servicios1.afip.gov.ar/wsfev1/service.asmx?WSDL",
			"ws_sr_constancia_inscripcion": "https://aws.afip.gov.ar/sr-padron/webservices/personaServiceA5?WSDL",
			"ws_sr_padron_a13":             "https://aws.afip.gov.ar/sr-padron/webservices/personaServiceA13?WSDL",
		},
		"sandbox": {
			"wsaa":                         "https://wsaahomo.afip.gov.ar/ws/services/LoginCms?wsdl",
			"wsfe":                         "https://wswhomo.afip.gov.ar/wsfev1/service.asmx?WSDL",
			"ws_sr_constancia_inscripcion": "https://awshomo.afip.gov.ar/sr-padron/webservices/personaServiceA5?WSDL",
			"ws_sr_padron_a13":             "https://awshomo.afip.gov.ar/sr-padron/webservices/personaServiceA13?WSDL",
		},
	}
	cacheOnce       sync.Once
	cachedTransport *Transport
)

type Transport struct {
	client *http.Client
	cache  map[string]string
}

func NewAFIPAdapter() *AFIPAdapter {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			MinVersion: tls.VersionTLS10,
			CipherSuites: []uint16{
				tls.TLS_RSA_WITH_AES_128_CBC_SHA,
			},
		},
	}

	client := &http.Client{Transport: tr}

	return &AFIPAdapter{client: client}
}

func GetOrCreateTransport() *Transport {

	cacheOnce.Do(func() {

		client := &http.Client{}

		for _, env := range WSDLS {
			for _, wsdl := range env {
				parsedURL, err := url.Parse(wsdl)
				if err != nil {
					log.Fatalf("Failed to parse URL: %v", err)
				}
				baseURL := parsedURL.Scheme + "://" + parsedURL.Host
				client.Transport = NewAFIPAdapter().client.Transport
				log.Printf("Registered adapter for %s", baseURL)
			}
		}

		cachedTransport = &Transport{
			client: client,
			cache:  make(map[string]string),
		}
	})

	return cachedTransport
}

func GetClient(serviceName string, sandbox bool) (*http.Client, error) {

	environment := "production"

	if sandbox {
		environment = "sandbox"
	}

	key := serviceName

	wsdl, exists := WSDLS[environment][key]

	if !exists {
		return nil, errors.New("unknown service name: " + serviceName)
	}

	transport := GetOrCreateTransport()

	transport.cache[serviceName] = wsdl

	return transport.client, nil
}

func main() {

	client, err := GetClient("wsaa", false)

	if err != nil {
		log.Fatalf("Error getting client: %v", err)
	}

	log.Println("Client configured:", client)
}
