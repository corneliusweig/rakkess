/*
Copyright 2020 Cornelius Weig

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

package client

import (
	"testing"

	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	openapi_v2 "github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
)

type fakeCachedDiscoveryInterface struct {
	invalidateCalls int
	next            metav1.APIResourceList
	fresh           bool
}

var _ discovery.CachedDiscoveryInterface = &fakeCachedDiscoveryInterface{}

func (c *fakeCachedDiscoveryInterface) Fresh() bool {
	return c.fresh
}

func (c *fakeCachedDiscoveryInterface) Invalidate() {
	c.invalidateCalls++
	c.fresh = true
}

func (c *fakeCachedDiscoveryInterface) RESTClient() restclient.Interface {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerGroups() (*metav1.APIGroupList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerGroupsAndResources() ([]*metav1.APIGroup, []*metav1.APIResourceList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerResources() ([]*metav1.APIResourceList, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	if c.fresh {
		return []*metav1.APIResourceList{&c.next}, nil
	}
	return nil, nil
}

func (c *fakeCachedDiscoveryInterface) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	if c.fresh {
		return []*metav1.APIResourceList{&c.next}, nil
	}
	return nil, nil
}

func (c *fakeCachedDiscoveryInterface) ServerVersion() (*version.Info, error) {
	panic("not implemented")
}

func (c *fakeCachedDiscoveryInterface) OpenAPISchema() (*openapi_v2.Document, error) {
	panic("not implemented")
}

var (
	aFoo = metav1.APIResource{
		Name:       "foo",
		Kind:       "Foo",
		Namespaced: false,
		Verbs:      []string{"list"},
	}
	aNoVerbs = metav1.APIResource{
		Name:       "baz",
		Kind:       "Baz",
		Namespaced: false,
		Verbs:      []string{},
	}
	bBar = metav1.APIResource{
		Name:       "bar",
		Kind:       "Bar",
		Namespaced: true,
		Verbs:      []string{"list"},
	}
)

func TestFetchAvailableGroupResources(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		verbs     []string
		resources metav1.APIResourceList
		expected  interface{}
	}{
		{
			name:      "cluster resources",
			namespace: "",
			verbs:     []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "a/v1",
				APIResources: []metav1.APIResource{aFoo, aNoVerbs},
			},
			expected: []GroupResource{{APIGroup: "a", APIResource: aFoo}},
		},
		{
			name:      "namespaced resources",
			namespace: "any-namespace",
			verbs:     []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "b/v1",
				APIResources: []metav1.APIResource{bBar},
			},
			expected: []GroupResource{{APIGroup: "b", APIResource: bBar}},
		},
		{
			name:      "empty api-resources",
			namespace: "any-namespace",
			verbs:     []string{"list"},
			resources: metav1.APIResourceList{
				GroupVersion: "c/v1",
				APIResources: []metav1.APIResource{},
			},
			expected: []GroupResource(nil),
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeRbacClient := &fakeCachedDiscoveryInterface{next: test.resources}

			getDiscoveryClient = func(opts *options.RakkessOptions) (discovery.CachedDiscoveryInterface, error) {
				return fakeRbacClient, nil
			}
			defer func() { getDiscoveryClient = getDiscoveryClientImpl }()

			opts := &options.RakkessOptions{
				ConfigFlags: &genericclioptions.ConfigFlags{
					Namespace: &test.namespace,
				},
			}
			grs, err := FetchAvailableGroupResources(opts)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, grs)
		})
	}
}

func TestGroupResource_fullName(t *testing.T) {
	grNoGroup := &GroupResource{
		APIGroup: "",
		APIResource: metav1.APIResource{
			Name: "foo",
		},
	}
	assert.Equal(t, "foo", grNoGroup.fullName())

	grGroup := &GroupResource{
		APIGroup: "v1",
		APIResource: metav1.APIResource{
			Name: "foo",
		},
	}
	assert.Equal(t, "foo.v1", grGroup.fullName())
}
