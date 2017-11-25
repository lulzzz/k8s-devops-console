package services

import (
	"os"
	"fmt"
	"errors"
	"path/filepath"
	"k8s-management/app"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/api/core/v1"
)

type Kubernetes struct {

}

func (k *Kubernetes) homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

func (k *Kubernetes) Client() (*kubernetes.Clientset) {
	kubeconfig := filepath.Join(k.homeDir(), ".kube", "config")

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// creates the in-cluster config
	/*
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	*/

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
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

func (k *Kubernetes) NamespaceDelete(name string) (error) {
	if err := k.namespaceValidate(name); err != nil {
		return err
	}

	opts := metav1.DeleteOptions{}
	return k.Client().CoreV1().Namespaces().Delete(name, &opts)
}

func (k *Kubernetes) namespaceValidate(name string) (err error) {
	if ! app.RegexpNamespaceFilter.MatchString(name) {
		err = errors.New(fmt.Sprintf("Namespace %s not allowed", name))
	}

	return
}