// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package manifest

import (
	"testing"

	"github.com/aws/copilot-cli/internal/pkg/config"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/copilot-cli/internal/pkg/template"
	"github.com/stretchr/testify/require"
)

func TestFromEnvConfig(t *testing.T) {
	testCases := map[string]struct {
		in     *config.Environment
		wanted *Environment
	}{
		"converts configured VPC settings with availability zones after v1.14": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				CustomConfig: &config.CustomizeEnv{
					VPCConfig: &config.AdjustVPC{
						CIDR:               "10.0.0.0/16",
						AZs:                []string{"us-west-2a", "us-west-2b", "us-west-2c"},
						PublicSubnetCIDRs:  []string{"10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"},
						PrivateSubnetCIDRs: []string{"10.0.3.0/24", "10.0.4.0/24", "10.0.5.0/24"},
					},
				},
			},

			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Network: environmentNetworkConfig{
						VPC: environmentVPCConfig{
							CIDR: ipNetP("10.0.0.0/16"),
							Subnets: subnetsConfiguration{
								Public: []subnetConfiguration{
									{
										CIDR: ipNetP("10.0.0.0/24"),
										AZ:   stringP("us-west-2a"),
									},
									{
										CIDR: ipNetP("10.0.1.0/24"),
										AZ:   stringP("us-west-2b"),
									},
									{
										CIDR: ipNetP("10.0.2.0/24"),
										AZ:   stringP("us-west-2c"),
									},
								},
								Private: []subnetConfiguration{
									{
										CIDR: ipNetP("10.0.3.0/24"),
										AZ:   stringP("us-west-2a"),
									},
									{
										CIDR: ipNetP("10.0.4.0/24"),
										AZ:   stringP("us-west-2b"),
									},
									{
										CIDR: ipNetP("10.0.5.0/24"),
										AZ:   stringP("us-west-2c"),
									},
								},
							},
						},
					},
				},
			},
		},
		"converts configured VPC settings without any availability zones set": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				CustomConfig: &config.CustomizeEnv{
					VPCConfig: &config.AdjustVPC{
						CIDR:               "10.0.0.0/16",
						PublicSubnetCIDRs:  []string{"10.0.0.0/24", "10.0.1.0/24", "10.0.2.0/24"},
						PrivateSubnetCIDRs: []string{"10.0.3.0/24", "10.0.4.0/24", "10.0.5.0/24"},
					},
				},
			},

			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Network: environmentNetworkConfig{
						VPC: environmentVPCConfig{
							CIDR: ipNetP("10.0.0.0/16"),
							Subnets: subnetsConfiguration{
								Public: []subnetConfiguration{
									{
										CIDR: ipNetP("10.0.0.0/24"),
									},
									{
										CIDR: ipNetP("10.0.1.0/24"),
									},
									{
										CIDR: ipNetP("10.0.2.0/24"),
									},
								},
								Private: []subnetConfiguration{
									{
										CIDR: ipNetP("10.0.3.0/24"),
									},
									{
										CIDR: ipNetP("10.0.4.0/24"),
									},
									{
										CIDR: ipNetP("10.0.5.0/24"),
									},
								},
							},
						},
					},
				},
			},
		},
		"converts imported VPC settings": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				CustomConfig: &config.CustomizeEnv{
					ImportVPC: &config.ImportVPC{
						ID:               "vpc-3f139646",
						PublicSubnetIDs:  []string{"pub1", "pub2", "pub3"},
						PrivateSubnetIDs: []string{"priv1", "priv2", "priv3"},
					},
				},
			},
			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Network: environmentNetworkConfig{
						VPC: environmentVPCConfig{
							ID: stringP("vpc-3f139646"),
							Subnets: subnetsConfiguration{
								Public: []subnetConfiguration{
									{
										SubnetID: stringP("pub1"),
									},
									{
										SubnetID: stringP("pub2"),
									},
									{
										SubnetID: stringP("pub3"),
									},
								},
								Private: []subnetConfiguration{
									{
										SubnetID: stringP("priv1"),
									},
									{
										SubnetID: stringP("priv2"),
									},
									{
										SubnetID: stringP("priv3"),
									},
								},
							},
						},
					},
				},
			},
		},
		"converts imported certificates for a public load balancer": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				CustomConfig: &config.CustomizeEnv{
					ImportCertARNs: []string{"arn:aws:acm:region:account:certificate/certificate_ID_1", "arn:aws:acm:region:account:certificate/certificate_ID_2"},
					ImportVPC: &config.ImportVPC{
						PublicSubnetIDs: []string{"subnet1"},
					},
				},
			},

			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Network: environmentNetworkConfig{
						VPC: environmentVPCConfig{
							Subnets: subnetsConfiguration{
								Public: []subnetConfiguration{
									{
										SubnetID: aws.String("subnet1"),
										CIDR:     nil,
										AZ:       nil,
									},
								},
								Private: []subnetConfiguration{},
							},
						},
					},
					HTTPConfig: environmentHTTPConfig{
						Public: publicHTTPConfig{
							Certificates: []string{"arn:aws:acm:region:account:certificate/certificate_ID_1", "arn:aws:acm:region:account:certificate/certificate_ID_2"},
						},
					},
				},
			},
		},
		"converts imported certificates for a public load balancer without an imported vpc": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				CustomConfig: &config.CustomizeEnv{
					ImportCertARNs: []string{"arn:aws:acm:region:account:certificate/certificate_ID_1", "arn:aws:acm:region:account:certificate/certificate_ID_2"},
				},
			},

			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					HTTPConfig: environmentHTTPConfig{
						Public: publicHTTPConfig{
							Certificates: []string{"arn:aws:acm:region:account:certificate/certificate_ID_1", "arn:aws:acm:region:account:certificate/certificate_ID_2"},
						},
					},
				},
			},
		},
		"converts imported certificates for a private load balancer with subnet placement specified": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				CustomConfig: &config.CustomizeEnv{
					ImportCertARNs: []string{"arn:aws:acm:region:account:certificate/certificate_ID_1", "arn:aws:acm:region:account:certificate/certificate_ID_2"},
					ImportVPC: &config.ImportVPC{
						PrivateSubnetIDs: []string{"subnet1", "subnet2"},
					},
					InternalALBSubnets: []string{"subnet2"},
				},
			},

			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Network: environmentNetworkConfig{
						VPC: environmentVPCConfig{
							Subnets: subnetsConfiguration{
								Private: []subnetConfiguration{
									{
										SubnetID: aws.String("subnet1"),
									},
									{
										SubnetID: aws.String("subnet2"),
									},
								},
								Public: []subnetConfiguration{},
							},
						},
					},
					HTTPConfig: environmentHTTPConfig{
						Private: privateHTTPConfig{
							InternalALBSubnets: []string{"subnet2"},
							Certificates:       []string{"arn:aws:acm:region:account:certificate/certificate_ID_1", "arn:aws:acm:region:account:certificate/certificate_ID_2"},
						},
					},
				},
			},
		},
		"converts container insights": {
			in: &config.Environment{
				App:  "phonetool",
				Name: "test",
				Telemetry: &config.Telemetry{
					EnableContainerInsights: false,
				},
			},

			wanted: &Environment{
				Workload: Workload{
					Name: stringP("test"),
					Type: stringP("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Observability: environmentObservability{
						ContainerInsights: aws.Bool(false),
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, tc.wanted, FromEnvConfig(tc.in, nil))
		})
	}
}

