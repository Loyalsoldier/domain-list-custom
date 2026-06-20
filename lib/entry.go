package lib

import (
	"strings"

	router "github.com/v2fly/v2ray-core/v5/app/router/routercommon"
)

// Entry is a single domain list entry
type Entry struct {
	Name   string
	Domains []*router.Domain
}

// NewEntry creates a new Entry
func NewEntry(name string) *Entry {
	name = strings.ToUpper(strings.TrimSpace(name))
	return &Entry{
		Name:    name,
		Domains: make([]*router.Domain, 0),
	}
}

// GetName returns the name of the entry
func (e *Entry) GetName() string {
	return e.Name
}

// AddDomain adds a domain to the entry
func (e *Entry) AddDomain(domain *router.Domain) {
	if domain != nil {
		e.Domains = append(e.Domains, domain)
	}
}

// AddDomains adds multiple domains to the entry
func (e *Entry) AddDomains(domains []*router.Domain) {
	for _, domain := range domains {
		e.AddDomain(domain)
	}
}

// GetDomains returns all domains in the entry
func (e *Entry) GetDomains() []*router.Domain {
	return e.Domains
}

// MarshalText converts the entry to text format
func (e *Entry) MarshalText() ([]byte, error) {
	result := make([]byte, 0, 1024*512)
	
	for _, domain := range e.Domains {
		ruleVal := strings.TrimSpace(domain.GetValue())
		if len(ruleVal) == 0 {
			continue
		}

		var ruleString string
		switch domain.Type {
		case router.Domain_Full:
			ruleString = "full:" + ruleVal
		case router.Domain_RootDomain:
			ruleString = "domain:" + ruleVal
		case router.Domain_Plain:
			ruleString = "keyword:" + ruleVal
		case router.Domain_Regex:
			ruleString = "regexp:" + ruleVal
		}

		if len(domain.Attribute) > 0 {
			ruleString += ":"
			for _, attr := range domain.Attribute {
				ruleString += "@" + attr.GetKey() + ","
			}
			ruleString = strings.TrimRight(ruleString, ",")
		}
		result = append(result, []byte(ruleString+"\n")...)
	}

	return result, nil
}
