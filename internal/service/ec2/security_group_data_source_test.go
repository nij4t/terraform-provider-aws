package ec2_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccEC2SecurityGroupDataSource_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccSecurityGroupDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_id"),
					resource.TestCheckResourceAttr("data.aws_security_group.by_id", "description", "sg description"),
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_tag"),
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_filter"),
					testAccSecurityGroupCheckDataSource("data.aws_security_group.by_name"),
					testAccSecurityGroupCheckDefaultDataSource("data.aws_security_group.default_by_name"),
				),
			},
		},
	})
}

func testAccSecurityGroupCheckDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		SGRs, ok := s.RootModule().Resources["aws_security_group.test"]
		if !ok {
			return fmt.Errorf("can't find aws_security_group.test in state")
		}
		vpcRs, ok := s.RootModule().Resources["aws_vpc.test"]
		if !ok {
			return fmt.Errorf("can't find aws_vpc.test in state")
		}
		attr := rs.Primary.Attributes

		if attr["id"] != SGRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				SGRs.Primary.Attributes["id"],
			)
		}

		if attr["vpc_id"] != vpcRs.Primary.Attributes["id"] {
			return fmt.Errorf(
				"vpc_id is %s; want %s",
				attr["vpc_id"],
				vpcRs.Primary.Attributes["id"],
			)
		}

		if attr["tags.Name"] != "tf-acctest" {
			return fmt.Errorf("bad Name tag %s", attr["tags.Name"])
		}

		if !strings.Contains(attr["arn"], attr["id"]) {
			return fmt.Errorf("bad ARN %s", attr["arn"])
		}

		return nil
	}
}

func testAccSecurityGroupCheckDefaultDataSource(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		vpcRs, ok := s.RootModule().Resources["aws_vpc.test"]
		if !ok {
			return fmt.Errorf("can't find aws_vpc.test in state")
		}
		attr := rs.Primary.Attributes

		if attr["id"] != vpcRs.Primary.Attributes["default_security_group_id"] {
			return fmt.Errorf(
				"id is %s; want %s",
				attr["id"],
				vpcRs.Primary.Attributes["default_security_group_id"],
			)
		}

		return nil
	}
}

func testAccSecurityGroupDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = "terraform-testacc-security-group-data-source"
  }
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "test-%d"

  tags = {
    Name = "tf-acctest"
    Seed = "%d"
  }

  description = "sg description"
}

data "aws_security_group" "by_id" {
  id = aws_security_group.test.id
}

data "aws_security_group" "by_name" {
  name = aws_security_group.test.name
}

data "aws_security_group" "default_by_name" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_security_group" "by_tag" {
  tags = {
    Seed = aws_security_group.test.tags["Seed"]
  }
}

data "aws_security_group" "by_filter" {
  filter {
    name   = "group-name"
    values = [aws_security_group.test.name]
  }
}
`, rInt, rInt)
}
