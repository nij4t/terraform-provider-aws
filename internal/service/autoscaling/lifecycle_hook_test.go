package autoscaling_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccAutoScalingLifecycleHook_basic(t *testing.T) {
	resourceName := fmt.Sprintf("tf-test-%s", sdkacctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLifecycleHookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleHookConfig(resourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecycleHookExists("aws_autoscaling_lifecycle_hook.foobar"),
					resource.TestCheckResourceAttr("aws_autoscaling_lifecycle_hook.foobar", "autoscaling_group_name", resourceName),
					resource.TestCheckResourceAttr("aws_autoscaling_lifecycle_hook.foobar", "default_result", "CONTINUE"),
					resource.TestCheckResourceAttr("aws_autoscaling_lifecycle_hook.foobar", "heartbeat_timeout", "2000"),
					resource.TestCheckResourceAttr("aws_autoscaling_lifecycle_hook.foobar", "lifecycle_transition", "autoscaling:EC2_INSTANCE_LAUNCHING"),
				),
			},
			{
				ResourceName:      "aws_autoscaling_lifecycle_hook.foobar",
				ImportState:       true,
				ImportStateIdFunc: testAccLifecycleHookImportStateIdFunc("aws_autoscaling_lifecycle_hook.foobar"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAutoScalingLifecycleHook_omitDefaultResult(t *testing.T) {
	rName := sdkacctest.RandString(10)
	rInt := sdkacctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckLifecycleHookDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccLifecycleHookConfig_omitDefaultResult(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckLifecycleHookExists("aws_autoscaling_lifecycle_hook.foobar"),
					resource.TestCheckResourceAttr("aws_autoscaling_lifecycle_hook.foobar", "default_result", "ABANDON"),
				),
			},
		},
	})
}

func testAccCheckLifecycleHookExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		return checkLifecycleHookExistsByName(
			rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.ID)
	}
}

func checkLifecycleHookExistsByName(asgName, hookName string) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn
	params := &autoscaling.DescribeLifecycleHooksInput{
		AutoScalingGroupName: aws.String(asgName),
		LifecycleHookNames:   []*string{aws.String(hookName)},
	}
	resp, err := conn.DescribeLifecycleHooks(params)
	if err != nil {
		return err
	}
	if len(resp.LifecycleHooks) == 0 {
		return fmt.Errorf("LifecycleHook not found")
	}

	return nil
}

func testAccCheckLifecycleHookDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).AutoScalingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_autoscaling_group" {
			continue
		}

		params := autoscaling.DescribeLifecycleHooksInput{
			AutoScalingGroupName: aws.String(rs.Primary.Attributes["autoscaling_group_name"]),
			LifecycleHookNames:   []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeLifecycleHooks(&params)

		if err == nil {
			if len(resp.LifecycleHooks) != 0 &&
				*resp.LifecycleHooks[0].LifecycleHookName == rs.Primary.ID {
				return fmt.Errorf("Lifecycle Hook Still Exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccLifecycleHookImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["autoscaling_group_name"], rs.Primary.Attributes["name"]), nil
	}
}

func testAccLifecycleHookConfig(name string) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
resource "aws_launch_configuration" "foobar" {
  name          = "%s"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_sqs_queue" "foobar" {
  name                      = "foobar"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_role" "foobar" {
  name = "foobar"

  assume_role_policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [ {
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": [ "sts:AssumeRole" ]
  } ]
}
EOF
}

resource "aws_iam_role_policy" "foobar" {
  name = "foobar"
  role = aws_iam_role.foobar.id

  policy = <<EOF
{
    "Version" : "2012-10-17",
    "Statement": [ {
      "Effect": "Allow",
      "Action": [
	"sqs:SendMessage",
	"sqs:GetQueueUrl",
	"sns:Publish"
      ],
      "Resource": [
	"${aws_sqs_queue.foobar.arn}"
      ]
    } ]
}
EOF
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "foobar" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "%s"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_lifecycle_hook" "foobar" {
  name                   = "foobar"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
  default_result         = "CONTINUE"
  heartbeat_timeout      = 2000
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = <<EOF
{
  "foo": "bar"
}
EOF

  notification_target_arn = aws_sqs_queue.foobar.arn
  role_arn                = aws_iam_role.foobar.arn
}
`, name, name)
}

func testAccLifecycleHookConfig_omitDefaultResult(name string, rInt int) string {
	return acctest.ConfigLatestAmazonLinuxHvmEbsAmi() + fmt.Sprintf(`
resource "aws_launch_configuration" "foobar" {
  name          = "%s"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "t1.micro"
}

resource "aws_sqs_queue" "foobar" {
  name                      = "foobar-%d"
  delay_seconds             = 90
  max_message_size          = 2048
  message_retention_seconds = 86400
  receive_wait_time_seconds = 10
}

resource "aws_iam_role" "foobar" {
  name = "foobar-%d"

  assume_role_policy = <<EOF
{
  "Version" : "2012-10-17",
  "Statement": [ {
    "Effect": "Allow",
    "Principal": {"AWS": "*"},
    "Action": [ "sts:AssumeRole" ]
  } ]
}
EOF
}

resource "aws_iam_role_policy" "foobar" {
  name = "foobar-%d"
  role = aws_iam_role.foobar.id

  policy = <<EOF
{
    "Version" : "2012-10-17",
    "Statement": [ {
      "Effect": "Allow",
      "Action": [
	"sqs:SendMessage",
	"sqs:GetQueueUrl",
	"sns:Publish"
      ],
      "Resource": [
	"${aws_sqs_queue.foobar.arn}"
      ]
    } ]
}
EOF
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_autoscaling_group" "foobar" {
  availability_zones        = [data.aws_availability_zones.available.names[1]]
  name                      = "%s"
  max_size                  = 5
  min_size                  = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  force_delete              = true
  termination_policies      = ["OldestInstance"]
  launch_configuration      = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_lifecycle_hook" "foobar" {
  name                   = "foobar-%d"
  autoscaling_group_name = aws_autoscaling_group.foobar.name
  heartbeat_timeout      = 2000
  lifecycle_transition   = "autoscaling:EC2_INSTANCE_LAUNCHING"

  notification_metadata = <<EOF
{
  "foo": "bar"
}
EOF

  notification_target_arn = aws_sqs_queue.foobar.arn
  role_arn                = aws_iam_role.foobar.arn
}
`, name, rInt, rInt, rInt, name, rInt)
}
