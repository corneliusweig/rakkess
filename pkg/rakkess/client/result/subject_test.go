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

package result

import (
	"testing"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
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
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sa := NewSubjectAccess(resource)
			if test.initialVerbs != nil {
				sa.roles[r] = sets.NewString(test.initialVerbs...)
			}
			sa.MatchRules(r, test.rule)

			if test.expectedVerbs != nil {
				assert.Equal(t, sets.NewString(test.expectedVerbs...), sa.roles[r])
			} else {
				_, ok := sa.roles[r]
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
	mainSubject := toSubject(subject)
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
				subjectAccess: map[SubjectRef]sets.String{mainSubject: sets.NewString("initial-verb")},
				roles:         make(map[RoleRef]sets.String),
			}
			if test.verbsForRole != nil {
				sa.roles[r] = sets.NewString(test.verbsForRole...)
			}

			sa.ResolveRoleRef(r, makeSubjects(test.subjects))

			assert.Equal(t, sets.NewString(test.expectedVerbs...), sa.subjectAccess[mainSubject])
		})
	}
}

func makeSubjects(in []string) []v1.Subject {
	var subjects []v1.Subject
	for _, s := range in {
		subjects = append(subjects, v1.Subject{
			Name: s,
			Kind: "some-kind",
		})
	}
	return subjects
}

func toSubject(name string) SubjectRef {
	return SubjectRef{Name: name, Kind: "some-kind"}
}
