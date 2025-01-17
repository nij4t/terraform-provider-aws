package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`.

func TestAccOpsWorksNodejsAppLayer_basic(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_nodejs_app_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNodejsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNodejsAppLayerVPCCreateConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", stackName)),
			},
		},
	})
}

func TestAccOpsWorksNodejsAppLayer_tags(t *testing.T) {
	var opslayer opsworks.Layer
	stackName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_nodejs_app_layer.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNodejsAppLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNodejsAppLayerTags1Config(stackName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccNodejsAppLayerTags2Config(stackName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccNodejsAppLayerTags1Config(stackName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckNodejsAppLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_nodejs_app_layer", s)
}

func testAccNodejsAppLayerVPCCreateConfig(name string) string {
	return testAccStackVPCCreateConfig(name) +
		testAccCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_nodejs_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
}
`, name)
}

func testAccNodejsAppLayerTags1Config(name, tagKey1, tagValue1 string) string {
	return testAccStackVPCCreateConfig(name) +
		testAccCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_nodejs_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccNodejsAppLayerTags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccStackVPCCreateConfig(name) +
		testAccCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_nodejs_app_layer" "test" {
  stack_id = aws_opsworks_stack.tf-acc.id
  name     = "%s"

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
