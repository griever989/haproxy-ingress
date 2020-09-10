package resolver

import (
	"plugin"
)

type resolveCookieFunc func(string, int, string) string

type ResolverPlugin struct {
	canResolveCookie  bool
	resolveCookieFunc resolveCookieFunc
}

// SHA1 returns the SHA1 of a file.
func createResolver(path string) *ResolverPlugin {
	resolver := &ResolverPlugin{}
	if path == "" {
		return resolver
	}
	p, err := plugin.Open(path)
	if err != nil {
		panic(err)
	}
	resolveCookieFuncLookup, err := p.Lookup("ResolveEndpointCookieValue")
	if err == nil {
		resolver.canResolveCookie = true
		resolver.resolveCookieFunc = resolveCookieFuncLookup.(func(string, int, string) string)
	}

	return resolver
}

func (r *ResolverPlugin) CanResolveCookie() bool {
	return r.canResolveCookie
}

func (r *ResolverPlugin) ResolveEndpointCookieValue(ip string, port int, targetRef string) (string, error) {
	return r.resolveCookieFunc(ip, port, targetRef), nil
}
