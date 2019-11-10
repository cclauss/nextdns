package flag

import (
	"context"
	"flag"
	"fmt"
	"strings"

	"github.com/nextdns/nextdns/resolver"
)

// Resolver defines a forwarder server with some optional conditions.
type Resolver struct {
	resolver.Resolver
	Domain string
}

// newResolver parses a server definition with an optional condition.
func newResolver(v string) (Resolver, error) {
	idx := strings.IndexByte(v, '=')
	addr := v
	var r Resolver
	if idx != -1 {
		addr = strings.TrimSpace(v[idx+1:])
		r.Domain = fqdn(strings.TrimSpace(v[:idx]))
	}
	var err error
	r.Resolver, err = resolver.New(addr)
	return r, err
}

// Match resturns true if the rule matches domain.
func (r Resolver) Match(domain string) bool {
	if r.Domain != "" {
		if domain != r.Domain && !isSubDomain(domain, r.Domain) {
			return false
		}
	}
	return true
}

func fqdn(s string) string {
	if !strings.HasSuffix(s, ".") {
		s += "."
	}
	return s
}

func isSubDomain(sub, domain string) bool {
	return strings.HasSuffix(sub, "."+domain)
}

// Forwarders is a list of Resolver with rules.
type Forwarders []Resolver

// Get returns the server matching the domain conditions.
func (f *Forwarders) Get(domain string) resolver.Resolver {
	for _, s := range *f {
		if s.Match(domain) {
			return s.Resolver
		}
	}
	return nil
}

// String is the method to format the flag's value
func (f *Forwarders) String() string {
	return fmt.Sprint(*f)
}

// Set is the method to set the flag value, part of the flag.Value interface.
func (f *Forwarders) Set(value string) error {
	r, err := newResolver(value)
	if err != nil {
		return err
	}
	*f = append(*f, r)
	return nil
}

// Forwarder defines a string flag defining forwarder rules. The flag can be
// repeated.
func Forwarder(name, usage string) *Forwarders {
	f := &Forwarders{}
	flag.Var(f, name, usage)
	return f
}

// Resolve implements proxy.Resolver interface.
func (f *Forwarders) Resolve(ctx context.Context, q resolver.Query, buf []byte) (int, error) {
	r := f.Get(q.Name)
	if r == nil {
		return -1, fmt.Errorf("%s: no forwarder defined", q.Name)
	}
	return r.Resolve(ctx, q, buf)
}
