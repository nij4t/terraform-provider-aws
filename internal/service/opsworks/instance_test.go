package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccOpsWorksInstance_basic(t *testing.T) {
	stackName := fmt.Sprintf("tf-%d", sdkacctest.RandInt())
	var opsinst opsworks.Instance
	resourceName := "aws_opsworks_instance.tf-acc"
	dataSourceName := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceCreateConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &opsinst),
					testAccCheckInstanceAttributes(&opsinst),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.micro"),
					resource.TestCheckResourceAttr(resourceName, "state", "stopped"),
					resource.TestCheckResourceAttr(resourceName, "layer_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "install_updates_on_boot", "true"),
					resource.TestCheckResourceAttr(resourceName, "architecture", "x86_64"),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
					resource.TestCheckResourceAttr(resourceName, "os", "Amazon Linux 2016.09"),                       // inherited from opsworks_stack_test
					resource.TestCheckResourceAttr(resourceName, "root_device_type", "ebs"),                          // inherited from opsworks_stack_test
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone", dataSourceName, "names.0"), // inherited from opsworks_stack_test
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"}, //state is something we pass to the API and get back as status :(
			},
			{
				Config: testAccInstanceUpdateConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &opsinst),
					testAccCheckInstanceAttributes(&opsinst),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "t2.small"),
					resource.TestCheckResourceAttr(resourceName, "layer_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "os", "Amazon Linux 2015.09"),
					resource.TestCheckResourceAttr(resourceName, "tenancy", "default"),
				),
			},
		},
	})
}

func TestAccOpsWorksInstance_updateHostNameForceNew(t *testing.T) {
	stackName := fmt.Sprintf("tf-%d", sdkacctest.RandInt())
	resourceName := "aws_opsworks_instance.tf-acc"
	var before, after opsworks.Instance

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceCreateConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"state"},
			},
			{
				Config: testAccInstanceUpdateHostNameConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInstanceExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "hostname", "tf-acc2"),
					testAccCheckInstanceRecreated(t, &before, &after),
				),
			},
		},
	})
}

func testAccCheckInstanceRecreated(t *testing.T,
	before, after *opsworks.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.InstanceId == *after.InstanceId {
			t.Fatalf("Expected change of OpsWorks Instance IDs, but both were %s", *before.InstanceId)
		}
		return nil
	}
}

func testAccCheckInstanceExists(
	n string, opsinst *opsworks.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Opsworks Instance is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribeInstancesInput{
			InstanceIds: []*string{&rs.Primary.ID},
		}
		resp, err := conn.DescribeInstances(params)

		if err != nil {
			return err
		}

		if v := len(resp.Instances); v != 1 {
			return fmt.Errorf("Expected 1 request returned, got %d", v)
		}

		*opsinst = *resp.Instances[0]

		return nil
	}
}

func testAccCheckInstanceAttributes(
	opsinst *opsworks.Instance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Depending on the timing, the state could be requested or stopped
		if *opsinst.Status != "stopped" && *opsinst.Status != "requested" {
			return fmt.Errorf("Unexpected request status: %s", *opsinst.Status)
		}
		if *opsinst.Architecture != "x86_64" {
			return fmt.Errorf("Unexpected architecture: %s", *opsinst.Architecture)
		}
		if *opsinst.Tenancy != "default" {
			return fmt.Errorf("Unexpected tenancy: %s", *opsinst.Tenancy)
		}
		if *opsinst.InfrastructureClass != "ec2" {
			return fmt.Errorf("Unexpected infrastructure class: %s", *opsinst.InfrastructureClass)
		}
		if *opsinst.RootDeviceType != "ebs" {
			return fmt.Errorf("Unexpected root device type: %s", *opsinst.RootDeviceType)
		}
		if *opsinst.VirtualizationType != "hvm" {
			return fmt.Errorf("Unexpected virtualization type: %s", *opsinst.VirtualizationType)
		}
		return nil
	}
}

func testAccCheckInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_instance" {
			continue
		}
		req := &opsworks.DescribeInstancesInput{
			InstanceIds: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		_, err := conn.DescribeInstances(req)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() == "ResourceNotFoundException" {
					// not found, good to go
					return nil
				}
			}
			return err
		}
	}

	return fmt.Errorf("Fall through error on OpsWorks instance test")
}

func testAccInstanceUpdateHostNameConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  layer_ids = [
    aws_opsworks_static_web_layer.tf-acc.id,
  ]

  instance_type = "t2.micro"
  state         = "stopped"
  hostname      = "tf-acc2"
}

%s
`, name, name, testAccStackVPCCreateConfig(name))
}

func testAccInstanceCreateConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  layer_ids = [
    aws_opsworks_static_web_layer.tf-acc.id,
  ]

  instance_type = "t2.micro"
  state         = "stopped"
  hostname      = "tf-acc1"
}

%s
`, name, name, testAccStackVPCCreateConfig(name))
}

func testAccInstanceUpdateConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "tf-ops-acc-web" {
  name = "%s-web"

  ingress {
    from_port   = 80
    to_port     = 80
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_security_group" "tf-ops-acc-php" {
  name = "%s-php"

  ingress {
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_opsworks_static_web_layer" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-web.id,
  ]
}

resource "aws_opsworks_php_app_layer" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  custom_security_group_ids = [
    aws_security_group.tf-ops-acc-php.id,
  ]
}

resource "aws_opsworks_instance" "tf-acc" {
  stack_id = aws_opsworks_stack.tf-acc.id

  layer_ids = [
    aws_opsworks_static_web_layer.tf-acc.id,
    aws_opsworks_php_app_layer.tf-acc.id,
  ]

  instance_type = "t2.small"
  state         = "stopped"
  hostname      = "tf-acc1"
  os            = "Amazon Linux 2015.09"

  timeouts {
    update = "15s"
  }
}

%s
`, name, name, testAccStackVPCCreateConfig(name))
}
