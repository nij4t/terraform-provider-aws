package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccEC2InstanceTypesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_types.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckInstanceTypes(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceName, "instance_types.#", "0"),
				),
			},
		},
	})
}

func TestAccEC2InstanceTypesDataSource_filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_instance_types.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckInstanceTypes(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceTypesDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrGreaterThanValue(dataSourceName, "instance_types.#", "0"),
				),
			},
		},
	})
}

func testAccPreCheckInstanceTypes(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeInstanceTypesInput{}

	_, err := conn.DescribeInstanceTypes(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccInstanceTypesDataSourceConfig() string {
	return `
data "aws_ec2_instance_types" "test" {}
`
}

func testAccInstanceTypesDataSourceConfigFilter() string {
	return `
data "aws_ec2_instance_types" "test" {
  filter {
    name   = "current-generation"
    values = ["true"]
  }
}
`
}
