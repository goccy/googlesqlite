package net

import (
	"net/netip"
	"net/url"
	"regexp"
	"strings"

	"golang.org/x/net/idna"
	"golang.org/x/net/publicsuffix"
)

func parseURL(v string) *url.URL {
	parsed, err := url.Parse(v)
	if err != nil {
		return nil
	}
	if parsed.Host == "" {
		parsed, err = url.Parse("//" + strings.TrimSpace(v))
		if err != nil {
			return nil
		}
	}
	return parsed
}

func publicSuffix(host string) (string, error) {
	if publicSuffixMatcher.MatchString(host) {
		return "", nil
	}
	splitHost := strings.Split(host, ".")
	encoded, err := idna.ToASCII(strings.ToLower(host))
	if err != nil {
		return "", err
	}
	suffix, icann := publicsuffix.PublicSuffix(encoded)
	if !icann {
		return "", nil
	}
	splitSuffix := strings.Split(suffix, ".")
	return strings.Join(splitHost[len(splitHost)-len(splitSuffix):], "."), nil
}

func parseIP(v string) ([]byte, error) {
	ip, err := netip.ParseAddr(v)
	if err != nil {
		return nil, err
	}
	return ip.AsSlice(), nil
}

var publicSuffixMatcher = regexp.MustCompile(`[^.]\.{2,}[^.]`)
