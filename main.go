package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"log"

	// "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

)


// Retrieve Pods in a given namespace
func GetPodListInNamespace(namespace string, clientset *kubernetes.Clientset) ([]corev1.Pod, error) {
	pods, err := clientset.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get pods in namespace %v: %v\n",namespace, err)
		return nil, err
	} else {
		return pods.Items, nil
	}
}
// Print Slice of Pods
func PrintPodList(pods []corev1.Pod) {

	if pods == nil {
		log.Fatalf("Failed to print pods in PrintPodList.")
	}
	for i, pod := range pods {
		fmt.Printf("%v: %v\n", i, pod.Name)
	}
}

func main() {
	// Try in-cluster config first, then fall back to local kubeconfig for local runs.
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load Kubernetes config (in-cluster and local kubeconfig): %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Using local kubeconfig: %s\n", kubeconfigPath)
	} else {
		fmt.Println("Using in-cluster Kubernetes configuration")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create Kubernetes clientset: %v\n", err)
		os.Exit(1)
	}
	
		namespace := "lokilab"
		pods, err := GetPodListInNamespace(namespace, clientset)
		if err != nil {
			log.Fatalf("Error returned from GetPodListInNamespace: %v", err)
		}
		PrintPodList(pods)

}
