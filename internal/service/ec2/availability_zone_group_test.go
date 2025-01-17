package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccEC2AvailabilityZoneGroup_optInStatus(t *testing.T) {
	resourceName := "aws_ec2_availability_zone_group.test"

	// Filter to one Availability Zone Group per Region as Local Zones become available
	// e.g. ensure there are not two us-west-2-XXX when adding to this list
	// (Not including in config to avoid lintignoring entire config.)
	localZone := "us-west-2-lax-1" // lintignore:AWSAT003 // currently the only generally available local zone

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckAvailabilityZoneGroup(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(localZone, ec2.AvailabilityZoneOptInStatusOptedIn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// InvalidOptInStatus: Opting out of Local Zones is not currently supported. Contact AWS Support for additional assistance.
			/*
				{
					Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusNotOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusNotOptedIn),
					),
				},
				{
					Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					),
				},
			*/
		},
	})
}

func testAccPreCheckAvailabilityZoneGroup(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("opt-in-status"),
				Values: aws.StringSlice([]string{
					ec2.AvailabilityZoneOptInStatusNotOptedIn,
					ec2.AvailabilityZoneOptInStatusOptedIn,
				}),
			},
		},
	}

	output, err := conn.DescribeAvailabilityZones(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if output == nil || len(output.AvailabilityZones) == 0 || output.AvailabilityZones[0] == nil {
		t.Skipf("skipping acceptance testing: no opt-in EC2 Availability Zone Groups found")
	}
}

func testAccEc2AvailabilityZoneGroupConfigOptInStatus(localZone, optInStatus string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  all_availability_zones = true

  filter {
    name = "group-name"
    values = [
      %[1]q,
    ]
  }
}

resource "aws_ec2_availability_zone_group" "test" {
  # The above group-name filter should ensure one Availability Zone Group per Region
  group_name    = tolist(data.aws_availability_zones.test.group_names)[0]
  opt_in_status = %[2]q
}
`, localZone, optInStatus)
}
