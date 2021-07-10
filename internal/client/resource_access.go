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
	v1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	authv1 "k8s.io/client-go/kubernetes/typed/authorization/v1"
	"k8s.io/klog/v2"
)

// CheckResourceAccess determines the access rights for the given GroupResources and verbs.
// Since it needs to do a lot of requests, the SelfSubjectAccessReviewInterface needs to
// be configured for high queries per second.
func CheckResourceAccess(ctx context.Context, sar authv1.SelfSubjectAccessReviewInterface, grs []GroupResource, verbs []string, namespace *string) result.MatrixPrinter {
	mu := sync.Mutex{} // guards res
	var res []result.ResourceAccessItem

	var ns string
	if namespace == nil {
		ns = ""
	} else {
		ns = *namespace
	}

	wg := sync.WaitGroup{}
	for _, gr := range grs {
		wg.Add(1)
		// copy captured variables
		namespace := ns
		gr := gr
		go func() {
			defer wg.Done()

			klog.V(2).Infof("Checking access for %s", gr.fullName())

			// This seems to be a bug in kubernetes. If namespace is set for non-namespaced
			// resources, the access is reported as "allowed", but in fact it is forbidden.
			if !gr.APIResource.Namespaced {
				namespace = ""
			}

			allowedVerbs := sets.NewString(gr.APIResource.Verbs...)

			access := make(map[string]result.Access)
			for _, v := range verbs {

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

				a := result.AccessDenied
				r, err := sar.Create(ctx, review, metav1.CreateOptions{})
				switch {
				case err != nil:
					a = result.AccessRequestErr
				case r.Status.Allowed:
					a = result.AccessAllowed
				}
				access[v] = a
			}

			mu.Lock()
			res = append(res, result.ResourceAccessItem{
				Name:   gr.fullName(),
				Access: access,
			})
			mu.Unlock()
		}()
	}

	wg.Wait()

	return result.NewResourceAccess(res)
}
