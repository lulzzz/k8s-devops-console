package services

import (
	"os"
	"fmt"
	"errors"
	"k8s-devops-console/app"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	v12 "k8s.io/api/rbac/v1"
)

type Kubernetes struct {

}

func (k *Kubernetes) homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func (k *Kubernetes) Client() (clientset *kubernetes.Clientset) {
	var err error
	var config *rest.Config

	if kubeconfig := os.Getenv("KUBECONFIG"); kubeconfig != "" {
		// KUBECONFIG
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			panic(err.Error())
		}
	} else {
		// K8S in cluster
		config, err = rest.InClusterConfig()
		if err != nil {
			panic(err.Error())
		}
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	return clientset
}

func (k *Kubernetes) NamespaceList() (nsList map[string]v1.Namespace, err error) {
	result, err := k.Client().CoreV1().Namespaces().List(metav1.ListOptions{})

	if err == nil {
		nsList = make(map[string]v1.Namespace, len(result.Items))
		for _, ns := range result.Items {
			if err := k.namespaceValidate(ns.Name); err == nil {
				nsList[ns.Name] = ns
			}
		}
	}

	return
}

func (k *Kubernetes) Nodes() (*v1.NodeList, error) {
	opts := metav1.ListOptions{}
	return k.Client().CoreV1().Nodes().List(opts)
}

func (k *Kubernetes) NamespaceGet(namespace string) (*v1.Namespace, error) {
	if err := k.namespaceValidate(namespace); err != nil {
		return nil, err
	}

	opts := metav1.GetOptions{}
	return k.Client().CoreV1().Namespaces().Get(namespace, opts)
}

func (k *Kubernetes) NamespaceCreate(namespace v1.Namespace) (error) {
	if err := k.namespaceValidate(namespace.Name); err != nil {
		return err
	}

	_, err := k.Client().CoreV1().Namespaces().Create(&namespace)

	return err
}

func (k *Kubernetes) NamespaceDelete(namespace string) (error) {
	if err := k.namespaceValidate(namespace); err != nil {
		return err
	}

	opts := metav1.DeleteOptions{}
	return k.Client().CoreV1().Namespaces().Delete(namespace, &opts)
}

func (k *Kubernetes) RoleBindingCreateNamespaceUser(namespace, user string) (roleBinding *v12.RoleBinding, error error) {
	roleBindName := fmt.Sprintf("user:%s", user)

	getOpts := metav1.GetOptions{}
	if rb, _ := k.Client().RbacV1().RoleBindings(namespace).Get(roleBindName, getOpts); rb != nil && rb.GetUID() != "" {
		deleteOpts := metav1.DeleteOptions{}
		k.Client().RbacV1().RoleBindings(namespace).Delete(roleBindName, &deleteOpts)
	}

	annotiations := map[string]string{}
	annotiations["user"] = user

	subject := v12.Subject{}
	subject.Name = user
	subject.Kind = "User"

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = "admin"

	roleBinding = &v12.RoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.SetNamespace(namespace)
	roleBinding.RoleRef = role
	roleBinding.Subjects = []v12.Subject{subject}

	return k.Client().RbacV1().RoleBindings(namespace).Create(roleBinding)
}

func (k *Kubernetes) RoleBindingCreateNamespaceGroup(namespace, group string) (roleBinding *v12.RoleBinding, error error) {
	roleBindName := fmt.Sprintf("group:%s", group)

	getOpts := metav1.GetOptions{}
	if rb, _ := k.Client().RbacV1().RoleBindings(namespace).Get(roleBindName, getOpts); rb != nil && rb.GetUID() != "" {
		deleteOpts := metav1.DeleteOptions{}
		k.Client().RbacV1().RoleBindings(namespace).Delete(roleBindName, &deleteOpts)
	}

	annotiations := map[string]string{}
	annotiations["group"] = group

	subject := v12.Subject{}
	subject.Name = group
	subject.Kind = "Group"

	role := v12.RoleRef{}
	role.Kind = "ClusterRole"
	role.Name = "admin"

	roleBinding = &v12.RoleBinding{}
	roleBinding.SetAnnotations(annotiations)
	roleBinding.SetName(roleBindName)
	roleBinding.SetNamespace(namespace)
	roleBinding.RoleRef = role
	roleBinding.Subjects = []v12.Subject{subject}

	return k.Client().RbacV1().RoleBindings(namespace).Create(roleBinding)
}

func (k *Kubernetes) namespaceValidate(name string) (err error) {
	if ! app.RegexpNamespaceFilter.MatchString(name) {
		err = errors.New(fmt.Sprintf("Namespace %s not allowed", name))
	}

	return
}

