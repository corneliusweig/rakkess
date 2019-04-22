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
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
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
	Name, Kind string
}

type SubjectAccess struct {
	Resource      string
	roles         map[RoleRef]sets.String
	subjectAccess map[SubjectRef]sets.String
}

func NewSubjectAccess(resource string) *SubjectAccess {
	return &SubjectAccess{
		Resource:      resource,
		roles:         make(map[RoleRef]sets.String),
		subjectAccess: make(map[SubjectRef]sets.String),
	}
}

func (sa *SubjectAccess) Get() map[SubjectRef]sets.String {
	return sa.subjectAccess
}

func (sa *SubjectAccess) ResolveRoleRef(r RoleRef, subjects []v1.Subject) {
	verbsForRole, ok := sa.roles[r]
	if !ok {
		return
	}
	for _, subject := range subjects {
		s := SubjectRef{
			Name: subject.Name,
			Kind: subject.Kind,
		}
		if verbs, ok := sa.subjectAccess[s]; ok {
			sa.subjectAccess[s] = verbs.Union(verbsForRole)
		} else {
			sa.subjectAccess[s] = verbsForRole
		}
	}
}

func (sa *SubjectAccess) MatchRules(r RoleRef, rule v1.PolicyRule) {
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

func expandVerbs(verbs []string) []string {
	for _, verb := range verbs {
		if verb == v1.VerbAll {
			return constants.ValidVerbs
		}
	}
	return verbs
}

func (ra *SubjectAccess) Print(w io.Writer, converter CodeConverter, requestedVerbs []string) {
	// table header
	fmt.Fprint(w, "NAME\tKIND")
	for _, v := range requestedVerbs {
		fmt.Fprintf(w, "\t%s", strings.ToUpper(v))
	}
	fmt.Fprint(w, "\n")

	subjects := make([]SubjectRef, 0, len(ra.subjectAccess))
	for s, _ := range ra.subjectAccess {
		subjects = append(subjects, s)
	}
	sort.Stable(sortableSubjects(subjects))

	// table body
	for _, subject := range subjects {
		verbs := ra.subjectAccess[subject]
		fmt.Fprintf(w, "%s\t%s", subject.Name, subject.Kind)
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
