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

package rakkess

import (
	"context"
	"fmt"

	"github.com/corneliusweig/rakkess/pkg/rakkess/client"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/corneliusweig/rakkess/pkg/rakkess/printer"
	"github.com/corneliusweig/rakkess/pkg/rakkess/validation"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func RakkessResource(ctx context.Context, opts *options.RakkessOptions) error {
	if err := validation.Options(opts); err != nil {
		return err
	}

	grs, err := client.FetchAvailableGroupResources(opts)
	if err != nil {
		return errors.Wrap(err, "fetch available group resources")
	}
	logrus.Debug(grs)

	authClient, err := opts.GetAuthClient()
	if err != nil {
		return errors.Wrap(err, "get auth client")
	}

	namespace := opts.ConfigFlags.Namespace
	results, err := client.CheckResourceAccess(ctx, authClient, grs, opts.Verbs, namespace)
	if err != nil {
		return errors.Wrap(err, "check resource access")
	}

	printer.PrintResults(opts.Streams.Out, opts.Verbs, opts.OutputFormat, results)

	if namespace == nil || *namespace == "" {
		fmt.Fprintf(opts.Streams.Out, "No namespace given, this implies cluster scope (try -n if this is not intended)\n")
	}

	return nil
}

func RakkessSubject(opts *options.RakkessOptions, resource string) error {
	if err := validation.Options(opts); err != nil {
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

	subjectAccess, err := client.GetSubjectAccess(opts, versionedResource.Resource)
	if err != nil {
		return errors.Wrap(err, "get subject access")
	}

	if len(subjectAccess.Get()) == 0 {
		logrus.Warnf("No subjects with access found. This most likely means that you have insufficient rights to review authorization.")
		return nil
	}

	printer.PrintResults(opts.Streams.Out, opts.Verbs, opts.OutputFormat, subjectAccess)

	namespace := opts.ConfigFlags.Namespace
	if namespace == nil || *namespace == "" {
		fmt.Fprintf(opts.Streams.Out, "Only ClusterRoleBindings are considered, because no namespace is given.\n")
	}

	return nil
}
