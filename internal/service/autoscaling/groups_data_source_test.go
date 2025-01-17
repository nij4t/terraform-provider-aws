package autoscaling_test

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccAutoScalingGroupsDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, autoscaling.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckGroupsConfig(sdkacctest.RandInt(), sdkacctest.RandInt(), sdkacctest.RandInt()),
			},
			{
				Config: testAccCheckGroupsWithDataSourceConfig(sdkacctest.RandInt(), sdkacctest.RandInt(), sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckGroups("data.aws_autoscaling_groups.group_list"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_groups.group_list", "names.#", "3"),
					resource.TestCheckResourceAttr("data.aws_autoscaling_groups.group_list", "arns.#", "3"),
				),
			},
		},
	})
}

func testAccCheckGroups(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find ASG resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AZ resource ID not set.")
		}

		actual, err := testAccCheckGroupsAvailable(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		expected := actual
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("ASG not sorted - expected %v, got %v", expected, actual)
		}
		return nil
	}
}

func testAccCheckGroupsAvailable(attrs map[string]string) ([]string, error) {
	v, ok := attrs["names.#"]
	if !ok {
		return nil, fmt.Errorf("Available ASG list is missing.")
	}
	qty, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	if qty < 1 {
		return nil, fmt.Errorf("No ASG found in region, this is probably a bug.")
	}
	zones := make([]string, qty)
	for n := range zones {
		zone, ok := attrs["names."+strconv.Itoa(n)]
		if !ok {
			return nil, fmt.Errorf("ASG list corrupt, this is definitely a bug.")
		}
		zones[n] = zone
	}
	return zones, nil
}

func testAccCheckGroupsConfig(rInt1, rInt2, rInt3 int) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t1.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  name               = "test-asg-%d"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "foo" {
  availability_zones = [data.aws_availability_zones.available.names[1]]
  name               = "test-asg-%d"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "barbaz" {
  availability_zones = [data.aws_availability_zones.available.names[2]]
  name               = "test-asg-%d"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}
`, rInt1, rInt2, rInt3)
}

func testAccCheckGroupsWithDataSourceConfig(rInt1, rInt2, rInt3 int) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
data "aws_ami" "test_ami" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_launch_configuration" "foobar" {
  image_id      = data.aws_ami.test_ami.id
  instance_type = "t1.micro"
}

resource "aws_autoscaling_group" "bar" {
  availability_zones = [data.aws_availability_zones.available.names[0]]
  name               = "test-asg-%d"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "foo" {
  availability_zones = [data.aws_availability_zones.available.names[1]]
  name               = "test-asg-%d"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

resource "aws_autoscaling_group" "barbaz" {
  availability_zones = [data.aws_availability_zones.available.names[2]]
  name               = "test-asg-%d"
  max_size           = 1
  min_size           = 0
  health_check_type  = "EC2"
  desired_capacity   = 0
  force_delete       = true

  launch_configuration = aws_launch_configuration.foobar.name

  tag {
    key                 = "Foo"
    value               = "foo-bar"
    propagate_at_launch = true
  }
}

data "aws_autoscaling_groups" "group_list" {
  filter {
    name   = "key"
    values = ["Foo"]
  }

  filter {
    name   = "value"
    values = ["foo-bar"]
  }
}
`, rInt1, rInt2, rInt3)
}
