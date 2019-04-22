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
	"github.com/corneliusweig/rakkess/pkg/rakkess/constants"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	clientv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

var (
	// for testing
	getRbacClient = GetRbacClient
)

const (
	clusterRoleName = "ClusterRole"
	roleName        = "Role"
)

// roleRef uniquely identifies a ClusterRole or namespaced Role. The namespace
// is always fixed and need not be part of roleRef to identify a namespaced Role.
type roleRef struct {
	Name, Kind string
}

// SubjectRef uniquely identifies the subject of a RoleBinding or ClusterRoleBinding
type SubjectRef struct {
	Name, Kind string
}

type SubjectAccess struct {
	Resource      string
	roles         map[roleRef]sets.String
	subjectAccess map[SubjectRef]sets.String
}

func GetSubjectAccess(opts *options.RakkessOptions, resource string) (*SubjectAccess, error) {
	rbacClient, err := getRbacClient(opts)
	if err != nil {
		return nil, err
	}

	sa := newSubjectAccess(resource)

	if err := fetchMatchingClusterRoles(rbacClient, sa); err != nil {
		logrus.Warnf("ClusterRoles could not be fetched (%s): result will be incomplete", err)
	} else if err := resolveClusterRoleBindings(rbacClient, sa); err != nil {
		logrus.Warnf("ClusterRolesBindings could not be fetched (%s): result will be incomplete", err)
	}

	namespace := opts.ConfigFlags.Namespace
	if err := fetchMatchingRoles(rbacClient, sa, namespace); err != nil {
		return nil, errors.Wrapf(err, "fetching Roles failed")
	}
	if err := resolveRoleBindings(rbacClient, sa, namespace); err != nil {
		return nil, errors.Wrapf(err, "fetching RoleBindings failed")
	}

	return sa, nil
}

func resolveRoleBindings(rbacClient clientv1.RoleBindingsGetter, sa *SubjectAccess, namespace *string) error {
	if namespace == nil || *namespace == "" {
		logrus.Debugf("Skipping role resolution because namespace not set")
		return nil
	}

	logrus.Debugf("fetching RoleBindings for namespace %s", *namespace)
	roleBindings, err := rbacClient.RoleBindings(*namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, rb := range roleBindings.Items {
		r := roleRef{
			Name: rb.RoleRef.Name,
			Kind: rb.RoleRef.Kind,
		}
		sa.resolveRoleRef(r, rb.Subjects)
	}
	return nil
}

func resolveClusterRoleBindings(rbacClient clientv1.ClusterRoleBindingsGetter, sa *SubjectAccess) error {
	logrus.Debugf("fetching ClusterRoleBindings")
	clusterRoleBindings, err := rbacClient.ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, crb := range clusterRoleBindings.Items {
		r := roleRef{
			Name: crb.RoleRef.Name,
			Kind: crb.RoleRef.Kind,
		}
		sa.resolveRoleRef(r, crb.Subjects)
	}
	return nil
}

func fetchMatchingClusterRoles(rbacClient clientv1.ClusterRolesGetter, sa *SubjectAccess) error {
	logrus.Debugf("fetching clusterRoles")
	roleList, err := rbacClient.ClusterRoles().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	logrus.Tracef("roles: %s", roleList)

	for _, role := range roleList.Items {
		r := roleRef{
			Name: role.Name,
			Kind: clusterRoleName,
		}
		for _, rule := range role.Rules {
			sa.matchRules(r, rule)
		}
	}
	return nil
}

func fetchMatchingRoles(rbacClient clientv1.RolesGetter, sa *SubjectAccess, namespace *string) error {
	if namespace == nil || *namespace == "" {
		logrus.Debugf("Skipping role fetching because namespace not set")
		return nil
	}

	logrus.Debugf("fetching roles for namespace %s", *namespace)
	roleList, err := rbacClient.Roles(*namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, role := range roleList.Items {
		r := roleRef{
			Name: role.Name,
			Kind: roleName,
		}
		for _, rule := range role.Rules {
			sa.matchRules(r, rule)
		}
	}
	return nil
}

func newSubjectAccess(resource string) *SubjectAccess {
	return &SubjectAccess{
		Resource:      resource,
		roles:         make(map[roleRef]sets.String),
		subjectAccess: make(map[SubjectRef]sets.String),
	}
}

func (sa *SubjectAccess) Get() map[SubjectRef]sets.String {
	return sa.subjectAccess
}

func (sa *SubjectAccess) resolveRoleRef(r roleRef, subjects []rbacv1.Subject) {
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

func (sa *SubjectAccess) matchRules(r roleRef, rule rbacv1.PolicyRule) {
	for _, resource := range rule.Resources {
		if resource == rbacv1.ResourceAll || resource == sa.Resource {
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
		if verb == rbacv1.VerbAll {
			return constants.ValidVerbs
		}
	}
	return verbs
}

func GetRbacClient(o *options.RakkessOptions) (clientv1.RbacV1Interface, error) {
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return clientv1.NewForConfigOrDie(restConfig), nil
}
