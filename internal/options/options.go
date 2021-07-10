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

package options

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/corneliusweig/rakkess/internal/constants"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"k8s.io/client-go/discovery"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/klog/v2"
)

// RakkessOptions holds all user configuration options.
type RakkessOptions struct {
	ConfigFlags      *genericclioptions.ConfigFlags
	Verbs            []string
	AsServiceAccount string
	OutputFormat     string
	Streams          *genericclioptions.IOStreams
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

// Sets up options with in-memory buffers as in- and output-streams
func NewTestRakkessOptions() (*RakkessOptions, *bytes.Buffer, *bytes.Buffer, *bytes.Buffer) {
	iostreams, in, out, errout := genericclioptions.NewTestIOStreams()
	klog.SetOutput(errout)
	return &RakkessOptions{
		ConfigFlags: genericclioptions.NewConfigFlags(true),
		Streams:     &iostreams,
	}, in, out, errout
}

// GetAuthClient creates a client for SelfSubjectAccessReviews with high queries per second.
func (o *RakkessOptions) GetAuthClient() (v1.SelfSubjectAccessReviewInterface, error) {
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	restConfig.QPS = 500
	restConfig.Burst = 1000

	authClient := v1.NewForConfigOrDie(restConfig)
	return authClient.SelfSubjectAccessReviews(), nil
}

// DiscoveryClient creates a kubernetes discovery client.
func (o *RakkessOptions) DiscoveryClient() (discovery.CachedDiscoveryInterface, error) {
	return o.ConfigFlags.ToDiscoveryClient()
}

func (o *RakkessOptions) ExpandServiceAccount() error {
	if o.AsServiceAccount == "" {
		return nil
	}

	if o.ConfigFlags.Impersonate != nil && *o.ConfigFlags.Impersonate != "" {
		return fmt.Errorf("--%s cannot be mixed with --as", constants.FlagServiceAccount)
	}

	qualifiedServiceAccount, err := o.namespacedServiceAccount()
	if err != nil {
		return err
	}

	impersonate := fmt.Sprintf("system:serviceaccount:%s", qualifiedServiceAccount)
	o.ConfigFlags.Impersonate = &impersonate
	return nil
}

func (o *RakkessOptions) namespacedServiceAccount() (string, error) {
	if strings.Contains(o.AsServiceAccount, ":") {
		return o.AsServiceAccount, nil
	}

	if o.ConfigFlags.Namespace != nil && *o.ConfigFlags.Namespace != "" {
		return fmt.Sprintf("%s:%s", *o.ConfigFlags.Namespace, o.AsServiceAccount), nil
	}

	return "", fmt.Errorf("serviceAccounts are namespaced, either provide --namespace or fully qualify the serviceAccount: '<namespace>:%s'", o.AsServiceAccount)
}

// ExpandVerbs expands wildcard verbs `*` and `all`.
func (o *RakkessOptions) ExpandVerbs() {
	for _, verb := range o.Verbs {
		if verb == "*" || verb == "all" {
			o.Verbs = constants.ValidVerbs
		}
	}
}
