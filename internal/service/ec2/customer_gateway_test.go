package ec2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfec2 "github.com/nij4t/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2CustomerGateway_basic(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`customer-gateway/cgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(rBgpAsn)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomerGatewayConfigForceReplace(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
				),
			},
		},
	})
}

func TestAccEC2CustomerGateway_tags(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfigTags1(rBgpAsn, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1")),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCustomerGatewayConfigTags2(rBgpAsn, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
			{
				Config: testAccCustomerGatewayConfigTags1(rBgpAsn, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccEC2CustomerGateway_similarAlreadyExists(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccCustomerGatewayConfigIdentical(rBgpAsn),
				ExpectError: regexp.MustCompile("An existing customer gateway"),
			},
		},
	})
}

func TestAccEC2CustomerGateway_deviceName(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfigDeviceName(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`customer-gateway/cgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", strconv.Itoa(rBgpAsn)),
					resource.TestCheckResourceAttr(resourceName, "device_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccEC2CustomerGateway_disappears(t *testing.T) {
	rBgpAsn := sdkacctest.RandIntRange(64512, 65534)
	var gateway ec2.CustomerGateway
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceCustomerGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2CustomerGateway_4ByteASN(t *testing.T) {
	var gateway ec2.CustomerGateway
	rBgpAsn := strconv.FormatInt(int64(sdkacctest.RandIntRange(64512, 65534))*10000, 10)
	resourceName := "aws_customer_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayConfig4ByteAsn(rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCustomerGateway(resourceName, &gateway),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`customer-gateway/cgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "bgp_asn", rBgpAsn),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func testAccCheckCustomerGatewayDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_customer_gatewah" {
			continue
		}

		gatewayFilter := &ec2.Filter{
			Name:   aws.String("customer-gateway-id"),
			Values: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
			Filters: []*ec2.Filter{gatewayFilter},
		})

		if tfawserr.ErrMessageContains(err, "InvalidCustomerGatewayID.NotFound", "") {
			continue
		}

		if err == nil {
			if len(resp.CustomerGateways) > 0 {
				return fmt.Errorf("Customer gateway still exists: %v", resp.CustomerGateways)
			}

			if aws.StringValue(resp.CustomerGateways[0].State) == "deleted" {
				continue
			}
		}

		return err
	}

	return nil
}

func testAccCheckCustomerGateway(gatewayResource string, cgw *ec2.CustomerGateway) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[gatewayResource]
		if !ok {
			return fmt.Errorf("Not found: %s", gatewayResource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		gateway, ok := s.RootModule().Resources[gatewayResource]
		if !ok {
			return fmt.Errorf("Not found: %s", gatewayResource)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		gatewayFilter := &ec2.Filter{
			Name:   aws.String("customer-gateway-id"),
			Values: []*string{aws.String(gateway.Primary.ID)},
		}

		resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
			Filters: []*ec2.Filter{gatewayFilter},
		})

		if err != nil {
			return err
		}

		respGateway := resp.CustomerGateways[0]
		*cgw = *respGateway

		return nil
	}
}

func testAccCustomerGatewayConfig(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}
`, rBgpAsn)
}

func testAccCustomerGatewayConfigTags1(rBgpAsn int, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"

  tags = {
    %[2]q = %[3]q
  }
}
`, rBgpAsn, tagKey1, tagValue1)
}

func testAccCustomerGatewayConfigTags2(rBgpAsn int, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rBgpAsn, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccCustomerGatewayConfigIdentical(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}

resource "aws_customer_gateway" "identical" {
  bgp_asn    = %[1]d
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}
`, rBgpAsn)
}

func testAccCustomerGatewayConfigDeviceName(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn     = %[1]d
  ip_address  = "172.0.0.1"
  type        = "ipsec.1"
  device_name = "test"
}
`, rBgpAsn)
}

// Change the ip_address.
func testAccCustomerGatewayConfigForceReplace(rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "172.10.10.1"
  type       = "ipsec.1"
}
`, rBgpAsn)
}

func testAccCustomerGatewayConfig4ByteAsn(rBgpAsn string) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %[1]q
  ip_address = "172.0.0.1"
  type       = "ipsec.1"
}
`, rBgpAsn)
}
