package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccEC2VPNGatewayDataSource_unattached(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dataSourceNameById := "data.aws_vpn_gateway.test_by_id"
	dataSourceNameByTags := "data.aws_vpn_gateway.test_by_tags"
	dataSourceNameByAsn := "data.aws_vpn_gateway.test_by_amazon_side_asn"
	resourceName := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNGatewayUnattachedDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceNameById, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameById, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByAsn, "id", resourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceNameById, "state"),
					resource.TestCheckResourceAttr(dataSourceNameByTags, "tags.%", "3"),
					resource.TestCheckNoResourceAttr(dataSourceNameById, "attached_vpc_id"),
					resource.TestCheckResourceAttr(dataSourceNameByAsn, "amazon_side_asn", "4294967293"),
				),
			},
		},
	})
}

func TestAccEC2VPNGatewayDataSource_attached(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dataSourceName := "data.aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPNGatewayAttachedDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", "aws_vpn_gateway.test", "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "attached_vpc_id", "aws_vpc.test", "id"),
					resource.TestMatchResourceAttr(dataSourceName, "state", regexp.MustCompile("(?i)available")),
				),
			},
		},
	})
}

func testAccVPNGatewayUnattachedDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpn_gateway" "test" {
  tags = {
    Name = "terraform-testacc-vpn-gateway-data-source-unattached-%d"
    ABC  = "testacc-%d"
    XYZ  = "testacc-%d"
  }

  amazon_side_asn = 4294967293
}

data "aws_vpn_gateway" "test_by_id" {
  id = aws_vpn_gateway.test.id
}

data "aws_vpn_gateway" "test_by_tags" {
  tags = aws_vpn_gateway.test.tags
}

data "aws_vpn_gateway" "test_by_amazon_side_asn" {
  amazon_side_asn = aws_vpn_gateway.test.amazon_side_asn
  state           = "available"
}
`, rInt, rInt+1, rInt-1)
}

func testAccVPNGatewayAttachedDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpn-gateway-data-source-attached-%d"
  }
}

resource "aws_vpn_gateway" "test" {
  tags = {
    Name = "terraform-testacc-vpn-gateway-data-source-attached-%d"
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = aws_vpn_gateway.test.id
}

data "aws_vpn_gateway" "test" {
  attached_vpc_id = aws_vpn_gateway_attachment.test.vpc_id
}
`, rInt, rInt)
}