func Test_UnmarshalEnvironment(t *testing.T) {
	var (
		mockVPCCIDR            = IPNet("10.0.0.0/16")
		mockPublicSubnet1CIDR  = IPNet("10.0.0.0/24")
		mockPublicSubnet2CIDR  = IPNet("10.0.1.0/24")
		mockPrivateSubnet1CIDR = IPNet("10.0.3.0/24")
		mockPrivateSubnet2CIDR = IPNet("10.0.4.0/24")
	)
	testCases := map[string]struct {
		inContent       string
		wantedStruct    *Environment
		wantedErrPrefix string
	}{
		"unmarshal with managed VPC": {
			inContent: `name: test
type: Environment

network:
    vpc:
        cidr: '10.0.0.0/16'       
        subnets:
            public:
                - cidr: '10.0.0.0/24'  
                  az: 'us-east-2a'      
                - cidr: '10.0.1.0/24'   
                  az: 'us-east-2b'      
            private:
                - cidr: '10.0.3.0/24'   
                  az: 'us-east-2a'      
                - cidr: '10.0.4.0/24'   
                  az: 'us-east-2b'  
`,
			wantedStruct: &Environment{
				Workload: Workload{
					Name: aws.String("test"),
					Type: aws.String("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Network: environmentNetworkConfig{
						VPC: environmentVPCConfig{
							CIDR: &mockVPCCIDR,
							Subnets: subnetsConfiguration{
								Public: []subnetConfiguration{
									{
										CIDR: &mockPublicSubnet1CIDR,
										AZ:   aws.String("us-east-2a"),
									},
									{
										CIDR: &mockPublicSubnet2CIDR,
										AZ:   aws.String("us-east-2b"),
									},
								},
								Private: []subnetConfiguration{
									{
										CIDR: &mockPrivateSubnet1CIDR,
										AZ:   aws.String("us-east-2a"),
									},
									{
										CIDR: &mockPrivateSubnet2CIDR,
										AZ:   aws.String("us-east-2b"),
									},
								},
							},
						},
					},
				},
			},
		},
		"unmarshal with observability": {
			inContent: `name: prod
type: Environment

observability:
    container_insights: true
`,
			wantedStruct: &Environment{
				Workload: Workload{
					Name: aws.String("prod"),
					Type: aws.String("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					Observability: environmentObservability{
						ContainerInsights: aws.Bool(true),
					},
				},
			},
		},
		"unmarshal with http": {
			inContent: `name: prod
type: Environment

http:
    public:
        certificates:
            - cert-1
            - cert-2
`,
			wantedStruct: &Environment{
				Workload: Workload{
					Name: aws.String("prod"),
					Type: aws.String("Environment"),
				},
				EnvironmentConfig: EnvironmentConfig{
					HTTPConfig: environmentHTTPConfig{
						Public: publicHTTPConfig{
							Certificates: []string{"cert-1", "cert-2"},
						},
					},
				},
			},
		},
		"fail to unmarshal": {
			inContent:       `watermelon in easter hay`,
			wantedErrPrefix: "unmarshal environment manifest: ",
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got, gotErr := UnmarshalEnvironment([]byte(tc.inContent))
			if tc.wantedErrPrefix != "" {
				require.ErrorContains(t, gotErr, tc.wantedErrPrefix)
			} else {
				require.NoError(t, gotErr)
				require.Equal(t, tc.wantedStruct, got)
			}
		})
	}
}

