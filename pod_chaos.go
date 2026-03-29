package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"log"
	"math/rand"
	"time"
	// "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type PodFilterOptions struct {
	Name string `json:"name"`
	Namespace string `json:"namespace"`
	Labels map[string]string `json:"labels,omitempty"`
}

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
		fmt.Printf("%v: %v\n    - Namespace: %v\n    - Generation: %v\n    - Labels: %v\n", i+1, pod.Name, pod.Namespace, pod.Generation, pod.Labels)
	}
	fmt.Print("\n")
}
// Helper Function for Matching Pod Labels to pods
func matchesLabels(podLabels map[string]string, required map[string]string) bool {
    for key, value := range required {
        if podLabels[key] != value {
            return false
        }
    }
    return true
}
// Filter pods based on attributes
func filterPods(pods []corev1.Pod, opts PodFilterOptions) []corev1.Pod {
	filteredPods := []corev1.Pod{}
	for _, pod := range pods {
		if opts.Name != "" && pod.Name != opts.Name {
			continue
		}
		if opts.Namespace != "" && pod.Namespace != opts.Namespace {
			continue
		}
		if opts.Labels != nil && !matchesLabels(pod.Labels, opts.Labels) {
            continue
        }
		filteredPods = append(filteredPods, pod)
	}
	return filteredPods
}
// Delete a single pod
func deletePod(ctx context.Context, clientset *kubernetes.Clientset, pod corev1.Pod, options metav1.DeleteOptions) error {
	podName := pod.Name
	err := clientset.CoreV1().Pods(pod.Namespace).Delete(ctx, pod.Name, options)
	if err != nil {
		log.Fatalf("Failed to delete Pod %v", podName)
	}

	return nil
}
// Kill several pods
func KillPods(clientset *kubernetes.Clientset, pods []corev1.Pod, percentage float64, opts PodFilterOptions) []corev1.Pod {
	validPods := filterPods(pods, opts)
	listLength := len(validPods)
	numPodsInKillPool := int(percentage * float64(listLength))
	killed := []corev1.Pod{}

	rand.Shuffle(len(validPods), func(i, j int) {
		validPods[i], validPods[j] = validPods[j], validPods[i]
	}) 

	gracePeriod := int64(0)
	options := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	for i := 0; i < numPodsInKillPool; i++ {
		pod := validPods[i]
		err := deletePod(context.TODO(), clientset, pod, options)
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
	
	namespace := ""
	pods, err := GetPodListInNamespace(namespace, clientset)
	if err != nil {
		log.Fatalf("Error returned from GetPodListInNamespace: %v", err)
	}
	PrintPodList(pods)

	opts := PodFilterOptions{
		Name: "",
		Namespace: "lokilab",
		Labels: nil,
	}
	percentage := 0.4
	KillPods(clientset, pods, percentage, opts)
	time.Sleep(60 * time.Second)
	pods, err = GetPodListInNamespace(namespace, clientset)
	if err != nil {
		log.Fatalf("Error returned from GetPodListInNamespace: %v", err)
	}
	PrintPodList(pods)
}
