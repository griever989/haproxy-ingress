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
	"testing"
)

// We should always be able to create an empty resolver that just returns false
// for everything for ease of use
func TestCreateEmptyResolver(t *testing.T) {
	resolver, err := CreateResolver("")
	if err != nil {
		t.Fatalf("Failed to load empty resolver")
	}

	expectedCanResolveCookie := false

	resultCanResolveCookie := resolver.CanResolveCookie()
	if expectedCanResolveCookie != resultCanResolveCookie {
		t.Errorf("CanResolveCookie() result differs. Expected: %v, Actual: %v", expectedCanResolveCookie, resultCanResolveCookie)
	}
}

func TestLoadAndCallResolverPlugin(t *testing.T) {
	// it's expected that this plugin is precompiled before running this test
	filepath := "./pkg/common/resolver/testplugin/resolver_test_plugin.so"
	resolver, err := CreateResolver(filepath)
	if err != nil {
		t.Fatalf("Failed to load resolver at path '%s'", filepath)
	}

	testIp := "127.0.0.100"
	testPort := 8080
	testTargetRef := "default/abc-vwxyz"
	expectedCanResolveCookie := true
	expectedResolvedCookie := "127.0.0.100_8080_cook"

	resultCanResolveCookie := resolver.CanResolveCookie()
	if expectedCanResolveCookie != resultCanResolveCookie {
		t.Errorf("CanResolveCookie() result differs. Expected: %v, Actual: %v", expectedCanResolveCookie, resultCanResolveCookie)
	}

	resultResolvedCookie, err := resolver.ResolveEndpointCookieValue(testIp, testPort, testTargetRef)
	if err != nil {
		t.Errorf("Failed to call ResolveEndpointCookieValue: %v", err)
	} else if expectedResolvedCookie != resultResolvedCookie {
		t.Errorf("ResolveEndpointCookieValue() result differs. Expected: %v, Actual: %v", expectedResolvedCookie, resultResolvedCookie)
	}
}
