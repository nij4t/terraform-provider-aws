package ec2_test

import (
	"fmt"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfec2 "github.com/nij4t/terraform-provider-aws/internal/service/ec2"
)

type TunnelOptions struct {
	psk                        string
	tunnelCidr                 string
	dpdTimeoutAction           string
	dpdTimeoutSeconds          int
	ikeVersions                string
	phase1DhGroupNumbers       string
	phase1EncryptionAlgorithms string
	phase1IntegrityAlgorithms  string
	phase1LifetimeSeconds      int
	phase2DhGroupNumbers       string
	phase2EncryptionAlgorithms string
	phase2IntegrityAlgorithms  string
	phase2LifetimeSeconds      int
	rekeyFuzzPercentage        int
	rekeyMarginTimeSeconds     int
	replayWindowSize           int
	startupAction              string
}

func TestAccEC2VPNConnection_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_attachment_id", ""),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpn-connection/vpn-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPNConnectionUpdateConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
				),
			},
		},
	})
}

func TestAccEC2VPNConnection_transitGatewayID(t *testing.T) {
	var vpn ec2.VpnConnection
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resourceName := "aws_vpn_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckTransitGateway(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionTransitGatewayIDConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestMatchResourceAttr(resourceName, "transit_gateway_attachment_id", regexp.MustCompile(`tgw-attach-.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2VPNConnection_tunnel1InsideCIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionTunnel1InsideCIDRConfig(rName, rBgpAsn, "169.254.8.0/30", "169.254.9.0/30"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccEC2VPNConnection_tunnel1InsideIPv6CIDR(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionTunnel1InsideIPv6CIDRConfig(rName, rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_ipv6_cidr", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_ipv6_cidr", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccEC2VPNConnection_tunnel1PreSharedKey(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionTunnel1PresharedKeyConfig(rName, rBgpAsn, "tunnel1presharedkey", "tunnel2presharedkey"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "tunnel1presharedkey"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "tunnel2presharedkey"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

func TestAccEC2VPNConnection_tunnelOptions(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	badCidrRangeErr := regexp.MustCompile(`expected \w+ to not be any of \[[\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}\/30\s?]+\]`)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	tunnel1 := TunnelOptions{
		psk:                        "12345678",
		tunnelCidr:                 "169.254.8.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	tunnel2 := TunnelOptions{
		psk:                        "abcdefgh",
		tunnelCidr:                 "169.254.9.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "2, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "2, 5, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA1\", \"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			// Checking CIDR blocks
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "not-a-cidr"),
				ExpectError: regexp.MustCompile(`invalid CIDR address: not-a-cidr`),
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.254.0/31"),
				ExpectError: regexp.MustCompile(`expected "\w+" to contain a network Value with between 30 and 30 significant bits`),
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "172.16.0.0/30"),
				ExpectError: regexp.MustCompile(`must be within 169.254.0.0/16`),
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.0.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.1.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.2.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.3.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.4.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.5.0/30"),
				ExpectError: badCidrRangeErr,
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "12345678", "169.254.169.252/30"),
				ExpectError: badCidrRangeErr,
			},

			// Checking PreShared Key
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "1234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, sdkacctest.RandStringFromCharSet(65, sdkacctest.CharSetAlpha), "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`expected length of \w+ to be in the range \(8 - 64\)`),
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "01234567", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`cannot start with zero character`),
			},
			{
				Config:      testAccVPNConnectionSingleTunnelOptionsConfig(rName, rBgpAsn, "1234567!", "169.254.254.0/30"),
				ExpectError: regexp.MustCompile(`can only contain alphanumeric, period and underscore characters`),
			},

			// Should pre-check:
			// - local_ipv4_network_cidr
			// - local_ipv6_network_cidr
			// - remote_ipv4_network_cidr
			// - remote_ipv6_network_cidr
			// - tunnel_inside_ip_version
			// - tunnel1_dpd_timeout_action
			// - tunnel1_dpd_timeout_seconds
			// - tunnel1_phase1_lifetime_seconds
			// - tunnel1_phase2_lifetime_seconds
			// - tunnel1_rekey_fuzz_percentage
			// - tunnel1_rekey_margin_time_seconds
			// - tunnel1_replay_window_size
			// - tunnel1_startup_action
			// - tunnel1_inside_cidr
			// - tunnel1_inside_ipv6_cidr

			//Try actual building
			{
				Config: testAccVPNConnectionTunnelOptionsConfig(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1, tunnel2),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),

					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),

					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
				),
			},
			// NOTE: Import does not currently have access to the Terraform configuration,
			//       so proper tunnel ordering is not guaranteed on import. The import
			//       identifier could potentially be updated to accept optional tunnel
			//       configuration information, however the format for this could be
			//       confusing and/or difficult to implement.
		},
	})
}

