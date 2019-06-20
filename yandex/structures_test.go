package yandex

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"google.golang.org/grpc"

	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1"
	"github.com/yandex-cloud/go-genproto/yandex/cloud/compute/v1/instancegroup"
)

type DiskClientGetter struct {
}

func (r *DiskClientGetter) Get(ctx context.Context, in *compute.GetDiskRequest, opts ...grpc.CallOption) (*compute.Disk, error) {
	return &compute.Disk{
		Id:          "",
		FolderId:    "",
		CreatedAt:   nil,
		Name:        "mock-disk-name",
		Description: "mock-disk-description",
		TypeId:      "network-hdd",
		ZoneId:      "",
		Size:        4 * (1 << 30),
		ProductIds:  nil,
	}, nil
}

func TestExpandLabels(t *testing.T) {
	cases := []struct {
		name     string
		labels   interface{}
		expected map[string]string
	}{
		{
			name: "two tags",
			labels: map[string]interface{}{
				"my_key":       "my_value",
				"my_other_key": "my_other_value",
			},
			expected: map[string]string{
				"my_key":       "my_value",
				"my_other_key": "my_other_value",
			},
		},
		{
			name:     "labels is nil",
			labels:   nil,
			expected: map[string]string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandLabels(tc.labels)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestExpandProductIds(t *testing.T) {
	cases := []struct {
		name       string
		productIds *schema.Set
		expected   []string
	}{
		{
			name: "two product ids",
			productIds: schema.NewSet(schema.HashString, []interface{}{
				"super-product",
				"very-good",
			}),
			expected: []string{
				"super-product",
				"very-good",
			},
		},
		{
			name:       "empty product ids",
			productIds: schema.NewSet(schema.HashString, []interface{}{}),
			expected:   []string{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := expandProductIds(tc.productIds)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceResources(t *testing.T) {
	cases := []struct {
		name      string
		resources *compute.Resources
		expected  []map[string]interface{}
	}{
		{
			name: "cores 1 fraction 100 memory 5 gb",
			resources: &compute.Resources{
				Cores:        1,
				CoreFraction: 100,
				Memory:       5 * (1 << 30),
			},
			expected: []map[string]interface{}{
				{
					"cores":         1,
					"core_fraction": 100,
					"memory":        5.0,
				},
			},
		},
		{
			name: "cores 8 fraction 5 memory 16 gb",
			resources: &compute.Resources{
				Cores:        8,
				CoreFraction: 5,
				Memory:       16 * (1 << 30),
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 5,
					"memory":        16.0,
				},
			},
		},
		{
			name: "cores 2 fraction 20 memory 0.5 gb",
			resources: &compute.Resources{
				Cores:        2,
				CoreFraction: 20,
				Memory:       (1 << 30) / 2,
			},
			expected: []map[string]interface{}{
				{
					"cores":         2,
					"core_fraction": 20,
					"memory":        0.5,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceResources(&compute.Instance{Resources: tc.resources})
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceBootDisk(t *testing.T) {
	cases := []struct {
		name     string
		bootDisk *compute.AttachedDisk
		expected []map[string]interface{}
	}{
		{
			name: "boot disk with diskID",
			bootDisk: &compute.AttachedDisk{
				Mode:       compute.AttachedDisk_READ_WRITE,
				DeviceName: "test-device-name",
				AutoDelete: false,
				DiskId:     "saeque9k",
			},
			expected: []map[string]interface{}{
				{
					"device_name": "test-device-name",
					"auto_delete": false,
					"disk_id":     "saeque9k",
					"mode":        "READ_WRITE",
					"initialize_params": []map[string]interface{}{
						{"snapshot_id": "",
							"name":        "mock-disk-name",
							"description": "mock-disk-description",
							"size":        4,
							"type":        "network-hdd",
							"image_id":    "",
						},
					},
				},
			},
		},
	}

	reducedDiskClient := &DiskClientGetter{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceBootDisk(&compute.Instance{BootDisk: tc.bootDisk}, reducedDiskClient)

			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceNetworkInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		instance   *compute.Instance
		want       []map[string]interface{}
		externalIP string
		internalIP string
		wantErr    bool
	}{
		{
			name: "no nics defined",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{},
			},
			want:       []map[string]interface{}{},
			externalIP: "",
			internalIP: "",
			wantErr:    false,
		},
		{
			name: "one nic with internal address",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &compute.PrimaryAddress{
							Address: "192.168.19.16",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":       1,
					"mac_address": "aa-bb-cc-dd-ee-ff",
					"subnet_id":   "some-subnet-id",
					"ip_address":  "192.168.19.16",
					"nat":         false,
				},
			},
			externalIP: "",
			internalIP: "192.168.19.16",
			wantErr:    false,
		},
		{
			name: "one nic with internal and external address",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &compute.PrimaryAddress{
							Address: "192.168.19.86",
							OneToOneNat: &compute.OneToOneNat{
								Address:   "92.68.12.34",
								IpVersion: compute.IpVersion_IPV4,
							},
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":          1,
					"mac_address":    "aa-bb-cc-dd-ee-ff",
					"subnet_id":      "some-subnet-id",
					"ip_address":     "192.168.19.86",
					"nat":            true,
					"nat_ip_address": "92.68.12.34",
					"nat_ip_version": "IPV4",
				},
			},
			externalIP: "92.68.12.34",
			internalIP: "192.168.19.86",
			wantErr:    false,
		},
		{
			name: "one nic with ipv6 address",
			instance: &compute.Instance{
				NetworkInterfaces: []*compute.NetworkInterface{
					{
						Index: "1",
						PrimaryV6Address: &compute.PrimaryAddress{
							Address: "2001:db8::370:7348",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":        1,
					"mac_address":  "aa-bb-cc-dd-ee-ff",
					"subnet_id":    "some-subnet-id",
					"ipv6":         true,
					"ipv6_address": "2001:db8::370:7348",
				},
			},
			externalIP: "2001:db8::370:7348",
			internalIP: "",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nics, externalIP, internalIP, err := flattenInstanceNetworkInterfaces(tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("flattenInstanceNetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(nics, tt.want) {
				t.Errorf("flattenInstanceNetworkInterfaces() nics = %v, want %v", nics, tt.want)
			}
			if externalIP != tt.externalIP {
				t.Errorf("flattenInstanceNetworkInterfaces() externalIP = %v, want %v", externalIP, tt.externalIP)
			}
			if internalIP != tt.internalIP {
				t.Errorf("flattenInstanceNetworkInterfaces() internalIP = %v, want %v", internalIP, tt.internalIP)
			}
		})
	}
}

func TestFlattenInstanceGroupManagedInstanceNetworkInterfaces(t *testing.T) {
	tests := []struct {
		name       string
		instance   *instancegroup.ManagedInstance
		want       []map[string]interface{}
		externalIP string
		internalIP string
		wantErr    bool
	}{
		{
			name: "no nics defined",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{},
			},
			want:       []map[string]interface{}{},
			externalIP: "",
			internalIP: "",
			wantErr:    false,
		},
		{
			name: "one nic with internal address",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &instancegroup.PrimaryAddress{
							Address: "192.168.19.16",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":       1,
					"mac_address": "aa-bb-cc-dd-ee-ff",
					"subnet_id":   "some-subnet-id",
					"ip_address":  "192.168.19.16",
					"nat":         false,
				},
			},
			externalIP: "",
			internalIP: "192.168.19.16",
			wantErr:    false,
		},
		{
			name: "one nic with internal and external address",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{
					{
						Index: "1",
						PrimaryV4Address: &instancegroup.PrimaryAddress{
							Address: "192.168.19.86",
							OneToOneNat: &instancegroup.OneToOneNat{
								Address:   "92.68.12.34",
								IpVersion: instancegroup.IpVersion_IPV4,
							},
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":          1,
					"mac_address":    "aa-bb-cc-dd-ee-ff",
					"subnet_id":      "some-subnet-id",
					"ip_address":     "192.168.19.86",
					"nat":            true,
					"nat_ip_address": "92.68.12.34",
					"nat_ip_version": "IPV4",
				},
			},
			externalIP: "92.68.12.34",
			internalIP: "192.168.19.86",
			wantErr:    false,
		},
		{
			name: "one nic with ipv6 address",
			instance: &instancegroup.ManagedInstance{
				NetworkInterfaces: []*instancegroup.NetworkInterface{
					{
						Index: "1",
						PrimaryV6Address: &instancegroup.PrimaryAddress{
							Address: "2001:db8::370:7348",
						},
						SubnetId:   "some-subnet-id",
						MacAddress: "aa-bb-cc-dd-ee-ff",
					},
				},
			},
			want: []map[string]interface{}{
				{
					"index":        1,
					"mac_address":  "aa-bb-cc-dd-ee-ff",
					"subnet_id":    "some-subnet-id",
					"ipv6":         true,
					"ipv6_address": "2001:db8::370:7348",
				},
			},
			externalIP: "2001:db8::370:7348",
			internalIP: "",
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nics, externalIP, internalIP, err := flattenInstanceGroupManagedInstanceNetworkInterfaces(tt.instance)
			if (err != nil) != tt.wantErr {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(nics, tt.want) {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() nics = %v, want %v", nics, tt.want)
			}
			if externalIP != tt.externalIP {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() externalIP = %v, want %v", externalIP, tt.externalIP)
			}
			if internalIP != tt.internalIP {
				t.Errorf("flattenInstanceGroupManagedInstanceNetworkInterfaces() internalIP = %v, want %v", internalIP, tt.internalIP)
			}
		})
	}
}

func TestFlattenInstanceGroupInstanceTemplateResources(t *testing.T) {
	cases := []struct {
		name      string
		resources *instancegroup.ResourcesSpec
		expected  []map[string]interface{}
	}{
		{
			name: "cores 1 fraction 100 memory 5 gb",
			resources: &instancegroup.ResourcesSpec{
				Cores:        1,
				CoreFraction: 100,
				Memory:       5 * (1 << 30),
			},
			expected: []map[string]interface{}{
				{
					"cores":         1,
					"core_fraction": 100,
					"memory":        5.0,
				},
			},
		},
		{
			name: "cores 8 fraction 5 memory 16 gb",
			resources: &instancegroup.ResourcesSpec{
				Cores:        8,
				CoreFraction: 5,
				Memory:       16 * (1 << 30),
			},
			expected: []map[string]interface{}{
				{
					"cores":         8,
					"core_fraction": 5,
					"memory":        16.0,
				},
			},
		},
		{
			name: "cores 2 fraction 20 memory 0.5 gb",
			resources: &instancegroup.ResourcesSpec{
				Cores:        2,
				CoreFraction: 20,
				Memory:       (1 << 30) / 2,
			},
			expected: []map[string]interface{}{
				{
					"cores":         2,
					"core_fraction": 20,
					"memory":        0.5,
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceGroupInstanceTemplateResources(tc.resources)
			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}

func TestFlattenInstanceGroupAttachedDisk(t *testing.T) {
	cases := []struct {
		name     string
		spec     *instancegroup.AttachedDiskSpec
		expected map[string]interface{}
	}{
		{
			name: "boot disk with diskID",
			spec: &instancegroup.AttachedDiskSpec{
				Mode:       instancegroup.AttachedDiskSpec_READ_WRITE,
				DeviceName: "test-device-name",
				DiskSpec: &instancegroup.AttachedDiskSpec_DiskSpec{
					Description: "mock-disk-description",
					TypeId:      "network-hdd",
					Size:        100 * (1 << 30),
					SourceOneof: &instancegroup.AttachedDiskSpec_DiskSpec_ImageId{
						ImageId: "imageId",
					},
				},
			},
			expected: map[string]interface{}{
				"device_name": "test-device-name",
				"mode":        "READ_WRITE",
				"initialize_params": []map[string]interface{}{
					{
						"description": "mock-disk-description",
						"size":        100,
						"type":        "network-hdd",
						"image_id":    "imageId",
						"snapshot_id": "",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := flattenInstanceGroupAttachedDisk(tc.spec)

			if err != nil {
				t.Fatalf("bad: %#v", err)
			}
			if !reflect.DeepEqual(result, tc.expected) {
				t.Fatalf("Got:\n\n%#v\n\nExpected:\n\n%#v\n", result, tc.expected)
			}
		})
	}
}
