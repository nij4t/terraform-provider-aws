package ec2_test

import (
	"fmt"
	"log"
	"reflect"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccEC2VPCPeeringConnection_basic(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
		},
	})
}

func TestAccEC2VPCPeeringConnection_plan(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))
	resourceName := "aws_vpc_peering_connection.test"

	// reach out and DELETE the VPC Peering connection outside of Terraform
	testDestroy := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		log.Printf("[DEBUG] Test deleting the VPC Peering Connection.")
		_, err := conn.DeleteVpcPeeringConnection(
			&ec2.DeleteVpcPeeringConnectionInput{
				VpcPeeringConnectionId: connection.VpcPeeringConnectionId,
			})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2VPCPeeringConnection_tags(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringTagsConfig1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
			{
				Config: testAccVpcPeeringTagsConfig2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccVpcPeeringTagsConfig1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(resourceName, &connection),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccEC2VPCPeeringConnection_options(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))
	resourceName := "aws_vpc_peering_connection.test"

	testAccepterChange := func(*terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		log.Printf("[DEBUG] Test change to the VPC Peering Connection Options.")

		_, err := conn.ModifyVpcPeeringConnectionOptions(
			&ec2.ModifyVpcPeeringConnectionOptionsInput{
				VpcPeeringConnectionId: connection.VpcPeeringConnectionId,
				AccepterPeeringConnectionOptions: &ec2.PeeringConnectionOptionsRequest{
					AllowDnsResolutionFromRemoteVpc: aws.Bool(false),
				},
			})

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig_options(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_remote_vpc_dns_resolution",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_classic_link_to_remote_vpc",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_vpc_to_remote_classic_link",
						"true",
					),
					testAccCheckVPCPeeringConnectionOptions(
						resourceName, "requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(true),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(true),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_vpc_to_remote_classic_link",
						"false",
					),
					testAccCheckVPCPeeringConnectionOptions(
						resourceName, "accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
					testAccepterChange,
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
			{
				Config: testAccVpcPeeringConfig_options(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					// Requester's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_remote_vpc_dns_resolution",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_classic_link_to_remote_vpc",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"requester.0.allow_vpc_to_remote_classic_link",
						"true",
					),
					testAccCheckVPCPeeringConnectionOptions(
						resourceName, "requester",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(false),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(true),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(true),
						},
					),
					// Accepter's view:
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.#",
						"1",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_remote_vpc_dns_resolution",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_classic_link_to_remote_vpc",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accepter.0.allow_vpc_to_remote_classic_link",
						"false",
					),
					testAccCheckVPCPeeringConnectionOptions(
						resourceName, "accepter",
						&ec2.VpcPeeringConnectionOptionsDescription{
							AllowDnsResolutionFromRemoteVpc:            aws.Bool(true),
							AllowEgressFromLocalClassicLinkToRemoteVpc: aws.Bool(false),
							AllowEgressFromLocalVpcToRemoteClassicLink: aws.Bool(false),
						},
					),
				),
			},
		},
	})
}

func TestAccEC2VPCPeeringConnection_failedState(t *testing.T) {
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVpcPeeringConfig_failedState(rName),
				ExpectError: regexp.MustCompile(`.*Error waiting.*\(pcx-\w+\).*incorrect.*VPC-ID.*`),
			},
		},
	})
}

func testAccCheckVPCPeeringConnectionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_peering_connection" {
			continue
		}

		describe, err := conn.DescribeVpcPeeringConnections(
			&ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []*string{aws.String(rs.Primary.ID)},
			})

		if err != nil {
			return err
		}

		var pc *ec2.VpcPeeringConnection
		for _, c := range describe.VpcPeeringConnections {
			if rs.Primary.ID == *c.VpcPeeringConnectionId {
				pc = c
			}
		}

		if pc == nil {
			// not found
			return nil
		}

		if pc.Status != nil {
			if *pc.Status.Code == "deleted" || *pc.Status.Code == "rejected" || *pc.Status.Code == "failed" {
				return nil
			}
			return fmt.Errorf("Found the VPC Peering Connection in an unexpected state: %s", pc)
		}

		// return error here; we've found the vpc_peering object we want, however
		// it's not in an expected state
		return fmt.Errorf("Fall through error for testAccCheckVPCPeeringConnectionDestroy.")
	}

	return nil
}

