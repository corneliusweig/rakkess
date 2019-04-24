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
	"bytes"
	"sort"
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

func TestSubjectAccess_Print(t *testing.T) {
	yesNoConverter := func(i int) string {
		if i == AccessAllowed {
			return "yes"
		}
		return "no"
	}
	tests := []struct {
		name          string
		subjectAccess map[SubjectRef]sets.String
		verbs         []string
		expected      string
	}{
		{
			name: "single row with multiple verbs",
			subjectAccess: map[SubjectRef]sets.String{
				{Name: "default", Kind: "service-account", Namespace: "some-ns"}: sets.NewString("list", "delete"),
			},
			verbs:    []string{"list", "get"},
			expected: "NAME\tKIND\tSA-NAMESPACE\tLIST\tGET\ndefault\tservice-account\tsome-ns\tyes\tno\n",
		},
		{
			name: "multiple rows with multiple verbs",
			subjectAccess: map[SubjectRef]sets.String{
				{Name: "c-default", Kind: "SA-c", Namespace: "ns-c"}: sets.NewString("get", "delete"),
				{Name: "b-default", Kind: "SA-b", Namespace: "ns-b"}: sets.NewString("list", "get"),
				{Name: "a-default", Kind: "SA-a", Namespace: "ns-a"}: sets.NewString("list", "delete"),
			},
			verbs:    []string{"list", "get"},
			expected: "NAME\tKIND\tSA-NAMESPACE\tLIST\tGET\na-default\tSA-a\tns-a\tyes\tno\nb-default\tSA-b\tns-b\tyes\tyes\nc-default\tSA-c\tns-c\tno\tyes\n",
		},
		{
			name: "ignore row without matches",
			subjectAccess: map[SubjectRef]sets.String{
				{Name: "c-default", Kind: "SA-c", Namespace: "ns-c"}: sets.NewString("get", "delete"),
				{Name: "b-default", Kind: "SA-b", Namespace: "ns-b"}: sets.NewString("delete", "update"),
				{Name: "a-default", Kind: "SA-a", Namespace: "ns-a"}: sets.NewString("list", "delete"),
			},
			verbs:    []string{"list", "get"},
			expected: "NAME\tKIND\tSA-NAMESPACE\tLIST\tGET\na-default\tSA-a\tns-a\tyes\tno\nc-default\tSA-c\tns-c\tno\tyes\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			sa := SubjectAccess{subjectAccess: test.subjectAccess}
			sa.Print(buf, yesNoConverter, test.verbs)

			assert.Equal(t, test.expected, buf.String())
		})
	}
}

func TestSortableSubjects(t *testing.T) {
	tests := []struct {
		name   string
		input  []SubjectRef
		sorted []SubjectRef
	}{
		{
			name:   "two inputs",
			input:  []SubjectRef{{Name: "b"}, {Name: "a"}},
			sorted: []SubjectRef{{Name: "a"}, {Name: "b"}},
		},
		{
			name:   "three inputs",
			input:  []SubjectRef{{Name: "b"}, {Name: "c"}, {Name: "a"}},
			sorted: []SubjectRef{{Name: "a"}, {Name: "b"}, {Name: "c"}},
		},
		{
			name:   "fallback to kind",
			input:  []SubjectRef{{Name: "a", Kind: "c"}, {Name: "a", Kind: "a"}, {Name: "a", Kind: "b"}},
			sorted: []SubjectRef{{Name: "a", Kind: "a"}, {Name: "a", Kind: "b"}, {Name: "a", Kind: "c"}},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			sort.Stable(sortableSubjects(test.input))
			assert.Equal(t, test.sorted, test.input)
		})
	}
}

func makeSubjects(in []string) []v1.Subject {
	subjects := make([]v1.Subject, 0, len(in))
	for _, s := range in {
		subjects = append(subjects, v1.Subject{
			Name:      s,
			Kind:      "some-kind",
			Namespace: "some-ns",
		})
	}
	return subjects
}
func toSubject(name string) SubjectRef {
	return SubjectRef{Name: name, Kind: "some-kind", Namespace: "some-ns"}
}
