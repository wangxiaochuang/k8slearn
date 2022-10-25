package serviceaccount

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clientset "k8s.io/client-go/kubernetes"
	v1listers "k8s.io/client-go/listers/core/v1"
	"k8s.io/kubernetes/pkg/serviceaccount"
)

type clientGetter struct {
	client               clientset.Interface
	secretLister         v1listers.SecretLister
	serviceAccountLister v1listers.ServiceAccountLister
	podLister            v1listers.PodLister
}

func NewGetterFromClient(c clientset.Interface, secretLister v1listers.SecretLister, serviceAccountLister v1listers.ServiceAccountLister, podLister v1listers.PodLister) serviceaccount.ServiceAccountTokenGetter {
	return clientGetter{c, secretLister, serviceAccountLister, podLister}
}

func (c clientGetter) GetServiceAccount(namespace, name string) (*v1.ServiceAccount, error) {
	if serviceAccount, err := c.serviceAccountLister.ServiceAccounts(namespace).Get(name); err == nil {
		return serviceAccount, nil
	}
	return c.client.CoreV1().ServiceAccounts(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c clientGetter) GetPod(namespace, name string) (*v1.Pod, error) {
	if pod, err := c.podLister.Pods(namespace).Get(name); err == nil {
		return pod, nil
	}
	return c.client.CoreV1().Pods(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}

func (c clientGetter) GetSecret(namespace, name string) (*v1.Secret, error) {
	if secret, err := c.secretLister.Secrets(namespace).Get(name); err == nil {
		return secret, nil
	}
	return c.client.CoreV1().Secrets(namespace).Get(context.TODO(), name, metav1.GetOptions{})
}
