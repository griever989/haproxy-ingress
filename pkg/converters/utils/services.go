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

package utils

import (
	"fmt"
	"net"
	"strconv"

	api "k8s.io/api/core/v1"

	"github.com/jcmoraisjr/haproxy-ingress/pkg/converters/types"
	hatypes "github.com/jcmoraisjr/haproxy-ingress/pkg/types"
)

// FindServicePort ...
func FindServicePort(svc *api.Service, servicePort string) *api.ServicePort {
	for _, port := range svc.Spec.Ports {
		if port.Name == servicePort || port.TargetPort.String() == servicePort {
			return &port
		}
	}
	svcPortNumber, err := strconv.ParseInt(servicePort, 10, 0)
	if err != nil {
		return nil
	}
	svcPort := int32(svcPortNumber)
	for _, port := range svc.Spec.Ports {
		if port.Port == svcPort {
			return &port
		}
	}
	return nil
}

// FindContainerPort Find the container's port number of a known servicePort
// Search criteria:
// 1. svcPort.TargetPort is a number: this is the right container's port
// 2. svcPort.TargetPort is a named port (not a number): find a container's port with that name and use its ContainerPort
// If targetPort is neither a valid port number nor a declared named port, return zero which means that the port was not found
func FindContainerPort(pod *api.Pod, svcPort *api.ServicePort) int {
	if targetPort := svcPort.TargetPort.IntValue(); targetPort > 0 {
		return targetPort
	}
	portName := svcPort.TargetPort.String()
	for _, c := range pod.Spec.Containers {
		for _, port := range c.Ports {
			if port.Protocol == svcPort.Protocol && port.Name == portName {
				return int(port.ContainerPort)
			}
		}
	}
	return 0
}

// FindEnvFromPod finds the value of an environment variable with name
// `name` from the containers of a pod. If the pod has multiple containers, it
// returns the first container that has a non-empty value for the env var.
func FindEnvFromPod(cache types.Cache, podTargetRef string, name string, logger hatypes.Logger) string {
	logger.Info("finding pod %s", podTargetRef)
	pod, err := cache.GetPod(podTargetRef)
	if err != nil {
		logger.Info("finding pod %s ... failed", podTargetRef)
		return ""
	}
	logger.Info("finding pod %s ... success", podTargetRef)
	logger.Info("iterating containers for pod %s ...", podTargetRef)
	containers := pod.Spec.Containers
	for i, container := range containers {
		for _, env := range container.Env {
			logger.Info("iterating env vars for container %v for pod %s ... name = %s, value = %s", i, podTargetRef, env.Name, env.Value)
			if env.Name == name {
				if env.Value != "" {
					logger.Info("iterating env vars for container %v for pod %s ... name = %s, value = %s ... MATCHED", i, podTargetRef, env.Name, env.Value)
					return env.Value
				} else if env.ValueFrom != nil {
					logger.Warn("Found environment variable %s for pod %s, however it is a fieldRef/keyRef and the value couldn't be read. Use a direct value instead.", env.Name, podTargetRef)
				}

			}
		}
	}
	return ""
}

// Endpoint ...
type Endpoint struct {
	IP        string
	Port      int
	TargetRef string
}

// CreateEndpoints ...
func CreateEndpoints(cache types.Cache, svc *api.Service, svcPort *api.ServicePort) (ready, notReady []*Endpoint, err error) {
	if svc.Spec.Type == api.ServiceTypeExternalName {
		ready, err := createEndpointsExternalName(svc, svcPort)
		return ready, nil, err
	}
	endpoints, err := cache.GetEndpoints(svc)
	if err != nil {
		return nil, nil, err
	}
	for _, subset := range endpoints.Subsets {
		for _, epPort := range subset.Ports {
			if matchPort(svcPort, &epPort) {
				port := int(epPort.Port)
				for _, addr := range subset.Addresses {
					ready = append(ready, newEndpointAddr(&addr, port))
				}
				for _, addr := range subset.NotReadyAddresses {
					notReady = append(notReady, newEndpointAddr(&addr, port))
				}
			}
		}
	}
	return ready, notReady, nil
}

func matchPort(svcPort *api.ServicePort, epPort *api.EndpointPort) bool {
	if epPort.Protocol != api.ProtocolTCP {
		return false
	}
	return svcPort.Name == "" || svcPort.Name == epPort.Name
}

// CreateSvcEndpoint ...
func CreateSvcEndpoint(svc *api.Service, svcPort *api.ServicePort) (endpoint *Endpoint, err error) {
	port := svcPort.Port
	if port <= 0 {
		return nil, fmt.Errorf("invalid port number: %d", port)
	}
	return newEndpointIP(svc.Spec.ClusterIP, int(port)), nil
}

var lookup = net.LookupIP

func createEndpointsExternalName(svc *api.Service, svcPort *api.ServicePort) (endpoints []*Endpoint, err error) {
	// TODO add support to undeclared ServicePort
	port := int(svcPort.Port)
	if port <= 0 {
		return nil, fmt.Errorf("invalid port number: %d", port)
	}
	addr, err := lookup(svc.Spec.ExternalName)
	if err != nil {
		return nil, err
	}
	endpoints = make([]*Endpoint, len(addr))
	for i, ip := range addr {
		endpoints[i] = newEndpointIP(ip.String(), port)
	}
	return endpoints, nil
}

func newEndpointAddr(addr *api.EndpointAddress, port int) *Endpoint {
	return &Endpoint{
		IP:        addr.IP,
		Port:      port,
		TargetRef: targetRefToString(addr.TargetRef),
	}
}

func targetRefToString(targetRef *api.ObjectReference) string {
	if targetRef == nil {
		return ""
	}
	return fmt.Sprintf("%s/%s", targetRef.Namespace, targetRef.Name)
}

func newEndpointIP(ip string, port int) *Endpoint {
	return &Endpoint{
		IP:   ip,
		Port: port,
	}
}

func (e *Endpoint) String() string {
	return fmt.Sprintf("%+v", *e)
}
