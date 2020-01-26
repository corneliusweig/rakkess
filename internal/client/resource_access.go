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
	"context"
	"sync"

	"github.com/corneliusweig/rakkess/internal/client/result"
	"github.com/sirupsen/logrus"
	v1 "k8s.io/api/authorization/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
)

// CheckResourceAccess determines the access rights for the given GroupResources and verbs.
// Since it needs to do a lot of requests, the SelfSubjectAccessReviewInterface needs to
// be configured for high queries per second.
func CheckResourceAccess(ctx context.Context, sar authv1.SelfSubjectAccessReviewInterface, grs []GroupResource, verbs []string, namespace *string) result.MatrixPrinter {
	results := make([]result.ResourceAccessItem, 0, len(grs))
	group := sync.WaitGroup{}
	semaphore := make(chan struct{}, 20)
	resultsChan := make(chan result.ResourceAccessItem)

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
		go func(ctx context.Context, allowed chan<- result.ResourceAccessItem) {
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
			for _, v := range verbs {

				// stop if cancelled
				select {
				case <-ctx.Done():
					<-semaphore
					return
				default:
				}

				if !allowedVerbs.Has(v) {
					access[v] = result.AccessNotApplicable
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

				if review, err := sar.Create(review); err != nil {
					access[v] = result.AccessRequestErr
				} else {
					access[v] = resultFor(&review.Status)
				}
			}
			<-semaphore
			allowed <- result.ResourceAccessItem{
				Name:   gr.fullName(),
				Access: access,
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

	return result.NewResourceAccess(results)
}

func resultFor(status *v1.SubjectAccessReviewStatus) int {
	if status.Allowed {
		return result.AccessAllowed
	}
	return result.AccessDenied
}
