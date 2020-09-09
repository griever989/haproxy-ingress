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

package helper_test

import (
	"strings"

	api "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
)

// CreateService ...
func CreateService(name, port, endpoints string) (*api.Service, *api.Endpoints, []api.Pod) {
	sname := strings.Split(name, "/") // namespace/name of the service
	sport := strings.Split(port, ":") // numeric-port -or- name:numeric-port -or- name:numeric-port:named-port
	if len(sport) < 2 {
		sport = []string{"", port, port}
	} else if len(sport) < 3 {
		sport = []string{sport[0], sport[1], sport[1]}
	}

	svc := CreateObject(`
apiVersion: v1
kind: Service
metadata:
  name: ` + sname[1] + `
  namespace: ` + sname[0] + `
spec:
  ports:
  - name: ` + sport[0] + `
    port: ` + sport[1] + `
    targetPort: ` + sport[2]).(*api.Service)

	ep := CreateObject(`
apiVersion: v1
kind: Endpoints
metadata:
  name: ` + sname[1] + `
  namespace: ` + sname[0] + `
subsets:
- addresses: []
  ports:
  - name: ` + sport[0] + `
    port: ` + sport[1] + `
    protocol: TCP`).(*api.Endpoints)

	addr := []api.EndpointAddress{}
	pods := []api.Pod{}
	for _, e := range strings.Split(endpoints, ",") {
		if e != "" {
			namespace := sname[0]
			podname := sname[1] + "-xxxxx"
			target := &api.ObjectReference{
				Kind:      "Pod",
				Name:      podname,
				Namespace: namespace,
			}
			addr = append(addr, api.EndpointAddress{IP: e, TargetRef: target})
			pods = append(pods, api.Pod{
				ObjectMeta: metav1.ObjectMeta{
					Name:      podname,
					Namespace: namespace,
					Labels:    map[string]string{},
				},
			})
		}
	}
	ep.Subsets[0].Addresses = addr

	return svc, ep, pods
}

// CreateObject ...
func CreateObject(cfg string) runtime.Object {
	decode := scheme.Codecs.UniversalDeserializer().Decode
	obj, _, err := decode([]byte(cfg), nil, nil)
	if err != nil {
		return nil
	}
	return obj
}
