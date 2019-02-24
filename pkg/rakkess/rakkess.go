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

	"github.com/corneliusweig/rakkess/pkg/rakkess/client"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/corneliusweig/rakkess/pkg/rakkess/util"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

func Rakkess(ctx context.Context, opts *options.RakkessOptions) error {
	grs, err := client.FetchAvailableGroupResources(opts.ConfigFlags)
	if err != nil {
		return errors.Wrap(err, "fetch available group resources")
	}
	logrus.Info(grs)

	restConfig, err := opts.ConfigFlags.ToRESTConfig()
	if err != nil {
		return err
	}

	restConfig.QPS = 50
	restConfig.Burst = 250

	authClient := v1.NewForConfigOrDie(restConfig)
	results, err := client.CheckResourceAccess(ctx, authClient, grs, opts.Verbs)
	if err != nil {
		return err
	}

	util.PrintResults(opts.Streams.Out, opts.Verbs, results)

	return nil
}
