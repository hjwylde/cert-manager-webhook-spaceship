package spaceship

import (
	"context"
	"fmt"
	"net/http"
)

type DNSRecords struct {
	Force bool                    `json:"force,omitempty"`
	Items []DNSRecordsListTxtItem `json:"items"`
}

type DNSRecordsListTxtItem struct {
	Type  string `json:"type"`
	Name  string `json:"name"`
	Value string `json:"value"`
	TTL   int32  `json:"ttl,omitempty"`
}

func NewDNSRecordsListTxtItem(name, value string) DNSRecordsListTxtItem {
	return DNSRecordsListTxtItem{Type: "TXT", Name: name, Value: value}
}

type DNSRecordsService service

// Required permissions:
// * dnsrecords:write
func (s *DNSRecordsService) Delete(ctx context.Context, domain string, body []DNSRecordsListTxtItem) (*http.Response, error) {
	req, err := s.client.NewRequest("DELETE", fmt.Sprintf("dns/records/%s", domain), body)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}

// Required permissions:
// * dnsrecords:write
func (s *DNSRecordsService) Put(ctx context.Context, domain string, body DNSRecords) (*http.Response, error) {
	req, err := s.client.NewRequest("PUT", fmt.Sprintf("dns/records/%s", domain), body)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, req, nil)
}
