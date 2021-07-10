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
	"sort"
	"strings"
	"testing"

	"github.com/corneliusweig/rakkess/internal/client/result"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/authorization/v1"
	apiV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/typed/authorization/v1/fake"
	authTesting "k8s.io/client-go/testing"
)

type SelfSubjectAccessReviewDecision struct {
	v1.ResourceAttributes
	decision result.Access
}

func (d *SelfSubjectAccessReviewDecision) matches(other *v1.SelfSubjectAccessReview) bool {
	return d.ResourceAttributes == *other.Spec.ResourceAttributes
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
		want      []string
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
					result.AccessAllowed,
				},
			},
			want: []string{"resource1.group1:list->ok"},
		},
		{
			name:  "single resource, invalid verb",
			verbs: []string{"patch"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list")},
			want:  []string{"resource1.group1:patch->n/a"},
		},
		{
			name:  "single resource, multiple verbs",
			verbs: []string{"list", "create", "delete"},
			input: []GroupResource{toGroupResource("group1", "resource1", "list", "create", "delete")},
			decisions: []*SelfSubjectAccessReviewDecision{
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "list"},
					result.AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "create"},
					result.AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "delete"},
					result.AccessDenied,
				},
			},
			want: []string{"resource1.group1:create->ok,delete->no,list->ok"},
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
					result.AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource2", Group: "group1", Verb: "list"},
					result.AccessDenied,
				},
			},
			want: []string{"resource1.group1:list->ok", "resource2.group1:list->no"},
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
					result.AccessAllowed,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group1", Verb: "create"},
					result.AccessDenied,
				},
				{
					v1.ResourceAttributes{Resource: "resource2", Group: "group1", Verb: "create"},
					result.AccessDenied,
				},
				{
					v1.ResourceAttributes{Resource: "resource1", Group: "group2", Verb: "list"},
					result.AccessAllowed,
				},
			},
			want: []string{"resource1.group1:create->no,list->ok", "resource1.group2:create->n/a,list->ok", "resource2.group1:create->no,list->n/a"},
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
							sar.Status.Allowed = d.decision == result.AccessAllowed
							return true, sar, nil
						}
					}
					return false, nil, nil
				})

			results := CheckResourceAccess(ctx, fakeReviews, test.input, test.verbs, nil)

			var got []string
			for name, access := range results {
				var as []string
				for verb, a := range access {
					var outcome string
					switch a {
					case result.AccessAllowed:
						outcome = "ok"
					case result.AccessDenied:
						outcome = "no"
					case result.AccessNotApplicable:
						outcome = "n/a"
					}
					as = append(as, verb+"->"+outcome)
				}
				sort.Strings(as)
				got = append(got, name+":"+strings.Join(as, ","))
			}
			sort.Strings(got)
			assert.Equal(t, test.want, got)
		})
	}
}