func TestEnvironmentVPCConfig_ImportedVPC(t *testing.T) {
	testCases := map[string]struct {
		inVPCConfig environmentVPCConfig
		wanted      *template.ImportVPC
	}{
		"vpc not imported": {},
		"only public subnets imported": {
			inVPCConfig: environmentVPCConfig{
				ID: aws.String("vpc-1234"),
				Subnets: subnetsConfiguration{
					Public: []subnetConfiguration{
						{
							SubnetID: aws.String("subnet-123"),
						},
						{
							SubnetID: aws.String("subnet-456"),
						},
					},
				},
			},
			wanted: &template.ImportVPC{
				ID:              "vpc-1234",
				PublicSubnetIDs: []string{"subnet-123", "subnet-456"},
			},
		},
		"only private subnets imported": {
			inVPCConfig: environmentVPCConfig{
				ID: aws.String("vpc-1234"),
				Subnets: subnetsConfiguration{
					Private: []subnetConfiguration{
						{
							SubnetID: aws.String("subnet-123"),
						},
						{
							SubnetID: aws.String("subnet-456"),
						},
					},
				},
			},
			wanted: &template.ImportVPC{
				ID:               "vpc-1234",
				PrivateSubnetIDs: []string{"subnet-123", "subnet-456"},
			},
		},
		"both subnets imported": {
			inVPCConfig: environmentVPCConfig{
				ID: aws.String("vpc-1234"),
				Subnets: subnetsConfiguration{
					Public: []subnetConfiguration{
						{
							SubnetID: aws.String("subnet-123"),
						},
						{
							SubnetID: aws.String("subnet-456"),
						},
					},
					Private: []subnetConfiguration{
						{
							SubnetID: aws.String("subnet-789"),
						},
						{
							SubnetID: aws.String("subnet-012"),
						},
					},
				},
			},
			wanted: &template.ImportVPC{
				ID:               "vpc-1234",
				PublicSubnetIDs:  []string{"subnet-123", "subnet-456"},
				PrivateSubnetIDs: []string{"subnet-789", "subnet-012"},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.inVPCConfig.ImportedVPC()
			require.Equal(t, tc.wanted, got)
		})
	}
}

