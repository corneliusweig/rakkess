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
	"testing"

	"github.com/corneliusweig/rakkess/pkg/rakkess/options"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	clientv1 "k8s.io/client-go/kubernetes/typed/rbac/v1"
	"k8s.io/client-go/kubernetes/typed/rbac/v1/fake"
	k8stesting "k8s.io/client-go/testing"
)

func TestGetSubjectAccess(t *testing.T) {
	tests := []struct {
		name                string
		namespace           string
		resource            string
		clusterRoles        []v1.ClusterRole
		clusterRoleBindings []v1.ClusterRoleBinding
		roles               []v1.Role
		roleBindings        []v1.RoleBinding
		expected            map[SubjectRef]sets.String
	}{
		{
			name:                "cluster-role and role matches",
			namespace:           "some-ns",
			resource:            "deployments",
			clusterRoles:        clusterRoles("clusterrole-1", "deployments", "create"),
			clusterRoleBindings: clusterRoleBindings("clusterrole-1", "test-user"),
			roles:               roles("role-1", "some-ns", "deployments", "list"),
			roleBindings:        roleBindings("role-1", "Role", "test-user"),
			expected: map[SubjectRef]sets.String{
				{Name: "test-user", Kind: "User"}: sets.NewString("create", "list"),
			},
		},
		{
			name:                "cluster-role and role matches, multiple subjects",
			namespace:           "some-ns",
			resource:            "deployments",
			clusterRoles:        clusterRoles("clusterrole-1", "deployments", "create"),
			clusterRoleBindings: clusterRoleBindings("clusterrole-1", "user1", "user2"),
			roles:               roles("role-1", "some-ns", "deployments", "list"),
			roleBindings:        roleBindings("role-1", "Role", "user2", "user3"),
			expected: map[SubjectRef]sets.String{
				{Name: "user1", Kind: "User"}: sets.NewString("create"),
				{Name: "user2", Kind: "User"}: sets.NewString("create", "list"),
				{Name: "user3", Kind: "User"}: sets.NewString("list"),
			},
		},
		{
			name:                "cluster-role and role matches, global scope",
			namespace:           "", // empty namespace means global scope
			resource:            "deployments",
			clusterRoles:        clusterRoles("clusterrole-1", "deployments", "create"),
			clusterRoleBindings: clusterRoleBindings("clusterrole-1", "test-user"),
			roles:               roles("role-1", "some-ns", "deployments", "list"),
			roleBindings:        roleBindings("role-1", "Role", "test-user"),
			expected: map[SubjectRef]sets.String{
				{Name: "test-user", Kind: "User"}: sets.NewString("create"),
			},
		},
		{
			name:         "rolebinding to clusterrole",
			namespace:    "some-ns",
			resource:     "deployments",
			clusterRoles: clusterRoles("clusterrole-1", "deployments", "create"),
			roleBindings: roleBindings("clusterrole-1", "ClusterRole", "test-user"),
			expected: map[SubjectRef]sets.String{
				{Name: "test-user", Kind: "User"}: sets.NewString("create"),
			},
		},
		{
			name:                "bindings for wrong resource",
			namespace:           "some-ns",
			resource:            "deployments",
			clusterRoles:        clusterRoles("clusterrole-1", "configmaps", "create"),
			clusterRoleBindings: clusterRoleBindings("clusterrole-1", "test-user"),
			roles:               roles("role-1", "some-ns", "configmaps", "list"),
			roleBindings:        roleBindings("role-1", "Role", "test-user"),
			expected:            map[SubjectRef]sets.String{},
		},
		{
			name:                "VerbAll role binding",
			namespace:           "some-ns",
			resource:            "configmaps",
			clusterRoles:        clusterRoles("clusterrole-1", "configmaps", "create"),
			clusterRoleBindings: clusterRoleBindings("clusterrole-1", "test-user"),
			roles:               roles("role-1", "some-ns", "configmaps", v1.VerbAll),
			roleBindings:        roleBindings("role-1", "Role", "test-user"),
			expected: map[SubjectRef]sets.String{
				{Name: "test-user", Kind: "User"}: sets.NewString(constants.ValidVerbs...),
			},
		},
		{
			name:                "VerbAll clusterrole binding",
			namespace:           "some-ns",
			resource:            "configmaps",
			clusterRoles:        clusterRoles("clusterrole-1", "configmaps", v1.VerbAll),
			clusterRoleBindings: clusterRoleBindings("clusterrole-1", "test-user"),
			expected: map[SubjectRef]sets.String{
				{Name: "test-user", Kind: "User"}: sets.NewString(constants.ValidVerbs...),
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			fakeRbacClient := &fake.FakeRbacV1{Fake: &k8stesting.Fake{}}
			fakeRbacClient.Fake.AddReactor("list", "roles",
				func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.RoleList{Items: test.roles}, nil
				})
			fakeRbacClient.Fake.AddReactor("list", "rolebindings",
				func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.RoleBindingList{Items: test.roleBindings}, nil
				})
			fakeRbacClient.Fake.AddReactor("list", "clusterroles",
				func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.ClusterRoleList{Items: test.clusterRoles}, nil
				})
			fakeRbacClient.Fake.AddReactor("list", "clusterrolebindings",
				func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
					return true, &v1.ClusterRoleBindingList{Items: test.clusterRoleBindings}, nil
				})

			getRbacClient = func(*options.RakkessOptions) (clientv1.RbacV1Interface, error) {
				return fakeRbacClient, nil
			}
			defer func() { getRbacClient = GetRbacClient }()

			opts := &options.RakkessOptions{
				ConfigFlags: &genericclioptions.ConfigFlags{
					Namespace: &test.namespace,
				},
			}
			sa, err := GetSubjectAccess(opts, test.resource)
			assert.NoError(t, err)
			assert.Equal(t, test.resource, sa.Resource)
			assert.Equal(t, test.expected, sa.Get())
		})
	}
}

func clusterRoles(name, resource string, verbs ...string) []v1.ClusterRole {
	return []v1.ClusterRole{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
			},
			Rules: []v1.PolicyRule{
				{
					Verbs:     verbs,
					Resources: []string{resource},
				},
			},
		},
	}
}

func clusterRoleBindings(clusterRole string, subjects ...string) []v1.ClusterRoleBinding {
	ss := make([]v1.Subject, 0, len(subjects))
	for _, s := range subjects {
		ss = append(ss, v1.Subject{
			Kind: "User",
			Name: s,
		})
	}
	return []v1.ClusterRoleBinding{
		{
			Subjects: ss,
			RoleRef: v1.RoleRef{
				Name: clusterRole,
				Kind: "ClusterRole",
			},
		},
	}
}

func roles(name, namespace, resource string, verbs ...string) []v1.Role {
	return []v1.Role{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Rules: []v1.PolicyRule{
				{
					Verbs:     verbs,
					Resources: []string{resource},
				},
			},
		},
	}
}

func roleBindings(role, kind string, subjects ...string) []v1.RoleBinding {
	ss := make([]v1.Subject, 0, len(subjects))
	for _, s := range subjects {
		ss = append(ss, v1.Subject{
			Kind: "User",
			Name: s,
		})
	}
	return []v1.RoleBinding{
		{
			Subjects: ss,
			RoleRef: v1.RoleRef{
				Name: role,
				Kind: kind,
			},
		},
	}
}
