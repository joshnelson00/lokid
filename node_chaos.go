package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"log"
	"sort"
	// "math/rand"
	// "time"
	// "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func GetNodeList(ctx context.Context, clientset *kubernetes.Clientset, opts metav1.ListOptions) ([]corev1.Node, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get list of nodes: %v\n", err)
		return nil, err
	}
	return nodes.Items, nil
}

func PrintNodeList(nodes []corev1.Node) {
	if nodes == nil {
		log.Fatal("Failed to print nodes in PrintNodeList.")
	}

	for i, node := range nodes {
		fmt.Printf("%v: %v\n", i+1, node.Name)
		fmt.Printf("    - Generation: %v\n", node.Generation)

		fmt.Println("    - Labels:")
		if len(node.Labels) == 0 {
			fmt.Println("        (none)")
		} else {
			keys := make([]string, 0, len(node.Labels))
			for k := range node.Labels {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				fmt.Printf("        %s: %s\n", k, node.Labels[k])
			}
		}
	}

	fmt.Print("\n")
}
func MatchesNodeLabels() {
	return
}

func FilterNodes() {
	return
}

func DeleteNode() {
	return
}

func CordonNode() {
	return
}

func UncordonNode() {
	return
}

func DrainNode() {
	return
}

func EvictPodsFromNode() {
	return
}

func SelectRandomNodes() {
	return
}

func ApplyNodeChaos() {
	return
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
	opts := metav1.ListOptions{}
	nodes, err := GetNodeList(context.TODO(), clientset, opts)
	if err != nil {
		os.Exit(1)
	}
	PrintNodeList(nodes)
}