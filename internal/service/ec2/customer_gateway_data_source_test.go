package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccEC2CustomerGatewayDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_customer_gateway.test"
	resourceName := "aws_customer_gateway.test"

	asn := sdkacctest.RandIntRange(64512, 65534)
	hostOctet := sdkacctest.RandIntRange(1, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayFilterDataSourceConfig(asn, hostOctet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_address", dataSourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
				),
			},
		},
	})
}

func TestAccEC2CustomerGatewayDataSource_id(t *testing.T) {
	dataSourceName := "data.aws_customer_gateway.test"
	resourceName := "aws_customer_gateway.test"

	asn := sdkacctest.RandIntRange(64512, 65534)
	hostOctet := sdkacctest.RandIntRange(1, 254)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomerGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomerGatewayIDDataSourceConfig(asn, hostOctet),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "bgp_asn", dataSourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(resourceName, "ip_address", dataSourceName, "ip_address"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "type", dataSourceName, "type"),
					resource.TestCheckResourceAttrPair(resourceName, "device_name", dataSourceName, "device_name"),
				),
			},
		},
	})
}

func testAccCustomerGatewayFilterDataSourceConfig(asn, hostOctet int) string {
	name := sdkacctest.RandomWithPrefix("test-filter")
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn    = %d
  ip_address = "50.0.0.%d"
  type       = "ipsec.1"

  tags = {
    Name = "%s"
  }
}

data "aws_customer_gateway" "test" {
  filter {
    name   = "tag:Name"
    values = [aws_customer_gateway.test.tags.Name]
  }
}
`, asn, hostOctet, name)
}

func testAccCustomerGatewayIDDataSourceConfig(asn, hostOctet int) string {
	return fmt.Sprintf(`
resource "aws_customer_gateway" "test" {
  bgp_asn     = %d
  ip_address  = "50.0.0.%d"
  device_name = "test"
  type        = "ipsec.1"
}

data "aws_customer_gateway" "test" {
  id = aws_customer_gateway.test.id
}
`, asn, hostOctet)
}