// TestAccEC2VPNConnection_tunnelOptionsLesser tests less algorithms such as those supported in GovCloud.
func TestAccEC2VPNConnection_tunnelOptionsLesser(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	tunnel1 := TunnelOptions{
		psk:                        "12345678",
		tunnelCidr:                 "169.254.8.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "14, 15, 16, 17, 18, 19, 20, 21",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "14, 15, 16, 17, 18, 19, 20, 21",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	tunnel2 := TunnelOptions{
		psk:                        "abcdefgh",
		tunnelCidr:                 "169.254.9.0/30",
		dpdTimeoutAction:           "clear",
		dpdTimeoutSeconds:          30,
		ikeVersions:                "\"ikev1\", \"ikev2\"",
		phase1DhGroupNumbers:       "14, 15, 16, 17, 18, 19, 20, 21",
		phase1EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase1IntegrityAlgorithms:  "\"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase1LifetimeSeconds:      28800,
		phase2DhGroupNumbers:       "14, 15, 16, 17, 18, 19, 20, 21",
		phase2EncryptionAlgorithms: "\"AES128\", \"AES256\", \"AES128-GCM-16\", \"AES256-GCM-16\"",
		phase2IntegrityAlgorithms:  "\"SHA2-256\", \"SHA2-384\", \"SHA2-512\"",
		phase2LifetimeSeconds:      3600,
		rekeyFuzzPercentage:        100,
		rekeyMarginTimeSeconds:     540,
		replayWindowSize:           1024,
		startupAction:              "add",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionTunnelOptionsConfig(rName, rBgpAsn, "192.168.1.1/32", "192.168.1.2/32", tunnel1, tunnel2),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),

					resource.TestCheckResourceAttr(resourceName, "tunnel1_inside_cidr", "169.254.8.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel1_preshared_key", "12345678"),

					resource.TestCheckResourceAttr(resourceName, "tunnel2_inside_cidr", "169.254.9.0/30"),
					resource.TestCheckResourceAttr(resourceName, "tunnel2_preshared_key", "abcdefgh"),
				),
			},
		},
	})
}

func TestAccEC2VPNConnection_withoutStaticRoutes(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionUpdateConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "static_routes_only", "false"),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2VPNConnection_withEnableAcceleration(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionEnableAccelerationConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "enable_acceleration", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2VPNConnection_withIPv6(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionIPv6Config(rName, rBgpAsn, "fd00:2001:db8:2:2d1:81ff:fe41:d201/128", "fd00:2001:db8:2:2d1:81ff:fe41:d202/128", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2VPNConnection_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionTags1Config(rName, rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVPNConnectionTags2Config(rName, rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVPNConnectionTags1Config(rName, rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2VPNConnection_specifyIPv4(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionLocalRemoteIPv4CIDRsConfig(rName, rBgpAsn, "10.111.0.0/16", "10.222.33.0/24"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "local_ipv4_network_cidr", "10.111.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv4_network_cidr", "10.222.33.0/24"),
				),
			},
		},
	})
}

func TestAccEC2VPNConnection_specifyIPv6(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionIPv6Config(rName, rBgpAsn, "1111:2222:3333:4444::/64", "5555:6666:7777::/48", "fd00:2001:db8:2:2d1:81ff:fe41:d200/126", "fd00:2001:db8:2:2d1:81ff:fe41:d204/126"),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					resource.TestCheckResourceAttr(resourceName, "local_ipv6_network_cidr", "1111:2222:3333:4444::/64"),
					resource.TestCheckResourceAttr(resourceName, "remote_ipv6_network_cidr", "5555:6666:7777::/48"),
				),
			},
		},
	})
}

func TestAccEC2VPNConnection_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_vpn_connection.test"
	var vpn ec2.VpnConnection

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccVPNConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNConnectionConfig(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccVPNConnectionExists(resourceName, &vpn),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceVPNConnection(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVPNConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpn_connection" {
			continue
		}

		resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpnConnectionID.NotFound" {
				// not found
				return nil
			}
			return err
		}

		var vpn *ec2.VpnConnection
		for _, v := range resp.VpnConnections {
			if v.VpnConnectionId != nil && *v.VpnConnectionId == rs.Primary.ID {
				vpn = v
			}
		}

		if vpn == nil {
			// vpn connection not found
			return nil
		}

		if vpn.State != nil && *vpn.State == "deleted" {
			return nil
		}

	}

	return nil
}

