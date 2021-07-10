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
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/internal/constants"
	"github.com/corneliusweig/rakkess/internal/printer"
	v1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

// RoleRef uniquely identifies a ClusterRole or namespaced Role. The namespace
// is always fixed and need not be part of RoleRef to identify a namespaced Role.
type RoleRef struct {
	Name, Kind string
}

// SubjectRef uniquely identifies the subject of a RoleBinding or ClusterRoleBinding
type SubjectRef struct {
	Name, Kind, Namespace string
}

// SubjectAccess holds the access information of all subjects for the given resource.
type SubjectAccess struct {
	// Resource is the kubernetes resource of this query.
	Resource string
	// ResourceName is the name of the kubernetes resource instance of this query.
	ResourceName string
	// roleToVerbs holds all rule data concerning this resource and is extracted from Roles and ClusterRoles.
	roleToVerbs map[RoleRef]sets.String
	// subjectToVerbs holds all subject access data for this resource and is extracted from RoleBindings and ClusterRoleBindings.
	subjectToVerbs map[SubjectRef]sets.String
}

// NewSubjectAccess creates a new SubjectAccess with initialized fields.
func NewSubjectAccess(resource, resourceName string) *SubjectAccess {
	return &SubjectAccess{
		Resource:       resource,
		ResourceName:   resourceName,
		roleToVerbs:    make(map[RoleRef]sets.String),
		subjectToVerbs: make(map[SubjectRef]sets.String),
	}
}

// Get provides access to the actual result (for testing).
func (sa *SubjectAccess) Get() map[SubjectRef]sets.String {
	return sa.subjectToVerbs
}

// Empty checks if any subjects with access were found.
func (sa *SubjectAccess) Empty() bool {
	return len(sa.subjectToVerbs) == 0
}

// ResolveRoleRef takes a RoleRef and a list of subjects and stores the access
// rights of the given role for each subject. The RoleRef and subjects usually
// come from a (Cluster)RoleBinding.
func (sa *SubjectAccess) ResolveRoleRef(r RoleRef, subjects []v1.Subject) {
	verbsForRole, ok := sa.roleToVerbs[r]
	if !ok {
		return
	}
	for _, subject := range subjects {
		s := SubjectRef{
			Name:      subject.Name,
			Kind:      subject.Kind,
			Namespace: subject.Namespace,
		}
		if verbs, ok := sa.subjectToVerbs[s]; ok {
			sa.subjectToVerbs[s] = verbs.Union(verbsForRole)
		} else {
			sa.subjectToVerbs[s] = verbsForRole
		}
	}
}

// MatchRules takes a RoleRef and a PolicyRule and adds the rule verbs to the
// allowed verbs for the RoleRef, if the sa.resource matches the rule.
// The RoleRef and rule usually come from a (Cluster)Role.
func (sa *SubjectAccess) MatchRules(ref RoleRef, rule v1.PolicyRule) {
	if len(rule.ResourceNames) > 0 && !includes(rule.ResourceNames, sa.ResourceName) {
		return
	}

	for _, r := range rule.Resources {
		if r == v1.ResourceAll || r == sa.Resource {
			expandedVerbs := expand(rule.Verbs)
			if verbs, ok := sa.roleToVerbs[ref]; ok {
				sa.roleToVerbs[ref] = sets.NewString(expandedVerbs...).Union(verbs)
			} else {
				sa.roleToVerbs[ref] = sets.NewString(expandedVerbs...)
			}
		}
	}
}

func includes(coll []string, x string) bool {
	if x == "" {
		return false
	}
	for _, s := range coll {
		if s == x {
			return true
		}
	}
	return false
}

func expand(verbs []string) []string {
	for _, verb := range verbs {
		if verb == v1.VerbAll {
			return constants.ValidVerbs
		}
	}
	return verbs
}

func (sa *SubjectAccess) ToPrinter(verbs []string) *printer.Printer {
	subjects := make([]SubjectRef, 0, len(sa.subjectToVerbs))
	for s := range sa.subjectToVerbs {
		subjects = append(subjects, s)
	}
	sort.Slice(subjects, func(i, j int) bool {
		comp := strings.Compare(subjects[i].Name, subjects[j].Name)
		if comp == 0 {
			return subjects[i].Kind < subjects[j].Kind
		}
		return comp < 0
	})

	headers := []string{"NAME", "KIND", "SA-NAMESPACE"}
	for _, v := range verbs {
		headers = append(headers, strings.ToUpper(v))
	}
	p := printer.New(headers)

	// table body
	for _, s := range subjects {
		valid := sa.subjectToVerbs[s]
		if !valid.HasAny(verbs...) {
			continue
		}
		var outcomes []printer.Outcome
		for _, v := range verbs {
			o := printer.Down
			if valid.Has(v) {
				o = printer.Up
			}
			outcomes = append(outcomes, o)
		}
		intro := []string{s.Name, s.Kind, s.Namespace}
		p.AddRow(intro, outcomes...)
	}

	return p
}
