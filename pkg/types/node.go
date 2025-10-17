package types

type Node struct {
	Name             string
	InstanceID       string
	InstanceType     string
	AvailabilityZone string
	InternalIP       string
	TotalIPs         int
	TotalENIs        int
	UsedIPs          int
	FreeIPs          int
}
