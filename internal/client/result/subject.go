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
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/internal/constants"
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
	// roles holds all rule data concerning this resource and is extracted from Roles and ClusterRoles.
	roles map[RoleRef]sets.String
	// subjectAccess holds all subject access data for this resource and is extracted from RoleBindings and ClusterRoleBindings.
	subjectAccess map[SubjectRef]sets.String
}

// NewSubjectAccess creates a new SubjectAccess with initialized fields.
func NewSubjectAccess(resource, resourceName string) *SubjectAccess {
	return &SubjectAccess{
		Resource:      resource,
		ResourceName:  resourceName,
		roles:         make(map[RoleRef]sets.String),
		subjectAccess: make(map[SubjectRef]sets.String),
	}
}

// Get provides access to the actual result (for testing).
func (sa *SubjectAccess) Get() map[SubjectRef]sets.String {
	return sa.subjectAccess
}

// Empty checks if any subjects with access were found.
func (sa *SubjectAccess) Empty() bool {
	return len(sa.subjectAccess) == 0
}

// ResolveRoleRef takes a RoleRef and a list of subjects and stores the access
// rights of the given role for each subject. The RoleRef and subjects usually
// come from a (Cluster)RoleBinding.
func (sa *SubjectAccess) ResolveRoleRef(r RoleRef, subjects []v1.Subject) {
	verbsForRole, ok := sa.roles[r]
	if !ok {
		return
	}
	for _, subject := range subjects {
		s := SubjectRef{
			Name:      subject.Name,
			Kind:      subject.Kind,
			Namespace: subject.Namespace,
		}
		if verbs, ok := sa.subjectAccess[s]; ok {
			sa.subjectAccess[s] = verbs.Union(verbsForRole)
		} else {
			sa.subjectAccess[s] = verbsForRole
		}
	}
}

// MatchRules takes a RoleRef and a PolicyRule and adds the rule verbs to the
// allowed verbs for the RoleRef, if the sa.resource matches the rule.
// The RoleRef and rule usually come from a (Cluster)Role.
func (sa *SubjectAccess) MatchRules(r RoleRef, rule v1.PolicyRule) {
	if len(rule.ResourceNames) > 0 && !includes(rule.ResourceNames, sa.ResourceName) {
		return
	}

	for _, resource := range rule.Resources {
		if resource == v1.ResourceAll || resource == sa.Resource {
			expandedVerbs := expandVerbs(rule.Verbs)
			if verbs, ok := sa.roles[r]; ok {
				sa.roles[r] = sets.NewString(expandedVerbs...).Union(verbs)
			} else {
				sa.roles[r] = sets.NewString(expandedVerbs...)
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

func expandVerbs(verbs []string) []string {
	for _, verb := range verbs {
		if verb == v1.VerbAll {
			return constants.ValidVerbs
		}
	}
	return verbs
}

// Print implements MatrixPrinter.Print. It prints a tab-separated table with a header.
func (sa *SubjectAccess) Print(w io.Writer, converter CodeConverter, requestedVerbs []string) {
	// table header
	fmt.Fprint(w, "NAME\tKIND\tSA-NAMESPACE")
	for _, v := range requestedVerbs {
		fmt.Fprintf(w, "\t%s", strings.ToUpper(v))
	}
	fmt.Fprint(w, "\n")

	subjects := make([]SubjectRef, 0, len(sa.subjectAccess))
	for s := range sa.subjectAccess {
		subjects = append(subjects, s)
	}
	sort.Stable(sortableSubjects(subjects))

	// table body
	for _, subject := range subjects {
		verbs := sa.subjectAccess[subject]
		if !verbs.HasAny(requestedVerbs...) {
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s", subject.Name, subject.Kind, subject.Namespace)
		for _, v := range requestedVerbs {
			var code int
			if verbs.Has(v) {
				code = AccessAllowed
			} else {
				code = AccessDenied
			}
			fmt.Fprintf(w, "\t%s", converter(code))
		}
		fmt.Fprint(w, "\n")
	}
}

type sortableSubjects []SubjectRef

func (s sortableSubjects) Len() int      { return len(s) }
func (s sortableSubjects) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
func (s sortableSubjects) Less(i, j int) bool {
	ret := strings.Compare(s[i].Name, s[j].Name)
	if ret > 0 {
		return false
	} else if ret == 0 {
		ret = strings.Compare(s[i].Kind, s[j].Kind)
		if ret > 0 {
			return false
		} else if ret == 0 {
			return i < j
		}
	}
	return true
}
