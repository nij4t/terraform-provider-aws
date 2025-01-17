package worklink_test

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfworklink "github.com/nij4t/terraform-provider-aws/internal/service/worklink"
)

func TestAccWorkLinkFleet_basic(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "company_code"),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
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

func TestAccWorkLinkFleet_displayName(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetDisplayNameConfig(suffix, "display1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "display1"),
				),
			},
			{
				Config: testAccFleetDisplayNameConfig(suffix, "display2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", "display2"),
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

func TestAccWorkLinkFleet_optimizeForEndUserLocation(t *testing.T) {
	suffix := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetOptimizeForEndUserLocationConfig(suffix, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "false"),
				),
			},
			{
				Config: testAccFleetOptimizeForEndUserLocationConfig(suffix, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "optimize_for_end_user_location", "true"),
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

func TestAccWorkLinkFleet_auditStreamARN(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetAuditStreamARNConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "audit_stream_arn", "aws_kinesis_stream.test_stream", "arn"),
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

func TestAccWorkLinkFleet_network(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetNetworkConfig(rName, "192.168.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccFleetNetworkConfig(rName, "10.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "network.0.vpc_id", "aws_vpc.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "network.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network.0.security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccFleetConfig(rName),
				ExpectError: regexp.MustCompile(`Company Network Configuration cannot be removed`),
			},
		},
	})
}

func TestAccWorkLinkFleet_deviceCaCertificate(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"
	fName := "test-fixtures/worklink-device-ca-certificate.pem"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetDeviceCaCertificateConfig(rName, fName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestMatchResourceAttr(resourceName, "device_ca_certificate", regexp.MustCompile("^-----BEGIN CERTIFICATE-----")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "device_ca_certificate", ""),
				),
			},
		},
	})
}

func TestAccWorkLinkFleet_identityProvider(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"
	fName := "test-fixtures/saml-metadata.xml"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetIdentityProviderConfig(rName, fName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "identity_provider.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "identity_provider.0.type", "SAML"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:      testAccFleetConfig(rName),
				ExpectError: regexp.MustCompile(`Identity Provider Configuration cannot be removed`),
			},
		},
	})
}

func TestAccWorkLinkFleet_disappears(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(20, sdkacctest.CharSetAlpha)
	resourceName := "aws_worklink_fleet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, worklink.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckFleetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFleetConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetExists(resourceName),
					testAccCheckFleetDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn

		input := &worklink.DeleteFleetInput{
			FleetArn: aws.String(rs.Primary.ID),
		}

		if _, err := conn.DeleteFleet(input); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"DELETING"},
			Target:     []string{"DELETED"},
			Refresh:    tfworklink.FleetStateRefresh(conn, rs.Primary.ID),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err := stateConf.WaitForState()

		return err
	}
}

func testAccCheckFleetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_worklink_fleet" {
			continue
		}

		_, err := conn.DescribeFleetMetadata(
			&worklink.DescribeFleetMetadataInput{
				FleetArn: aws.String(rs.Primary.ID),
			})

		if err != nil {
			// Return nil if the Worklink Fleet is already destroyed
			if tfawserr.ErrMessageContains(err, worklink.ErrCodeResourceNotFoundException, "") {
				return nil
			}

			return err
		}
		return fmt.Errorf("Worklink Fleet %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckFleetExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worklink Fleet ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn
		_, err := conn.DescribeFleetMetadata(&worklink.DescribeFleetMetadataInput{
			FleetArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccPreCheck(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).WorkLinkConn

	input := &worklink.ListFleetsInput{
		MaxResults: aws.Int64(1),
	}

	_, err := conn.ListFleets(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccFleetConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"
}
`, r)
}

func testAccFleetDisplayNameConfig(r, displayName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name         = "tf-worklink-fleet-%s"
  display_name = "%s"
}
`, r, displayName)
}

func testAccFleetOptimizeForEndUserLocationConfig(r string, b bool) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name                           = "tf-worklink-fleet-%s"
  optimize_for_end_user_location = %t
}
`, r, b)
}

func testAccFleetNetworkConfig_Base(rName, cidrBlock string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "%s"

  tags = {
    Name = %q
  }
}

resource "aws_security_group" "test" {
  name        = "tf_test_foo"
  description = "foo"
  vpc_id      = aws_vpc.test.id

  ingress {
    protocol  = "icmp"
    from_port = -1
    to_port   = -1
    self      = true
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %q
  }
}
`, cidrBlock, rName, rName)
}

func testAccFleetNetworkConfig(r, cidrBlock string) string {
	return fmt.Sprintf(`
%s

resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  network {
    vpc_id             = aws_vpc.test.id
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }
}
`, testAccFleetNetworkConfig_Base(r, cidrBlock), r)
}

func testAccFleetAuditStreamARNConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test_stream" {
  name        = "AmazonWorkLink-%s_kinesis_test"
  shard_count = 1
}

resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  audit_stream_arn = aws_kinesis_stream.test_stream.arn
}
`, r, r)
}

func testAccFleetDeviceCaCertificateConfig(r string, fName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  device_ca_certificate = file("%s")
}
`, r, fName)
}

func testAccFleetIdentityProviderConfig(r string, fName string) string {
	return fmt.Sprintf(`
resource "aws_worklink_fleet" "test" {
  name = "tf-worklink-fleet-%s"

  identity_provider {
    type          = "SAML"
    saml_metadata = file("%s")
  }
}
`, r, fName)
}
