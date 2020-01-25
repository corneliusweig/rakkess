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
	"github.com/corneliusweig/rakkess/pkg/rakkess/client/result"
	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
)

var (
	// for testing
	getRbacClient = getRbacClientImpl
)

const (
	clusterRoleName = "ClusterRole"
	roleName        = "Role"
)

// GetSubjectAccess determines subjects with access to the given resource.
func GetSubjectAccess(opts *options.RakkessOptions, resource, resourceName string) (*result.SubjectAccess, error) {
	rbacClient, err := getRbacClient(opts)
	if err != nil {
		return nil, err
	}

	namespace := opts.ConfigFlags.Namespace
	isNamespace := namespace != nil && *namespace != ""

	sa := result.NewSubjectAccess(resource, resourceName)

	if err := fetchMatchingClusterRoles(rbacClient, sa); err != nil {
		if !isNamespace {
			return nil, err
		}
		logrus.Warnf("incomplete result: %s", err)
	} else if err := resolveClusterRoleBindings(rbacClient, sa); err != nil {
		if !isNamespace {
			return nil, err
		}
		logrus.Warnf("incomplete result: %s", err)
	}

	if !isNamespace {
		logrus.Debugf("Skipping roles and rolebindings because namespace is missing")
		return sa, nil
	}

	if err := fetchMatchingRoles(rbacClient, sa, *namespace); err != nil {
		return nil, err
	}
	if err := resolveRoleBindings(rbacClient, sa, *namespace); err != nil {
		return nil, err
	}

	return sa, nil
}

func resolveRoleBindings(rbacClient clientv1.RoleBindingsGetter, sa *result.SubjectAccess, namespace string) error {
	logrus.Debugf("fetching RoleBindings for namespace %s", namespace)
	roleBindings, err := rbacClient.RoleBindings(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, rb := range roleBindings.Items {
		r := result.RoleRef{
			Name: rb.RoleRef.Name,
			Kind: rb.RoleRef.Kind,
		}
		sa.ResolveRoleRef(r, rb.Subjects)
	}
	return nil
}

func resolveClusterRoleBindings(rbacClient clientv1.ClusterRoleBindingsGetter, sa *result.SubjectAccess) error {
	logrus.Debugf("fetching ClusterRoleBindings")
	clusterRoleBindings, err := rbacClient.ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, crb := range clusterRoleBindings.Items {
		r := result.RoleRef{
			Name: crb.RoleRef.Name,
			Kind: crb.RoleRef.Kind,
		}
		sa.ResolveRoleRef(r, crb.Subjects)
	}
	return nil
}

func fetchMatchingClusterRoles(rbacClient clientv1.ClusterRolesGetter, sa *result.SubjectAccess) error {
	logrus.Debugf("fetching clusterRoles")
	roleList, err := rbacClient.ClusterRoles().List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, role := range roleList.Items {
		r := result.RoleRef{
			Name: role.Name,
			Kind: clusterRoleName,
		}
		for _, rule := range role.Rules {
			sa.MatchRules(r, rule)
		}
	}
	return nil
}

func fetchMatchingRoles(rbacClient clientv1.RolesGetter, sa *result.SubjectAccess, namespace string) error {
	logrus.Debugf("fetching roles for namespace %s", namespace)
	roleList, err := rbacClient.Roles(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}

	for _, role := range roleList.Items {
		r := result.RoleRef{
			Name: role.Name,
			Kind: roleName,
		}
		for _, rule := range role.Rules {
			sa.MatchRules(r, rule)
		}
	}
	return nil
}

func getRbacClientImpl(o *options.RakkessOptions) (clientv1.RbacV1Interface, error) {
	restConfig, err := o.ConfigFlags.ToRESTConfig()
	if err != nil {
		return nil, err
	}

	return clientv1.NewForConfigOrDie(restConfig), nil
}
