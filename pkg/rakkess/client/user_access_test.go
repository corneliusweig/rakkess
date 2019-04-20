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
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/authorization/v1"
	apiV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/typed/authorization/v1/fake"
	authTesting "k8s.io/client-go/testing"
)

type SelfSubjectAccessReviewDecision struct {
	v1.ResourceAttributes
	decision int
}

func (d *SelfSubjectAccessReviewDecision) matches(other *v1.SelfSubjectAccessReview) bool {
	return d.ResourceAttributes == *other.Spec.ResourceAttributes
}

type accessResult map[string]int

func buildAccess() accessResult {
	return make(map[string]int)
}
func (a accessResult) withResult(result int, verbs ...string) accessResult {
	for _, v := range verbs {
		a[v] = result
	}
	return a
}
func (a accessResult) allowed(verbs ...string) accessResult {
	return a.withResult(AccessAllowed, verbs...)
}
func (a accessResult) denied(verbs ...string) accessResult {
	return a.withResult(AccessDenied, verbs...)
}
func (a accessResult) get() map[string]int {
	return a
}

func toGroupResource(group, name string, verbs ...string) GroupResource {
	return GroupResource{
		APIGroup: group,
		APIResource: apiV1.APIResource{
			Name:  name,
			Verbs: verbs,
		},
	}
}

func TestCheckResourceAccess(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		verbs     []string
		input     []GroupResource
		decisions []*SelfSubjectAccessReviewDecision
		expected  []ResourceAccess
	}{
		{
			name:  "single resource, single verb",
			verbs: []string{"list"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list")},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{
						Resource: "resource1",
						Group:    "group1",
						Verb:     "list",
					},
					AccessAllowed,
				},
			},
			expected: []ResourceAccess{
				{Name: "resource1.group1", Access: buildAccess().allowed("list").get()},
			},
		},
		{
			name:  "single resource, invalid verb",
			verbs: []string{"patch"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list")},
			expected: []ResourceAccess{
				{Name: "resource1.group1", Access: buildAccess().withResult(AccessNotApplicable, "patch").get()},
			},
		},
		{
			name:  "single resource, multiple verbs",
			verbs: []string{"list", "create", "delete"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list", "create", "delete")},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "create"},
					AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "delete"},
					AccessDenied,
				},
			},
			expected: []ResourceAccess{
				{
					Name:   "resource1.group1",
					Access: buildAccess().allowed("list", "create").denied("delete").get(),
				},
			},
		},
		{
			name:  "multiple resources, single verb",
			verbs: []string{"list"},
			input: []GroupResource{
				toGroupResource("group1", "resource1", "list"),
				toGroupResource("group1", "resource2", "list"),
			},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource2", Group: "group1", Verb: "list"},
					AccessDenied,
				},
			},
			expected: []ResourceAccess{
				{
					Name:   "resource1.group1",
					Access: buildAccess().allowed("list").get(),
				},
				{
					Name:   "resource2.group1",
					Access: buildAccess().denied("list").get(),
				},
			},
		},
		{
			name:  "multiple resources, multiple verbs",
			verbs: []string{"list", "create"},
			input: []GroupResource{
				toGroupResource("group1", "resource1", "list", "create"),
				toGroupResource("group1", "resource2", "create"),
				toGroupResource("group2", "resource1", "list"),
			},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "create"},
					AccessDenied,
				},
				{
					v1.ResourceAttributes{Resource: "resource2", Group: "group1", Verb: "create"},
					AccessDenied,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group2", Verb: "list"},
					AccessAllowed,
				},
			},
			expected: []ResourceAccess{
				{
					Name:   "resource1.group1",
					Access: buildAccess().allowed("list").denied("create").get(),
				},
				{
					Name:   "resource1.group2",
					Access: buildAccess().withResult(AccessNotApplicable, "create").allowed("list").get(),
				},
				{
					Name:   "resource2.group1",
					Access: buildAccess().denied("create").withResult(AccessNotApplicable, "list").get(),
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			fakeReviews := &fake.FakeSelfSubjectAccessReviews{Fake: &fake.FakeAuthorizationV1{Fake: &authTesting.Fake{}}}
			fakeReviews.Fake.AddReactor("create", "selfsubjectaccessreviews",
				func(action authTesting.Action) (handled bool, ret runtime.Object, err error) {
					sar := action.(authTesting.CreateAction).GetObject().(*v1.SelfSubjectAccessReview)

					for _, d := range test.decisions {
						if d.matches(sar) {
							sar.Status.Allowed = d.decision == AccessAllowed
							return true, sar, nil
						}
					}
					return false, nil, nil
				})

			results, err := CheckResourceAccess(ctx, fakeReviews, test.input, test.verbs, nil)

			assert.NoError(t, err)
			assert.Equal(t, test.expected, results)
		})
	}
}

func TestSortResult(t *testing.T) {
	makeResult := func(key string, value int) map[string]int {
		result := make(map[string]int)
		result[key] = value
		return result
	}
	tests := []struct {
		name   string
		input  []ResourceAccess
		sorted []ResourceAccess
	}{
		{
			name:   "two inputs",
			input:  []ResourceAccess{{Name: "b second"}, {Name: "a first"}},
			sorted: []ResourceAccess{{Name: "a first"}, {Name: "b second"}},
		},
		{
			name:   "three inputs",
			input:  []ResourceAccess{{Name: "b second"}, {Name: "c third"}, {Name: "a first"}},
			sorted: []ResourceAccess{{Name: "a first"}, {Name: "b second"}, {Name: "c third"}},
		},
		{
			name: "three inputs, stable",
			input: []ResourceAccess{
				{Name: "same", Access: makeResult("b", 1)},
				{Name: "same", Access: makeResult("a", 2)},
				{Name: "same", Access: makeResult("c", 3)},
			},
			sorted: []ResourceAccess{
				{Name: "same", Access: makeResult("b", 1)},
				{Name: "same", Access: makeResult("a", 2)},
				{Name: "same", Access: makeResult("c", 3)},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sort.Stable(sortableResourceAccess(test.input))
			assert.Equal(t, test.sorted, test.input)
		})
	}
}
