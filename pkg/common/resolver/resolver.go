package resolver

import (
	"errors"
	"plugin"
)

type resolveCookieFunc func(string, int, string) string

type ResolverPlugin struct {
	canResolveCookie  bool
	resolveCookieFunc resolveCookieFunc
}

// Creates a ResolverPlugin which loads the following functions from
// the file at the path specified:
//     ResolveEndpointCookieValue(ip string, port int, targetRef string) string
// Any functions not defined will not be loaded; not all of them need to be defined.
// The file must be compiled with `go build -buildmode=plugin`. See the "plugin"
// package for more informaton.
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
		resolver.resolveCookieFunc = resolveCookieFuncLookup.(resolveCookieFunc)
	}

	return resolver
}

func (r *ResolverPlugin) CanResolveCookie() bool {
	return r.canResolveCookie
}

func (r *ResolverPlugin) ResolveEndpointCookieValue(ip string, port int, targetRef string) (string, error) {
	if !r.CanResolveCookie() {
		return "", errors.New("Cookie resolver function is not registered")
	}
	return r.resolveCookieFunc(ip, port, targetRef), nil
}
