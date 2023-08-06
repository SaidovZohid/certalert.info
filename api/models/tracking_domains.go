package models

type DomainsNewReq struct {
	Domains []string `json:"domains"`
}

type DomainsReq struct {
	Domains string `json:"domains"`
}
