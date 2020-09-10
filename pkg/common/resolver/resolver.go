/*
Copyright 2020 The HAProxy Ingress Controller Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package resolver

import (
	"errors"
	"plugin"
)

type ResolverPlugin struct {
	canResolveCookie  bool
	resolveCookieFunc func(string, int, string) string
}

// Creates a ResolverPlugin which loads the following functions from
// the file at the path specified:
//     ResolveEndpointCookieValue(ip string, port int, targetRef string) string
// Any functions not defined will not be loaded; not all of them need to be defined.
// The file must be compiled with `go build -buildmode=plugin`. See the "plugin"
// package for more informaton.
// If the path is empty, just returns a default resolver that returns false for
// all "CanResolve" functions.
func CreateResolver(path string) (*ResolverPlugin, error) {
	resolver := &ResolverPlugin{}
	if path == "" {
		return resolver, nil
	}
	p, err := plugin.Open(path)
	if err != nil {
		return nil, err
	}

	// lookupErr here is not a failure, because the function isn't required to be defined
	resolveCookieFuncLookup, lookupErr := p.Lookup("ResolveEndpointCookieValue")
	if lookupErr == nil {
		resolver.canResolveCookie = true
		resolver.resolveCookieFunc = resolveCookieFuncLookup.(func(string, int, string) string)
	}

	return resolver, nil
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
