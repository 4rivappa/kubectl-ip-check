package calculator

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type IPCalculator struct {
	ec2Client *ec2.Client
}

func NewIPCalculator() (*IPCalculator, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %v", err)
	}

	ec2Client := ec2.NewFromConfig(cfg)

	return &IPCalculator{
		ec2Client: ec2Client,
	}, nil
}

func (calc *IPCalculator) TotalIPsAllocated(instanceID string) (int, int, error) {
	if instanceID == "" {
		return 0, 0, fmt.Errorf("instance ID cannot be empty")
	}

	input := &ec2.DescribeNetworkInterfacesInput{
		Filters: []types.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []string{instanceID},
			},
		},
	}

	result, err := calc.ec2Client.DescribeNetworkInterfaces(context.TODO(), input)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to describe network interfaces for instance %s: %v", instanceID, err)
	}

	totalIPs := 0
	numberOfENIs := len(result.NetworkInterfaces)

	for _, eni := range result.NetworkInterfaces {
		totalIPs += len(eni.PrivateIpAddresses)
		totalIPs += len(eni.Ipv6Addresses)
	}

	// log.Printf("Instance %s has %d ENIs with %d total IP addresses", instanceID, numberOfENIs, totalIPs)
	return totalIPs, numberOfENIs, nil
}
