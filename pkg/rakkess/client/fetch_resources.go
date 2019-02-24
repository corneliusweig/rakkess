/*
Copyright 2019 Cornelius Weig

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
	"fmt"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GroupResource contains the APIGroup and APIResource
type GroupResource struct {
	APIGroup    string
	APIResource metav1.APIResource
}

// Extracts the full name including APIGroup, e.g. 'deployment.apps'
func (g GroupResource) fullName() string {
	if g.APIGroup == "" {
		return g.APIResource.Name
	}
	return fmt.Sprintf("%s.%s", g.APIResource.Name, g.APIGroup)
}

func FetchAvailableGroupResources(opts *options.RakkessOptions) ([]GroupResource, error) {
	client, err := opts.DiscoveryClient()
	if err != nil {
		return nil, errors.Wrap(err, "discovery client")
	}

	client.Invalidate()

	var resourcesFetcher func() ([]*metav1.APIResourceList, error)
	if opts.ConfigFlags.Namespace == nil || *opts.ConfigFlags.Namespace == "" {
		resourcesFetcher = client.ServerPreferredResources
	} else {
		resourcesFetcher = client.ServerPreferredNamespacedResources
	}

	resources, err := resourcesFetcher()
	if err != nil {
		return nil, errors.Wrap(err, "get preferred resources")
	}

	var grs []GroupResource
	for _, list := range resources {
		if len(list.APIResources) == 0 {
			continue
		}
		gv, err := schema.ParseGroupVersion(list.GroupVersion)
		if err != nil {
			logrus.Warnf("Cannot parse groupVersion: %s", err)
			continue
		}
		for _, r := range list.APIResources {
			if len(r.Verbs) == 0 {
				continue
			}

			grs = append(grs, GroupResource{
				APIGroup:    gv.Group,
				APIResource: r,
			})
		}
	}

	return grs, nil
}
