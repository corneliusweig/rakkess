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
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

const (
	AccessAllowed       = iota
	AccessDenied        = iota
	AccessNotApplicable = iota
	AccessRequestErr    = iota
)

type Result struct {
	Name   string
	Access map[string]int
	Err    []error
}

type sortableResult []Result

func (s sortableResult) Len() int      { return len(s) }
func (s sortableResult) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableResult) Less(i, j int) bool {
	ret := strings.Compare(s[i].Name, s[j].Name)
	if ret > 0 {
		return false
	} else if ret == 0 {
		return i < j
	}
	return true
}

func CheckResourceAccess(ctx context.Context, sar authv1.SelfSubjectAccessReviewInterface, grs []GroupResource, verbs []string, namespace *string) (results []Result, err error) {
	group := sync.WaitGroup{}
	semaphore := make(chan struct{}, 20)
	resultsChan := make(chan Result)

	var ns string
	if namespace == nil {
		ns = ""
	} else {
		ns = *namespace
	}
	for _, gr := range grs {
		group.Add(1)
		// copy captured variables
		namespace := ns
		gr := gr
		go func(ctx context.Context, allowed chan<- Result) {
			defer group.Done()

			// exit early, if context is done
			select {
			case <-ctx.Done():
				return
			case semaphore <- struct{}{}:
			}

			logrus.Debugf("Checking access for %s", gr.fullName())

			// This seems to be a bug in kubernetes. If namespace is set for non-namespaced
			// resources, the access is reported as "allowed", but in fact it is forbidden.
			if !gr.APIResource.Namespaced {
				namespace = ""
			}

			allowedVerbs := sets.NewString(gr.APIResource.Verbs...)

			access := make(map[string]int)
			var errs []error
			for _, v := range verbs {

				// stop if cancelled
				select {
				case <-ctx.Done():
					<-semaphore
					return
				default:
				}

				if !allowedVerbs.Has(v) {
					access[v] = AccessNotApplicable
					continue
				}

				review := &v1.SelfSubjectAccessReview{
					Spec: v1.SelfSubjectAccessReviewSpec{
						ResourceAttributes: &v1.ResourceAttributes{
							Verb:      v,
							Resource:  gr.APIResource.Name,
							Group:     gr.APIGroup,
							Namespace: namespace,
						},
					},
				}
				review, e := sar.Create(review)
				if e != nil {
					errs = append(errs, e)
					access[v] = AccessRequestErr
				} else {
					access[v] = resultFor(&review.Status)
				}

			}
			<-semaphore
			allowed <- Result{
				Name:   gr.fullName(),
				Access: access,
				Err:    errs,
			}
		}(ctx, resultsChan)
	}

	go func() {
		group.Wait()
		close(resultsChan)
	}()

	for gr := range resultsChan {
		results = append(results, gr)
	}

	sort.Stable(sortableResult(results))

	return
}

func resultFor(status *v1.SubjectAccessReviewStatus) int {
	if status.Allowed {
		return AccessAllowed
	}
	return AccessDenied
}
