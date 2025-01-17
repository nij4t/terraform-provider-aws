package opsworks_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/opsworks"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

// These tests assume the existence of predefined Opsworks IAM roles named `aws-opsworks-ec2-role`
// and `aws-opsworks-service-role`, and Opsworks stacks named `tf-acc`.

func TestAccOpsWorksCustomLayer_basic(t *testing.T) {
	name := sdkacctest.RandString(10)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.tf-acc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerVPCCreateConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "gp2",
						"number_of_disks": "2",
						"mount_point":     "/home",
						"size":            "100",
						"encrypted":       "false",
					}),
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

func TestAccOpsWorksCustomLayer_tags(t *testing.T) {
	name := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerTags1Config(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
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
				Config: testAccCustomLayerTags2Config(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccCustomLayerTags1Config(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccOpsWorksCustomLayer_noVPC(t *testing.T) {
	stackName := fmt.Sprintf("tf-%d", sdkacctest.RandInt())
	var opslayer opsworks.Layer
	resourceName := "aws_opsworks_custom_layer.tf-acc"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCustomLayerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomLayerNoVPCCreateConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLayerExists(resourceName, &opslayer),
					testAccCheckCreateLayerAttributes(&opslayer, stackName),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "auto_assign_elastic_ips", "false"),
					resource.TestCheckResourceAttr(resourceName, "auto_healing", "true"),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "true"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "300"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "gp2",
						"number_of_disks": "2",
						"mount_point":     "/home",
						"size":            "100",
						"encrypted":       "false",
					}),
				),
			},
			{
				Config: testAccCustomLayerUpdateConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "drain_elb_on_shutdown", "false"),
					resource.TestCheckResourceAttr(resourceName, "instance_shutdown_timeout", "120"),
					resource.TestCheckResourceAttr(resourceName, "custom_security_group_ids.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "system_packages.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "git"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "golang"),
					resource.TestCheckTypeSetElemAttr(resourceName, "system_packages.*", "subversion"),
					resource.TestCheckResourceAttr(resourceName, "ebs_volume.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "gp2",
						"number_of_disks": "2",
						"mount_point":     "/home",
						"size":            "100",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "ebs_volume.*", map[string]string{
						"type":            "io1",
						"number_of_disks": "4",
						"mount_point":     "/var",
						"size":            "100",
						"raid_level":      "1",
						"iops":            "3000",
						"encrypted":       "true",
					}),
					resource.TestCheckResourceAttr(resourceName, "custom_json", `{"layer_key":"layer_value2"}`),
				),
			},
		},
	})
}

func testAccCheckLayerExists(n string, opslayer *opsworks.Layer) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribeLayersInput{
			LayerIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeLayers(params)

		if err != nil {
			return err
		}

		if v := len(resp.Layers); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opslayer = *resp.Layers[0]

		return nil
	}
}

func testAccCheckCreateLayerAttributes(
	opslayer *opsworks.Layer, stackName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opslayer.Name != stackName {
			return fmt.Errorf("Unexpected name: %s", *opslayer.Name)
		}

		if *opslayer.AutoAssignElasticIps {
			return fmt.Errorf(
				"Unexpected AutoAssignElasticIps: %t", *opslayer.AutoAssignElasticIps)
		}

		if !*opslayer.EnableAutoHealing {
			return fmt.Errorf(
				"Unexpected EnableAutoHealing: %t", *opslayer.EnableAutoHealing)
		}

		if !*opslayer.LifecycleEventConfiguration.Shutdown.DelayUntilElbConnectionsDrained {
			return fmt.Errorf(
				"Unexpected DelayUntilElbConnectionsDrained: %t",
				*opslayer.LifecycleEventConfiguration.Shutdown.DelayUntilElbConnectionsDrained)
		}

		if *opslayer.LifecycleEventConfiguration.Shutdown.ExecutionTimeout != 300 {
			return fmt.Errorf(
				"Unexpected ExecutionTimeout: %d",
				*opslayer.LifecycleEventConfiguration.Shutdown.ExecutionTimeout)
		}

		if v := len(opslayer.CustomSecurityGroupIds); v != 2 {
			return fmt.Errorf("Expected 2 customSecurityGroupIds, got %d", v)
		}

		expectedPackages := []*string{
			aws.String("git"),
			aws.String("golang"),
		}

		if !reflect.DeepEqual(expectedPackages, opslayer.Packages) {
			return fmt.Errorf("Unexpected Packages: %v", aws.StringValueSlice(opslayer.Packages))
		}

		expectedEbsVolumes := []*opsworks.VolumeConfiguration{
			{
				Encrypted:     aws.Bool(false),
				MountPoint:    aws.String("/home"),
				NumberOfDisks: aws.Int64(2),
				RaidLevel:     aws.Int64(0),
				Size:          aws.Int64(100),
				VolumeType:    aws.String("gp2"),
			},
		}

		if !reflect.DeepEqual(expectedEbsVolumes, opslayer.VolumeConfigurations) {
			return fmt.Errorf("Unnexpected VolumeConfiguration: %s", opslayer.VolumeConfigurations)
		}

		return nil
	}
}

