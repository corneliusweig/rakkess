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

package result

import (
	"testing"

	"github.com/corneliusweig/rakkess/internal/constants"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

func TestSubjectAccess_MatchRules(t *testing.T) {
	r := RoleRef{
		Name: "some-role",
		Kind: "some-kind",
	}
	resource := "deployments"
	tests := []struct {
		name          string
		resourceName  string
		initialVerbs  []string
		rule          v1.PolicyRule
		expectedVerbs []string
	}{
		{
			name: "simple rule",
			rule: v1.PolicyRule{
				Resources: []string{resource},
				Verbs:     []string{"create", "get"},
			},
			expectedVerbs: []string{"create", "get"},
		},
		{
			name:         "simple rule with initial verbs",
			initialVerbs: []string{"initial", "other"},
			rule: v1.PolicyRule{
				Resources: []string{resource},
				Verbs:     []string{"create", "get"},
			},
			expectedVerbs: []string{"create", "get", "initial", "other"},
		},
		{
			name: "rule for multiple resources",
			rule: v1.PolicyRule{
				Resources: []string{"resource-other", resource, "resource-yet-another"},
				Verbs:     []string{"create", "get"},
			},
			expectedVerbs: []string{"create", "get"},
		},
		{
			name: "no matching resource",
			rule: v1.PolicyRule{
				Resources: []string{"resource-other", "resource-yet-another"},
				Verbs:     []string{"create", "get"},
			},
		},
		{
			name: "VerbAll",
			rule: v1.PolicyRule{
				Resources: []string{resource},
				Verbs:     []string{v1.VerbAll},
			},
			expectedVerbs: constants.ValidVerbs,
		},
		{
			name: "simple rule with resourceNames does not match",
			rule: v1.PolicyRule{
				Resources:     []string{resource},
				ResourceNames: []string{"no-match"},
				Verbs:         []string{"create", "get"},
			},
		},
		{
			name:         "simple rule with matching resourceName",
			resourceName: "my-resource-name",
			rule: v1.PolicyRule{
				Resources:     []string{resource},
				ResourceNames: []string{"my-resource-name"},
				Verbs:         []string{"create", "get"},
			},
			expectedVerbs: []string{"create", "get"},
		},
		{
			name:         "simple rule with wrong resourceName",
			resourceName: "my-resource-name",
			rule: v1.PolicyRule{
				Resources:     []string{resource},
				ResourceNames: []string{"wrong-resource-name"},
				Verbs:         []string{"create", "get"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sa := NewSubjectAccess(resource, test.resourceName)
			if test.initialVerbs != nil {
				sa.roleToVerbs[r] = sets.NewString(test.initialVerbs...)
			}
			sa.MatchRules(r, test.rule)

			if test.expectedVerbs != nil {
				assert.Equal(t, sets.NewString(test.expectedVerbs...), sa.roleToVerbs[r])
			} else {
				_, ok := sa.roleToVerbs[r]
				assert.False(t, ok)
			}
		})
	}
}

func TestSubjectAccess_ResolveRoleRef(t *testing.T) {
	r := RoleRef{
		Name: "some-role",
		Kind: "some-kind",
	}
	subject := "main"
	mainSubject := SubjectRef{Name: subject, Kind: "some-kind", Namespace: "some-ns"}
	tests := []struct {
		name          string
		verbsForRole  []string
		subjects      []string
		expectedVerbs []string
	}{
		{
			name:          "no role",
			subjects:      []string{subject},
			expectedVerbs: []string{"initial-verb"},
		},
		{
			name:          "match with one subject",
			verbsForRole:  []string{"get", "list"},
			subjects:      []string{subject},
			expectedVerbs: []string{"initial-verb", "get", "list"},
		},
		{
			name:          "match with multiple subject",
			verbsForRole:  []string{"get", "list"},
			subjects:      []string{"other", subject, "yet-another"},
			expectedVerbs: []string{"initial-verb", "get", "list"},
		},
		{
			name:          "no match with other subjects",
			verbsForRole:  []string{"get", "list"},
			subjects:      []string{"other", "yet-another"},
			expectedVerbs: []string{"initial-verb"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sa := SubjectAccess{
				subjectToVerbs: map[SubjectRef]sets.String{mainSubject: sets.NewString("initial-verb")},
				roleToVerbs:    make(map[RoleRef]sets.String),
			}
			if test.verbsForRole != nil {
				sa.roleToVerbs[r] = sets.NewString(test.verbsForRole...)
			}

			subjects := make([]v1.Subject, 0, len(test.subjects))
			for _, s := range test.subjects {
				subjects = append(subjects, v1.Subject{
					Name:      s,
					Kind:      "some-kind",
					Namespace: "some-ns",
				})
			}
			sa.ResolveRoleRef(r, subjects)

			assert.Equal(t, sets.NewString(test.expectedVerbs...), sa.subjectToVerbs[mainSubject])
		})
	}
}
