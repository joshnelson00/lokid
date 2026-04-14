package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	// "math/rand"
	"time"
	// "k8s.io/apimachinery/pkg/api/errors"
	corev1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

type NodeFilterOptions struct {
	Name        string            `json:"name,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`

	Ready         *bool `json:"ready,omitempty"`
	Unschedulable *bool `json:"unschedulable,omitempty"`

	Roles []string `json:"roles,omitempty"`

	ExcludeNames  []string          `json:"excludeNames,omitempty"`
	ExcludeLabels map[string]string `json:"excludeLabels,omitempty"`
}

func GetNodeList(ctx context.Context, clientset *kubernetes.Clientset, opts metav1.ListOptions) ([]corev1.Node, error) {
	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %w", err)
	}
	return nodes.Items, nil
}

func PrintNodeList(nodes []corev1.Node) {
	if nodes == nil {
		return
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

func MatchesMap(nodeMap map[string]string, required map[string]string) bool {
	for key, value := range required {
		if nodeMap[key] == value {
			return true
		}
	}
	return false
}

func MatchesSlice(nodeSlice []string, required []string) bool {
	for i := range required {
		if nodeSlice[i] == required[i] {
			return true
		}
	}
	return false
}

func IsExcludedName(nodeName string, names []string) (bool) {
	for _, name := range names {
		if nodeName == name {
			return true
		} else {
			continue
		}
	}
	return false
}

func FilterNodes(nodes []corev1.Node, opts NodeFilterOptions) {
	/*
		Name            string            `json:"name,omitempty"`
		Labels          map[string]string `json:"labels,omitempty"`
		Annotations     map[string]string `json:"annotations,omitempty"`

		Ready           *bool             `json:"ready,omitempty"`
		Unschedulable   *bool             `json:"unschedulable,omitempty"`

		Roles           []string          `json:"roles,omitempty"`

		ExcludeNames    []string          `json:"excludeNames,omitempty"`
		ExcludeLabels   map[string]string `json:"excludeLabels,omitempty"`
	*/

		for _, node := range nodes {
		// Positive name filter
		if opts.Name != "" && node.Name != opts.Name {
			continue
		}
		// Exclusion by name
		if IsExcludedName(node.Name, opts.ExcludeNames) {
			continue
		}
		// Label filter
		if len(opts.Labels) > 0 && !MatchesMap(node.Labels, opts.Labels) {
			continue
		}
		// Annotation filter
		if len(opts.Annotations) > 0 && !MatchesMap(node.Annotations, opts.Annotations) {
			continue
		}
		// Unschedulable filter
		if opts.Unschedulable != nil && *opts.Unschedulable != node.Spec.Unschedulable {
			continue
		}
		// Ready filter -- must read from Status.Conditions
		if opts.Ready != nil && !MatchesReadyCondition(node, *opts.Ready) {
			continue
		}
		// Roles filter -- roles are stored as labels: node-role.kubernetes.io/<role>
		if len(opts.Roles) > 0 && !MatchesRoles(node.Labels, opts.Roles) {
			continue
		}
		// Exclude by label -- skip if node matches ANY excluded label
		if len(opts.ExcludeLabels) > 0 && MatchesMap(node.Labels, opts.ExcludeLabels) {
			continue
		}

		filtered = append(filtered, node)
	}
}


func DeleteNode(clientset *kubernetes.Clientset, node corev1.Node) corev1.Node {
	gracePeriod := int64(0)
	opts := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}
	err := clientset.CoreV1().Nodes().Delete(context.TODO(), node.Name, opts)
	if err != nil {
		fmt.Printf("Couldn't delete node %v", node.Name)
	}

	return node
}

func CordonNode(ctx context.Context, clientset *kubernetes.Clientset, node corev1.Node, opts metav1.UpdateOptions) (*corev1.Node, error) {
	node.Spec.Unschedulable = true
	updatedNode, err := clientset.CoreV1().Nodes().Update(context.TODO(), &node, opts)
	if err != nil {
		return nil, err
	}

	return updatedNode, nil
}

func UncordonNode(ctx context.Context, clientset *kubernetes.Clientset, node corev1.Node, opts metav1.UpdateOptions) (*corev1.Node, error) {
	node.Spec.Unschedulable = false
	updatedNode, err := clientset.CoreV1().Nodes().Update(context.TODO(), &node, opts)
	if err != nil {
		return nil, err
	}

	return updatedNode, nil
}

func DrainNode(ctx context.Context, clientset *kubernetes.Clientset, node corev1.Node) error {
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
	})
	if err != nil {
		fmt.Printf("Couldn't get pods for draining in node %v", node.Name)
		return err
	}

	err = EvictPodsFromNode(clientset, pods.Items)
	if err != nil {
		fmt.Printf("Couldn't exict pods in node %v", node.Name)
		return err
	}

	return nil
}

func EvictPodsFromNode(clientset *kubernetes.Clientset, pods []corev1.Pod) error {
	for _, pod := range pods {
		eviction := policyv1.Eviction{
			ObjectMeta: metav1.ObjectMeta{
				Name:      pod.Name,
				Namespace: pod.Namespace,
			},
		}
		clientset.CoreV1().Pods(pod.Namespace).EvictV1(context.TODO(), &eviction)
	}
	return nil
}

func SelectRandomNodes() {
}

func ApplyNodeChaos() {
	return
}

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		kubeconfigPath := filepath.Join(homedir.HomeDir(), ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfigPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to load Kubernetes config: %v\n", err)
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

	ctx := context.TODO()
	listOpts := metav1.ListOptions{}

	// --- Before ---
	fmt.Println("=== Nodes Before Drain ===")
	nodes, err := GetNodeList(ctx, clientset, listOpts)
	if err != nil {
		log.Fatalf("GetNodeList failed: %v", err) // this will show the real reason
	}
	PrintNodeList(nodes)

	// --- Cordon + Drain first node ---
	target := nodes[0]
	fmt.Printf(">>> Cordoning node: %s\n", target.Name)
	cordonedNode, err := CordonNode(ctx, clientset, target, metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("cordoning failed: %v", err)
	}
	target = *cordonedNode

	fmt.Printf(">>> Draining node: %s\n", target.Name)
	err = DrainNode(ctx, clientset, target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "drain failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Waiting 30s for pods to reschedule...")
	time.Sleep(30 * time.Second)

	// --- After ---
	fmt.Println("=== Nodes After Drain ===")
	nodes, err = GetNodeList(ctx, clientset, listOpts)
	if err != nil {
		os.Exit(1)
	}
	PrintNodeList(nodes)

	uncordonedNode, err := UncordonNode(ctx, clientset, nodes[0], metav1.UpdateOptions{})
	if err != nil {
		log.Fatalf("uncordoning failed: %v", err)
	}
	fmt.Printf(">>> Uncordoned node: %s\n", uncordonedNode.Name)
}