func TestEnvironmentVPCConfig_ManagedVPC(t *testing.T) {
	var (
		mockVPCCIDR            = IPNet("10.0.0.0/16")
		mockPublicSubnet1CIDR  = IPNet("10.0.0.0/24")
		mockPublicSubnet2CIDR  = IPNet("10.0.1.0/24")
		mockPublicSubnet3CIDR  = IPNet("10.0.2.0/24")
		mockPrivateSubnet1CIDR = IPNet("10.0.3.0/24")
		mockPrivateSubnet2CIDR = IPNet("10.0.4.0/24")
		mockPrivateSubnet3CIDR = IPNet("10.0.5.0/24")
	)
	testCases := map[string]struct {
		inVPCConfig environmentVPCConfig
		wanted      *template.ManagedVPC
	}{
		"default vpc without custom configuration": {},
		"with imported vpc": {
			inVPCConfig: environmentVPCConfig{
				ID: aws.String("vpc-1234"),
			},
		},
		"ensure custom configuration is sorted by AZ": {
			inVPCConfig: environmentVPCConfig{
				CIDR: &mockVPCCIDR,
				Subnets: subnetsConfiguration{
					Public: []subnetConfiguration{
						{
							CIDR: &mockPublicSubnet1CIDR,
							AZ:   aws.String("us-east-2a"),
						},
						{
							CIDR: &mockPublicSubnet3CIDR,
							AZ:   aws.String("us-east-2c"),
						},
						{
							CIDR: &mockPublicSubnet2CIDR,
							AZ:   aws.String("us-east-2b"),
						},
					},
					Private: []subnetConfiguration{
						{
							CIDR: &mockPrivateSubnet2CIDR,
							AZ:   aws.String("us-east-2b"),
						},
						{
							CIDR: &mockPrivateSubnet1CIDR,
							AZ:   aws.String("us-east-2a"),
						},
						{
							CIDR: &mockPrivateSubnet3CIDR,
							AZ:   aws.String("us-east-2c"),
						},
					},
				},
			},
			wanted: &template.ManagedVPC{
				CIDR:               string(mockVPCCIDR),
				AZs:                []string{"us-east-2a", "us-east-2b", "us-east-2c"},
				PublicSubnetCIDRs:  []string{string(mockPublicSubnet1CIDR), string(mockPublicSubnet2CIDR), string(mockPublicSubnet3CIDR)},
				PrivateSubnetCIDRs: []string{string(mockPrivateSubnet1CIDR), string(mockPrivateSubnet2CIDR), string(mockPrivateSubnet3CIDR)},
			},
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.inVPCConfig.ManagedVPC()
			require.Equal(t, tc.wanted, got)
		})
	}
}

func TestEnvironmentObservability_IsEmpty(t *testing.T) {
	testCases := map[string]struct {
		in     environmentObservability
		wanted bool
	}{
		"empty": {
			in:     environmentObservability{},
			wanted: true,
		},
		"not empty": {
			in: environmentObservability{
				ContainerInsights: aws.Bool(false),
			},
			wanted: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			got := tc.in.IsEmpty()
			require.Equal(t, tc.wanted, got)
		})
	}
}