package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/4rivappa/kubectl-ip-check/pkg/calculator"
	"github.com/4rivappa/kubectl-ip-check/pkg/client"
	"github.com/4rivappa/kubectl-ip-check/pkg/types"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/printers"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "kubectl aws-ip",
		Short: "A kubectl plugin to calculate IP addresses usage in AWS EKS nodes",
		Run:   run,
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func run(cmd *cobra.Command, args []string) {
	k8sClient, err := client.NewKubernetesClient()
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		return
	}

	// Create IP calculator once and reuse it
	ipCalculator, err := calculator.NewIPCalculator()
	if err != nil {
		fmt.Printf("Error creating IP calculator: %v\n", err)
		return
	}

	nodes, err := k8sClient.GetNodes(context.TODO())
	if err != nil {
		fmt.Printf("Error retrieving nodes: %v\n", err)
		return
	}

	totalClusterIPs := 0
	totalUsedIPs := 0
	totalFreeIPs := 0

	processNodesInParallel(nodes, ipCalculator, k8sClient, &totalClusterIPs, &totalUsedIPs, &totalFreeIPs)

	fmt.Printf("Allocated IPs: %d | Used IPs: %d | Free IPs: %d\n", totalClusterIPs, totalUsedIPs, totalFreeIPs)

	printNodesTable(nodes)
}

func processNodesInParallel(nodes []types.Node, ipCalculator *calculator.IPCalculator, k8sClient *client.KubernetesClient, totalClusterIPs, totalUsedIPs, totalFreeIPs *int) {
	const maxWorkers = 5

	var wg sync.WaitGroup
	var mu sync.Mutex

	// Create a channel to limit concurrent workers
	semaphore := make(chan struct{}, maxWorkers)

	for i := range nodes {
		wg.Add(1)
		go func(nodeIndex int) {
			defer wg.Done()

			// Acquire semaphore (limit concurrency)
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			node := &nodes[nodeIndex]

			if node.InstanceID != "" {
				totalIPsAllocated, eniCount, err := ipCalculator.TotalIPsAllocated(node.InstanceID)
				if err != nil {
					fmt.Printf("Error getting IP details for %s: %v\n", node.InstanceID, err)
					return
				}

				podIPCount, err := k8sClient.CountPodIPsOnNode(context.TODO(), *node)
				if err != nil {
					fmt.Printf("Error counting pod IPs on node %s: %v\n", node.Name, err)
					return
				}

				node.TotalIPs = totalIPsAllocated
				node.TotalENIs = eniCount
				node.UsedIPs = podIPCount + eniCount
				node.FreeIPs = totalIPsAllocated - node.UsedIPs

				mu.Lock()
				*totalClusterIPs += totalIPsAllocated
				*totalUsedIPs += node.UsedIPs
				*totalFreeIPs += node.FreeIPs
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
}

func printNodesTable(nodes []types.Node) {
	table := &metav1.Table{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "meta.k8s.io/v1",
			Kind:       "Table",
		},
		ColumnDefinitions: []metav1.TableColumnDefinition{
			{Name: "NAME", Type: "string"},
			{Name: "INSTANCE-ID", Type: "string"},
			{Name: "INSTANCE-TYPE", Type: "string"},
			{Name: "AVAILABILITY-ZONE", Type: "string"},
			{Name: "INTERNAL-IP", Type: "string"},
			{Name: "TOTAL-IPS", Type: "integer"},
			{Name: "USED-IPS", Type: "integer"},
			{Name: "FREE-IPS", Type: "integer"},
		},
	}

	for _, node := range nodes {
		row := metav1.TableRow{
			Cells: []interface{}{
				node.Name,
				node.InstanceID,
				node.InstanceType,
				node.AvailabilityZone,
				node.InternalIP,
				node.TotalIPs,
				node.UsedIPs,
				node.FreeIPs,
			},
		}
		table.Rows = append(table.Rows, row)
	}

	printer := printers.NewTablePrinter(printers.PrintOptions{})
	fmt.Print("\n")
	err := printer.PrintObj(table, os.Stdout)
	if err != nil {
		fmt.Printf("Error printing table: %v\n", err)
	}
}
