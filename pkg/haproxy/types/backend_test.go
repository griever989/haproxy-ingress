/*
Copyright 2019 The HAProxy Ingress Controller Authors.

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

package types

import (
	"reflect"
	"strings"
	"testing"
)

func TestAddPath(t *testing.T) {
	testCases := []struct {
		input    []string
		expected []*BackendPath
	}{
		// 0
		{
			input: []string{"/"},
			expected: []*BackendPath{
				{ID: "path01", Link: PathLink{"d1.local", "/"}},
			},
		},
		// 1
		{
			input: []string{"/app", "/app"},
			expected: []*BackendPath{
				{ID: "path01", Link: PathLink{"d1.local", "/app"}},
			},
		},
		// 2
		{
			input: []string{"/app", "/root"},
			expected: []*BackendPath{
				{ID: "path02", Link: PathLink{"d1.local", "/root"}},
				{ID: "path01", Link: PathLink{"d1.local", "/app"}},
			},
		},
		// 3
		{
			input: []string{"/app", "/root", "/root"},
			expected: []*BackendPath{
				{ID: "path02", Link: PathLink{"d1.local", "/root"}},
				{ID: "path01", Link: PathLink{"d1.local", "/app"}},
			},
		},
		// 4
		{
			input: []string{"/app", "/root", "/app"},
			expected: []*BackendPath{
				{ID: "path02", Link: PathLink{"d1.local", "/root"}},
				{ID: "path01", Link: PathLink{"d1.local", "/app"}},
			},
		},
		// 5
		{
			input: []string{"/", "/app", "/root"},
			expected: []*BackendPath{
				{ID: "path03", Link: PathLink{"d1.local", "/root"}},
				{ID: "path02", Link: PathLink{"d1.local", "/app"}},
				{ID: "path01", Link: PathLink{"d1.local", "/"}},
			},
		},
	}
	for i, test := range testCases {
		b := &Backend{}
		for _, p := range test.input {
			b.AddBackendPath(CreatePathLink("d1.local", p))
		}
		if !reflect.DeepEqual(b.Paths, test.expected) {
			t.Errorf("backend.Paths differs on %d - actual: %v - expected: %v", i, b.Paths, test.expected)
		}
	}
}

func TestCreatePathConfig(t *testing.T) {
	type pathConfig struct {
		paths  string
		config interface{}
	}
	testCases := []struct {
		paths    []*BackendPath
		filter   string
		expected map[string][]pathConfig
	}{
		// 0
		{
			paths:  []*BackendPath{{ID: "path1"}},
			filter: "SSLRedirect",
			expected: map[string][]pathConfig{
				"SSLRedirect": {
					{paths: "path1", config: false},
				},
			},
		},
		// 1
		{
			paths: []*BackendPath{
				{ID: "path1", HSTS: HSTS{Enabled: true, MaxAge: 10}},
				{ID: "path2", HSTS: HSTS{Enabled: true, MaxAge: 10}},
				{ID: "path3", HSTS: HSTS{Enabled: true, MaxAge: 20}},
			},
			filter: "HSTS",
			expected: map[string][]pathConfig{
				"HSTS": {
					{
						paths:  "path1,path2",
						config: HSTS{Enabled: true, MaxAge: 10},
					},
					{
						paths:  "path3",
						config: HSTS{Enabled: true, MaxAge: 20},
					},
				},
			},
		},
		// 2
		{
			paths: []*BackendPath{
				{ID: "path1", HSTS: HSTS{Enabled: true, MaxAge: 10}, WhitelistHTTP: []string{"10.0.0.0/8"}},
				{ID: "path2", HSTS: HSTS{Enabled: true, MaxAge: 20}, WhitelistHTTP: []string{"10.0.0.0/8"}},
				{ID: "path3", HSTS: HSTS{Enabled: true, MaxAge: 20}},
			},
			filter: "HSTS,WhitelistHTTP",
			expected: map[string][]pathConfig{
				"HSTS": {
					{
						paths:  "path1",
						config: HSTS{Enabled: true, MaxAge: 10},
					},
					{
						paths:  "path2,path3",
						config: HSTS{Enabled: true, MaxAge: 20},
					},
				},
				"WhitelistHTTP": {
					{
						paths:  "path1,path2",
						config: []string{"10.0.0.0/8"},
					},
					{
						paths: "path3",
					},
				},
			},
		},
	}
	for i, test := range testCases {
		c := setup(t)
		backend := Backend{Paths: test.paths}
		backendPathConfig := backend.createPathConfig()
		actualConfig := map[string][]pathConfig{}
		for name, config := range backendPathConfig {
			if strings.Index(","+test.filter+",", ","+name+",") < 0 {
				continue
			}
			pathConfigs := actualConfig[name]
			for _, item := range config.items {
				paths := []string{}
				for _, p := range item.paths {
					paths = append(paths, p.ID)
				}
				itemConfig := item.config
				configValue := reflect.ValueOf(itemConfig)
				if configValue.Kind() == reflect.Slice && configValue.Len() == 0 {
					// empty slices and nil are semantically identical but DeepEquals disagrees
					itemConfig = nil
				}
				pathConfigs = append(pathConfigs, pathConfig{paths: strings.Join(paths, ","), config: itemConfig})
			}
			actualConfig[name] = pathConfigs
		}
		c.compareObjects("pathconfig", i, actualConfig, test.expected)
		c.teardown()
	}
}