func testAccVPNConnectionExists(vpnConnectionResource string, vpnConnection *ec2.VpnConnection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[vpnConnectionResource]
		if !ok {
			return fmt.Errorf("Not found: %s", vpnConnectionResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}
		connection, ok := s.RootModule().Resources[vpnConnectionResource]
		if !ok {
			return fmt.Errorf("Not found: %s", vpnConnectionResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
			VpnConnectionIds: []*string{aws.String(connection.Primary.ID)},
		})

		if err != nil {
			return err
		}

		*vpnConnection = *resp.VpnConnections[0]

		return nil
	}
}

func TestXmlConfigToTunnelInfo(t *testing.T) {
	testCases := []struct {
		Name                  string
		XML                   string
		Tunnel1PreSharedKey   string
		Tunnel1InsideCidr     string
		Tunnel1InsideIpv6Cidr string
		ExpectError           bool
		ExpectTunnelInfo      tfec2.TunnelInfo
	}{
		{
			Name: "outside address sort",
			XML:  testAccVPNTunnelInfoXML,
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "1.1.1.1",
				Tunnel1BGPASN:           "1111",
				Tunnel1BGPHoldTime:      31,
				Tunnel1CgwInsideAddress: "169.254.11.1",
				Tunnel1PreSharedKey:     "FIRST_KEY",
				Tunnel1VgwInsideAddress: "168.254.11.2",
				Tunnel2Address:          "2.2.2.2",
				Tunnel2BGPASN:           "2222",
				Tunnel2BGPHoldTime:      32,
				Tunnel2CgwInsideAddress: "169.254.12.1",
				Tunnel2PreSharedKey:     "SECOND_KEY",
				Tunnel2VgwInsideAddress: "169.254.12.2",
			},
		},
		{
			Name:                "Tunnel1PreSharedKey",
			XML:                 testAccVPNTunnelInfoXML,
			Tunnel1PreSharedKey: "SECOND_KEY",
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
		{
			Name:              "Tunnel1InsideCidr",
			XML:               testAccVPNTunnelInfoXML,
			Tunnel1InsideCidr: "169.254.12.0/30",
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
		// IPv6 logic is equivalent to IPv4, so we can reuse configuration, expected, etc.
		{
			Name:                  "Tunnel1InsideIpv6Cidr",
			XML:                   testAccVPNTunnelInfoXML,
			Tunnel1InsideIpv6Cidr: "169.254.12.1",
			ExpectTunnelInfo: tfec2.TunnelInfo{
				Tunnel1Address:          "2.2.2.2",
				Tunnel1BGPASN:           "2222",
				Tunnel1BGPHoldTime:      32,
				Tunnel1CgwInsideAddress: "169.254.12.1",
				Tunnel1PreSharedKey:     "SECOND_KEY",
				Tunnel1VgwInsideAddress: "169.254.12.2",
				Tunnel2Address:          "1.1.1.1",
				Tunnel2BGPASN:           "1111",
				Tunnel2BGPHoldTime:      31,
				Tunnel2CgwInsideAddress: "169.254.11.1",
				Tunnel2PreSharedKey:     "FIRST_KEY",
				Tunnel2VgwInsideAddress: "168.254.11.2",
			},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			tunnelInfo, err := tfec2.XmlConfigToTunnelInfo(testCase.XML, testCase.Tunnel1PreSharedKey, testCase.Tunnel1InsideCidr, testCase.Tunnel1InsideIpv6Cidr)

			if err == nil && testCase.ExpectError {
				t.Fatalf("expected error, got none")
			}

			if err != nil && !testCase.ExpectError {
				t.Fatalf("expected no error, got: %s", err)
			}

			if actual, expected := *tunnelInfo, testCase.ExpectTunnelInfo; !reflect.DeepEqual(actual, expected) { // nosemgrep: prefer-aws-go-sdk-pointer-conversion-assignment
				t.Errorf("expected tfec2.TunnelInfo:\n%+v\n\ngot:\n%+v\n\n", expected, actual)
			}
		})
	}
}

func testAccVPNConnectionConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true
}
`, rName, rBgpAsn)
}

// Change static_routes_only to be false, forcing a refresh.
func testAccVPNConnectionUpdateConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = "%[1]s-2"
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
}
`, rName, rBgpAsn)
}

func testAccVPNConnectionEnableAccelerationConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = true
}
`, rName, rBgpAsn)
}

func testAccVPNConnectionIPv6Config(rName string, rBgpAsn int, localIpv6NetworkCidr string, remoteIpv6NetworkCidr string, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false
  enable_acceleration = false

  local_ipv6_network_cidr  = %[3]q
  remote_ipv6_network_cidr = %[4]q
  tunnel_inside_ip_version = "ipv6"

  tunnel1_inside_ipv6_cidr = %[5]q
  tunnel2_inside_ipv6_cidr = %[6]q
}
`, rName, rBgpAsn, localIpv6NetworkCidr, remoteIpv6NetworkCidr, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccVPNConnectionSingleTunnelOptionsConfig(rName string, rBgpAsn int, psk string, tunnelCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false

  tunnel1_inside_cidr   = %[3]q
  tunnel1_preshared_key = %[4]q
}
`, rName, rBgpAsn, tunnelCidr, psk)
}

func testAccVPNConnectionTransitGatewayIDConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  transit_gateway_id  = aws_ec2_transit_gateway.test.id
  type                = aws_customer_gateway.test.type
}
`, rName, rBgpAsn)
}

func testAccVPNConnectionTunnel1InsideCIDRConfig(rName string, rBgpAsn int, tunnel1InsideCidr string, tunnel2InsideCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id = aws_customer_gateway.test.id
  tunnel1_inside_cidr = %[3]q
  tunnel2_inside_cidr = %[4]q
  type                = "ipsec.1"
  vpn_gateway_id      = aws_vpn_gateway.test.id
}
`, rName, rBgpAsn, tunnel1InsideCidr, tunnel2InsideCidr)
}

func testAccVPNConnectionTunnel1InsideIPv6CIDRConfig(rName string, rBgpAsn int, tunnel1InsideIpv6Cidr string, tunnel2InsideIpv6Cidr string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  description = %[1]q
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id      = aws_customer_gateway.test.id
  transit_gateway_id       = aws_ec2_transit_gateway.test.id
  tunnel_inside_ip_version = "ipv6"
  tunnel1_inside_ipv6_cidr = %[3]q
  tunnel2_inside_ipv6_cidr = %[4]q
  type                     = "ipsec.1"
}
`, rName, rBgpAsn, tunnel1InsideIpv6Cidr, tunnel2InsideIpv6Cidr)
}