func testAccCheckVPCPeeringConnectionExists(n string, connection *ec2.VpcPeeringConnection) resource.TestCheckFunc {
	return testAccCheckVPCPeeringConnectionExistsWithProvider(n, connection, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckVPCPeeringConnectionExistsWithProvider(n string, connection *ec2.VpcPeeringConnection, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Peering Connection ID is set.")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeVpcPeeringConnections(
			&ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []*string{aws.String(rs.Primary.ID)},
			})
		if err != nil {
			return err
		}
		if len(resp.VpcPeeringConnections) == 0 {
			return fmt.Errorf("VPC Peering Connection could not be found")
		}

		*connection = *resp.VpcPeeringConnections[0]

		return nil
	}
}

func testAccCheckVPCPeeringConnectionOptions(n, block string, options *ec2.VpcPeeringConnectionOptionsDescription) resource.TestCheckFunc {
	return testAccCheckVPCPeeringConnectionOptionsWithProvider(n, block, options, func() *schema.Provider { return acctest.Provider })
}

func testAccCheckVPCPeeringConnectionOptionsWithProvider(n, block string, options *ec2.VpcPeeringConnectionOptionsDescription, providerF func() *schema.Provider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Peering Connection ID is set.")
		}

		conn := providerF().Meta().(*conns.AWSClient).EC2Conn
		resp, err := conn.DescribeVpcPeeringConnections(
			&ec2.DescribeVpcPeeringConnectionsInput{
				VpcPeeringConnectionIds: []*string{aws.String(rs.Primary.ID)},
			})
		if err != nil {
			return err
		}

		pc := resp.VpcPeeringConnections[0]

		o := pc.AccepterVpcInfo
		if block == "requester" {
			o = pc.RequesterVpcInfo
		}

		if !reflect.DeepEqual(o.PeeringOptions, options) {
			return fmt.Errorf("Expected the VPC Peering Connection Options to be %#v, got %#v",
				options, o.PeeringOptions)
		}

		return nil
	}
}

func TestAccEC2VPCPeeringConnection_peerRegionAutoAccept(t *testing.T) {
	var providers []*schema.Provider
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVpcPeeringConfig_region_autoAccept(rName, true),
				ExpectError: regexp.MustCompile(`.*peer_region cannot be set whilst auto_accept is true when creating a vpc peering connection.*`),
			},
		},
	})
}

func TestAccEC2VPCPeeringConnection_region(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	var providers []*schema.Provider
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.FactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig_region_autoAccept(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"pending-acceptance",
					),
				),
			},
		},
	})
}

// Tests the peering connection acceptance functionality for same region, same account.
func TestAccEC2VPCPeeringConnection_accept(t *testing.T) {
	var connection ec2.VpcPeeringConnection
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcPeeringConfig_autoAccept(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"pending-acceptance",
					),
				),
			},
			{
				Config: testAccVpcPeeringConfig_autoAccept(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"active",
					),
				),
			},
			// Tests that changing 'auto_accept' back to false keeps the connection active.
			{
				Config: testAccVpcPeeringConfig_autoAccept(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVPCPeeringConnectionExists(
						resourceName,
						&connection,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"accept_status",
						"active",
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"auto_accept",
				},
			},
		},
	})
}

// Tests that VPC peering connection options can't be set on non-active connection.
func TestAccEC2VPCPeeringConnection_optionsNoAutoAccept(t *testing.T) {
	rName := fmt.Sprintf("tf-testacc-pcx-%s", sdkacctest.RandString(17))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVPCPeeringConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccVpcPeeringConfig_options_noAutoAccept(rName),
				ExpectError: regexp.MustCompile(`.*Unable to modify peering options\. The VPC Peering Connection "pcx-\w+" is not active\..*`),
			},
		},
	})
}

func testAccVpcPeeringConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true
}
`, rName)
}

func testAccVpcPeeringTagsConfig1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccVpcPeeringTagsConfig2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccVpcPeeringConfig_options(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`, rName)
}

func testAccVpcPeeringConfig_failedState(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccVpcPeeringConfig_region_autoAccept(rName string, autoAccept bool) string {
	return acctest.ConfigAlternateRegionProvider() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  peer_region = %[3]q
  auto_accept = %[2]t

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept, acctest.AlternateRegion())
}

func testAccVpcPeeringConfig_autoAccept(rName string, autoAccept bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = %t

  tags = {
    Name = %[1]q
  }
}
`, rName, autoAccept)
}

func testAccVpcPeeringConfig_options_noAutoAccept(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "peer" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.peer.id
  auto_accept = false

  tags = {
    Name = %[1]q
  }

  accepter {
    allow_remote_vpc_dns_resolution = true
  }

  requester {
    allow_vpc_to_remote_classic_link = true
    allow_classic_link_to_remote_vpc = true
  }
}
`, rName)
}