func testAccCheckLayerDestroy(resourceType string, s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != resourceType {
			continue
		}
		req := &opsworks.DescribeLayersInput{
			LayerIds: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		_, err := conn.DescribeLayers(req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, opsworks.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
	}

	return fmt.Errorf("Fall through error on OpsWorks layer test")
}

func testAccCheckCustomLayerDestroy(s *terraform.State) error {
	return testAccCheckLayerDestroy("aws_opsworks_custom_layer", s)
}

func testAccCustomLayerSecurityGroups(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-layer1" {
  name = "%s-layer1"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-layer2" {
  name = "%s-layer2"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, name, name)
}

func testAccCustomLayerNoVPCCreateConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = aws_opsworks_stack.tf-acc.id
  name                   = "%s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]
  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300
  system_packages = [
    "git",
    "golang",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
    encrypted       = false
  }
}

%s

%s
`, name, testAccStackNoVPCCreateConfig(name), testAccCustomLayerSecurityGroups(name))
}

func testAccCustomLayerVPCCreateConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = aws_opsworks_stack.tf-acc.id
  name                   = "%s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  system_packages = [
    "git",
    "golang",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }
}

%s

%s
`, name, testAccStackVPCCreateConfig(name), testAccCustomLayerSecurityGroups(name))
}

func testAccCustomLayerUpdateConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-layer3" {
  name = "tf-ops-acc-layer-%[1]s"

  ingress {
    from_port   = 8
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_custom_layer" "tf-acc" {
  stack_id               = aws_opsworks_stack.tf-acc.id
  name                   = "%[1]s"
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = true
  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
    aws_security_group.tf-ops-acc-layer3.id,
  ]
  drain_elb_on_shutdown     = false
  instance_shutdown_timeout = 120
  system_packages = [
    "git",
    "golang",
    "subversion",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
    encrypted       = true
  }

  ebs_volume {
    type            = "io1"
    number_of_disks = 4
    mount_point     = "/var"
    size            = 100
    raid_level      = 1
    iops            = 3000
    encrypted       = true
  }

  custom_json = <<EOF
{
  "layer_key": "layer_value2"
}
EOF
}

%s

%s
`, name, testAccStackNoVPCCreateConfig(name), testAccCustomLayerSecurityGroups(name))
}

func testAccCustomLayerTags1Config(name, tagKey1, tagValue1 string) string {
	return testAccStackVPCCreateConfig(name) +
		testAccCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.tf-acc.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  system_packages = [
    "git",
    "golang",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, name, tagKey1, tagValue1)
}

func testAccCustomLayerTags2Config(name, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccStackVPCCreateConfig(name) +
		testAccCustomLayerSecurityGroups(name) +
		fmt.Sprintf(`
resource "aws_opsworks_custom_layer" "test" {
  stack_id               = aws_opsworks_stack.tf-acc.id
  name                   = %[1]q
  short_name             = "tf-ops-acc-custom-layer"
  auto_assign_public_ips = false

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-layer1.id,
    aws_security_group.tf-ops-acc-layer2.id,
  ]

  drain_elb_on_shutdown     = true
  instance_shutdown_timeout = 300

  system_packages = [
    "git",
    "golang",
  ]

  ebs_volume {
    type            = "gp2"
    number_of_disks = 2
    mount_point     = "/home"
    size            = 100
    raid_level      = 0
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, name, tagKey1, tagValue1, tagKey2, tagValue2)
}