func testAccVPNConnectionTunnel1PresharedKeyConfig(rName string, rBgpAsn int, tunnel1PresharedKey string, tunnel2PresharedKey string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  customer_gateway_id   = aws_customer_gateway.test.id
  tunnel1_preshared_key = %[3]q
  tunnel2_preshared_key = %[4]q
  type                  = "ipsec.1"
  vpn_gateway_id        = aws_vpn_gateway.test.id
}
`, rName, rBgpAsn, tunnel1PresharedKey, tunnel2PresharedKey)
}

func testAccVPNConnectionTunnelOptionsConfig(
	rName string,
	rBgpAsn int,
	localIpv4NetworkCidr string,
	remoteIpv4NetworkCidr string,
	tunnel1 TunnelOptions,
	tunnel2 TunnelOptions,
) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false

  local_ipv4_network_cidr  = %[3]q
  remote_ipv4_network_cidr = %[4]q

  tunnel1_inside_cidr                  = %[5]q
  tunnel1_preshared_key                = %[6]q
  tunnel1_dpd_timeout_action           = %[7]q
  tunnel1_dpd_timeout_seconds          = %[8]d
  tunnel1_ike_versions                 = [%[9]s]
  tunnel1_phase1_dh_group_numbers      = [%[10]s]
  tunnel1_phase1_encryption_algorithms = [%[11]s]
  tunnel1_phase1_integrity_algorithms  = [%[12]s]
  tunnel1_phase1_lifetime_seconds      = %[13]d
  tunnel1_phase2_dh_group_numbers      = [%[14]s]
  tunnel1_phase2_encryption_algorithms = [%[15]s]
  tunnel1_phase2_integrity_algorithms  = [%[16]s]
  tunnel1_phase2_lifetime_seconds      = %[17]d
  tunnel1_rekey_fuzz_percentage        = %[18]d
  tunnel1_rekey_margin_time_seconds    = %[19]d
  tunnel1_replay_window_size           = %[20]d
  tunnel1_startup_action               = %[21]q

  tunnel2_inside_cidr                  = %[22]q
  tunnel2_preshared_key                = %[23]q
  tunnel2_dpd_timeout_action           = %[24]q
  tunnel2_dpd_timeout_seconds          = %[25]d
  tunnel2_ike_versions                 = [%[26]s]
  tunnel2_phase1_dh_group_numbers      = [%[27]s]
  tunnel2_phase1_encryption_algorithms = [%[28]s]
  tunnel2_phase1_integrity_algorithms  = [%[29]s]
  tunnel2_phase1_lifetime_seconds      = %[30]d
  tunnel2_phase2_dh_group_numbers      = [%[31]s]
  tunnel2_phase2_encryption_algorithms = [%[32]s]
  tunnel2_phase2_integrity_algorithms  = [%[33]s]
  tunnel2_phase2_lifetime_seconds      = %[34]d
  tunnel2_rekey_fuzz_percentage        = %[35]d
  tunnel2_rekey_margin_time_seconds    = %[36]d
  tunnel2_replay_window_size           = %[37]d
  tunnel2_startup_action               = %[38]q
}
`,
		rName,
		rBgpAsn,
		localIpv4NetworkCidr,
		remoteIpv4NetworkCidr,
		tunnel1.tunnelCidr,
		tunnel1.psk,
		tunnel1.dpdTimeoutAction,
		tunnel1.dpdTimeoutSeconds,
		tunnel1.ikeVersions,
		tunnel1.phase1DhGroupNumbers,
		tunnel1.phase1EncryptionAlgorithms,
		tunnel1.phase1IntegrityAlgorithms,
		tunnel1.phase1LifetimeSeconds,
		tunnel1.phase2DhGroupNumbers,
		tunnel1.phase2EncryptionAlgorithms,
		tunnel1.phase2IntegrityAlgorithms,
		tunnel1.phase2LifetimeSeconds,
		tunnel1.rekeyFuzzPercentage,
		tunnel1.rekeyMarginTimeSeconds,
		tunnel1.replayWindowSize,
		tunnel1.startupAction,
		tunnel2.tunnelCidr,
		tunnel2.psk,
		tunnel2.dpdTimeoutAction,
		tunnel2.dpdTimeoutSeconds,
		tunnel2.ikeVersions,
		tunnel2.phase1DhGroupNumbers,
		tunnel2.phase1EncryptionAlgorithms,
		tunnel2.phase1IntegrityAlgorithms,
		tunnel2.phase1LifetimeSeconds,
		tunnel2.phase2DhGroupNumbers,
		tunnel2.phase2EncryptionAlgorithms,
		tunnel2.phase2IntegrityAlgorithms,
		tunnel2.phase2LifetimeSeconds,
		tunnel2.rekeyFuzzPercentage,
		tunnel2.rekeyMarginTimeSeconds,
		tunnel2.replayWindowSize,
		tunnel2.startupAction)
}

func testAccVPNConnectionTags1Config(rName string, rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, rBgpAsn, tagKey1, tagValue1)
}

func testAccVPNConnectionTags2Config(rName string, rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = true

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVPNConnectionLocalRemoteIPv4CIDRsConfig(rName string, rBgpAsn int, localIpv4Cidr string, remoteIpv4Cidr string) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_customer_gateway" "test" {
  bgp_asn    = %[2]d
  ip_address = "178.0.0.1"
  type       = "ipsec.1"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_connection" "test" {
  vpn_gateway_id      = aws_vpn_gateway.test.id
  customer_gateway_id = aws_customer_gateway.test.id
  type                = "ipsec.1"
  static_routes_only  = false

  local_ipv4_network_cidr  = %[3]q
  remote_ipv4_network_cidr = %[4]q
}
`, rName, rBgpAsn, localIpv4Cidr, remoteIpv4Cidr)
}

// Test our VPN tunnel config XML parsing
const testAccVPNTunnelInfoXML = `
<vpn_connection id="vpn-abc123">
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>22.22.22.22</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.12.1</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>2.2.2.2</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.12.2</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>2222</asn>
        <hold_time>32</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>SECOND_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
  <ipsec_tunnel>
    <customer_gateway>
      <tunnel_outside_address>
        <ip_address>11.11.11.11</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>169.254.11.1</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
    </customer_gateway>
    <vpn_gateway>
      <tunnel_outside_address>
        <ip_address>1.1.1.1</ip_address>
      </tunnel_outside_address>
      <tunnel_inside_address>
        <ip_address>168.254.11.2</ip_address>
        <network_mask>255.255.255.252</network_mask>
        <network_cidr>30</network_cidr>
      </tunnel_inside_address>
      <bgp>
        <asn>1111</asn>
        <hold_time>31</hold_time>
      </bgp>
    </vpn_gateway>
    <ike>
      <pre_shared_key>FIRST_KEY</pre_shared_key>
    </ike>
  </ipsec_tunnel>
</vpn_connection>
`
