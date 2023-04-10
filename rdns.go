package main

import (
	"log"

	"github.com/kost/dnstun"
)

func GenerateKey() string {
	return dnstun.GenerateKey()
}

func ServeDNS(dnslisten string, DnsDomain string, clients string, enckey string, dnsdelay string) error {
	dt := dnstun.NewDnsTunnel(DnsDomain, enckey)
	if dnsdelay != "" {
		err := dt.SetDnsDelay(dnsdelay)
		if err != nil {
			log.Printf("Error parsing DNS delay/sleep duration %s: %v", dnsdelay, err)
			return err
		}
	}
	dt.DnsServer(dnslisten, clients)
	err := dt.DnsServerStart()
	if err != nil {
		log.Printf("Error starting DNS server %s: %v", DnsDomain, err)
		return err
	}
	return nil
}
