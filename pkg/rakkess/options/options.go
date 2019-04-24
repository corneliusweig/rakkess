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

package options

import (
	"os"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

// RakkessOptions holds all user configuration options.
type RakkessOptions struct {
	ConfigFlags  *genericclioptions.ConfigFlags
	Verbs        []string
	OutputFormat string
	Streams      *genericclioptions.IOStreams
}

// NewRakkessOptions creates RakkessOptions with defaults.
func NewRakkessOptions() *RakkessOptions {
	return &RakkessOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		Streams: &genericclioptions.IOStreams{
			In:     os.Stdin,
			Out:    os.Stdout,
			ErrOut: os.Stderr,
		},
	}
}

// GetAuthClient creates a client for SelfSubjectAccessReviews with high queries per second.
func (o *RakkessOptions) GetAuthClient() (v1.SelfSubjectAccessReviewInterface, error) {
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	restConfig.QPS = 50
	restConfig.Burst = 250

	authClient := v1.NewForConfigOrDie(restConfig)
	return authClient.SelfSubjectAccessReviews(), nil
}

// DiscoveryClient creates a kubernetes discovery client.
func (o *RakkessOptions) DiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return o.ConfigFlags.ToDiscoveryClient()
}

// ExpandVerbs expands wildcard verbs `*` and `all`.
func (o *RakkessOptions) ExpandVerbs() {
	for _, verb := range o.Verbs {
		if verb == "*" || verb == "all" {
			o.Verbs = constants.ValidVerbs
		}
	}
}
