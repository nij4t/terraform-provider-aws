package xray_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/xray"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfxray "github.com/nij4t/terraform-provider-aws/internal/service/xray"
)

func TestAccXRayGroup_basic(t *testing.T) {
	var Group xray.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, xray.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupBasicConfig(rName, "responsetime > 5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "xray", regexp.MustCompile(`group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "filter_expression", "responsetime > 5"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGroupBasicConfig(rName, "responsetime > 10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "xray", regexp.MustCompile(`group/.+`)),
					resource.TestCheckResourceAttr(resourceName, "group_name", rName),
					resource.TestCheckResourceAttr(resourceName, "filter_expression", "responsetime > 10"),
				),
			},
		},
	})
}

func TestAccXRayGroup_tags(t *testing.T) {
	var Group xray.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, xray.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupBasicTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
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
				Config: testAccGroupBasicTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccGroupBasicTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2")),
			},
		},
	})
}

func TestAccXRayGroup_disappears(t *testing.T) {
	var Group xray.Group
	resourceName := "aws_xray_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, xray.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccGroupBasicConfig(rName, "responsetime > 5"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckXrayGroupExists(resourceName, &Group),
					acctest.CheckResourceDisappears(acctest.Provider, tfxray.ResourceGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckXrayGroupExists(n string, Group *xray.Group) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No XRay Group ID is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayConn

		input := &xray.GetGroupInput{
			GroupARN: aws.String(rs.Primary.ID),
		}

		group, err := conn.GetGroup(input)

		if err != nil {
			return err
		}

		*Group = *group.Group

		return nil
	}
}

func testAccCheckGroupDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_xray_group" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).XRayConn

		input := &xray.GetGroupInput{
			GroupARN: aws.String(rs.Primary.ID),
		}

		group, err := conn.GetGroup(input)

		if tfawserr.ErrMessageContains(err, xray.ErrCodeInvalidRequestException, "Group not found") {
			continue
		}

		if err != nil {
			return err
		}

		if group != nil {
			return fmt.Errorf("Expected XRay Group to be destroyed, %s found", rs.Primary.ID)
		}
	}

	return nil
}

func testAccGroupBasicConfig(rName, expression string) string {
	return fmt.Sprintf(`
resource "aws_xray_group" "test" {
  group_name        = %[1]q
  filter_expression = %[2]q
}
`, rName, expression)
}

func testAccGroupBasicTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_xray_group" "test" {
  group_name        = %[1]q
  filter_expression = "responsetime > 5"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccGroupBasicTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_xray_group" "test" {
  group_name        = %[1]q
  filter_expression = "responsetime > 5"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
