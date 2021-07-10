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

package internal

import (
	"context"
	"fmt"

	"github.com/corneliusweig/rakkess/internal/client"
	"github.com/corneliusweig/rakkess/internal/options"
	"github.com/corneliusweig/rakkess/internal/validation"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/klog/v2"
)

// Resource determines the access right of the current (or impersonated) user
// and prints the result as a matrix with verbs in the horizontal and resource names
// in the vertical direction.
func Resource(ctx context.Context, opts *options.RakkessOptions) error {
	if err := validation.Options(opts); err != nil {
		return err
	}

	grs, err := client.FetchAvailableGroupResources(opts)
	if err != nil {
		return errors.Wrap(err, "fetch available group resources")
	}
	klog.V(2).Info(grs)

	authClient, err := opts.GetAuthClient()
	if err != nil {
		return errors.Wrap(err, "get auth client")
	}

	namespace := opts.ConfigFlags.Namespace
	results := client.CheckResourceAccess(ctx, authClient, grs, opts.Verbs, namespace)
	p := results.ToPrinter(opts.Verbs)
	p.Print(opts.Streams.Out, opts.OutputFormat)

	if namespace == nil || *namespace == "" {
		fmt.Fprintf(opts.Streams.Out, "No namespace given, this implies cluster scope (try -n if this is not intended)\n")
	}

	return nil
}

// Subject determines the subjects with access right to the given resource and
// prints the result as a matrix with verbs in the horizontal and subject names
// in the vertical direction.
func Subject(ctx context.Context, opts *options.RakkessOptions, resource, resourceName string) error {
	if err := validation.OutputFormat(opts.OutputFormat); err != nil {
		return err
	}

	mapper, err := opts.ConfigFlags.ToRESTMapper()
	if err != nil {
		return errors.Wrap(err, "cannot create k8s REST mapper")
	}
	versionedResource, err := mapper.ResourceFor(schema.GroupVersionResource{Resource: resource})
	if err != nil {
		return errors.Wrap(err, "determine requested resource")
	}

	subjectAccess, err := client.GetSubjectAccess(ctx, opts, versionedResource.Resource, resourceName)
	if err != nil {
		return errors.Wrap(err, "get subject access")
	}

	if subjectAccess.Empty() {
		klog.Warningf("No subjects with access found. This most likely means that you have insufficient rights to review authorization.")
		return nil
	}

	p := subjectAccess.ToPrinter(opts.Verbs)
	p.Print(opts.Streams.Out, opts.OutputFormat)

	namespace := opts.ConfigFlags.Namespace
	if namespace == nil || *namespace == "" {
		fmt.Fprintf(opts.Streams.Out, "Only ClusterRoleBindings are considered, because no namespace is given.\n")
	}

	return nil
}
