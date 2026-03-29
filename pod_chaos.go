package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"log"
	"time"
	"math/rand"

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
		log.Fatal("Failed to print pods in PrintPodList.")
	}
	for i, pod := range pods {
		fmt.Printf("%v: %v\n", i, pod.Name)
	}
}

func deletePod(clientset *kubernetes.Clientset, ctx context.Context, pod corev1.Pod, options metav1.DeleteOptions) error {
	podName := pod.Name
	err := clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, options)
	if err != nil {
		log.Fatalf("Failed to delete Pod %v", podName)
	}

	return nil
}

func KillPods(clientset *kubernetes.Clientset, pods []corev1.Pod, numTargets int, namespace string) []corev1.Pod {
	filteredPods := []corev1.Pod{}
	for _, pod := range pods {
		if pod.Namespace == namespace {
			filteredPods = append(filteredPods, pod)
		}
	}
	podListLength := len(filteredPods)
	numPodsInKillPool := min(podListLength, numTargets)

	killed := []corev1.Pod{}

	rand.Shuffle(len(filteredPods), func(i, j int) {
		filteredPods[i], filteredPods[j] = filteredPods[j], filteredPods[i]
	}) 

	gracePeriod := int64(0)
	options := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	for i := 0; i < numPodsInKillPool; i++ {
		pod := filteredPods[i]
		err := deletePod(clientset, context.TODO(), pod, options)
		if err != nil {
			fmt.Printf("failed to delete pod %s: %v\n", pod.Name, err)
			continue
		}
		killed = append(killed, pod)
	}

	return killed
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
	time.Sleep(5 * time.Second)
	KillPods(clientset, pods, 2, namespace)
	time.Sleep(30 * time.Second)
	newPods, err := GetPodListInNamespace(namespace, clientset)
	if err != nil {
		log.Fatalf("Error returned from GetPodListInNamespace: %v", err)
	}
	PrintPodList(newPods)
}
