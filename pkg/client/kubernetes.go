package client

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/4rivappa/kubectl-ip-check/pkg/types"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubernetesClient struct {
	clientset *kubernetes.Clientset
}

func NewKubernetesClient() (*KubernetesClient, error) {
	var config *rest.Config
	var err error

	config, err = rest.InClusterConfig()
	if err != nil {
		// If in-cluster config fails, try kubeconfig
		kubeconfig := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		if kubeconfigPath := os.Getenv("KUBECONFIG"); kubeconfigPath != "" {
			kubeconfig = kubeconfigPath
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get Kubernetes config: %v", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %v", err)
	}

	return &KubernetesClient{clientset: clientset}, nil
}

func (kc *KubernetesClient) GetNodes(ctx context.Context) ([]types.Node, error) {
	nodes, err := kc.clientset.CoreV1().Nodes().List(ctx, v1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list nodes: %v", err)
	}

	var extractedNodes []types.Node
	for _, node := range nodes.Items {
		instanceID, availabilityZone := extractAWSProviderInfo(node.Spec.ProviderID)
		instanceType := extractInstanceType(node.Labels)
		internalIP := extractInternalIP(node.Status.Addresses)

		extractedNodes = append(extractedNodes, types.Node{
			Name:             node.Name,
			InstanceID:       instanceID,
			InstanceType:     instanceType,
			AvailabilityZone: availabilityZone,
			InternalIP:       internalIP,
			TotalIPs:         0,
			TotalENIs:        0,
			UsedIPs:          0,
			FreeIPs:          0,
		})
	}
	return extractedNodes, nil
}

func (kc *KubernetesClient) CountPodIPsOnNode(ctx context.Context, node types.Node) (int, error) {
	pods, err := kc.clientset.CoreV1().Pods("").List(ctx, v1.ListOptions{
		FieldSelector: "spec.nodeName=" + node.Name,
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list pods for node %s: %v", node.Name, err)
	}

	usedIPCount := 0

	for _, pod := range pods.Items {
		if pod.Spec.HostNetwork {
			continue
		}

		for _, podIP := range pod.Status.PodIPs {
			if podIP.IP != "" && podIP.IP != node.InternalIP {
				usedIPCount++
			}
		}
	}

	return usedIPCount, nil
}

func extractAWSProviderInfo(providerID string) (instanceID string, availabilityZone string) {
	if providerID == "" {
		return "", ""
	}

	lastSlashIndex := strings.LastIndex(providerID, "/")
	if lastSlashIndex == -1 || lastSlashIndex == len(providerID)-1 {
		return "", ""
	}

	instanceID = providerID[lastSlashIndex+1:]
	if !strings.HasPrefix(instanceID, "i-") {
		return "", ""
	}

	beforeLastSlash := providerID[:lastSlashIndex]
	secondLastSlashIndex := strings.LastIndex(beforeLastSlash, "/")
	if secondLastSlashIndex == -1 {
		return "", ""
	}

	availabilityZone = providerID[secondLastSlashIndex+1 : lastSlashIndex]

	return instanceID, availabilityZone
}

func extractInstanceType(labels map[string]string) string {
	if labels == nil {
		return ""
	}

	if val, exists := labels["node.kubernetes.io/instance-type"]; exists {
		return val
	}
	return ""
}

func extractInternalIP(addresses []corev1.NodeAddress) string {
	for _, addr := range addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return ""
}
