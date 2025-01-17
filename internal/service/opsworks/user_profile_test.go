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

func TestAccOpsWorksUserProfile_basic(t *testing.T) {
	rName := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	updateRName := fmt.Sprintf("test-user-%d", sdkacctest.RandInt())
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, opsworks.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckUserProfileDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccUserProfileCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(
						"aws_opsworks_user_profile.user", rName),
					resource.TestCheckResourceAttr(
						"aws_opsworks_user_profile.user", "ssh_public_key", "",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_user_profile.user", "ssh_username", rName,
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_user_profile.user", "allow_self_management", "false",
					),
				),
			},
			{
				Config: testAccUserProfileUpdate(rName, updateRName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUserProfileExists(
						"aws_opsworks_user_profile.user", updateRName),
					resource.TestCheckResourceAttr(
						"aws_opsworks_user_profile.user", "ssh_public_key", "",
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_user_profile.user", "ssh_username", updateRName,
					),
					resource.TestCheckResourceAttr(
						"aws_opsworks_user_profile.user", "allow_self_management", "false",
					),
				),
			},
		},
	})
}

func testAccCheckUserProfileExists(
	n, username string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		if _, ok := rs.Primary.Attributes["user_arn"]; !ok {
			return fmt.Errorf("User Profile user arn is missing, should be set.")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribeUserProfilesInput{
			IamUserArns: []*string{aws.String(rs.Primary.Attributes["user_arn"])},
		}
		resp, err := conn.DescribeUserProfiles(params)

		if err != nil {
			return err
		}

		if v := len(resp.UserProfiles); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		opsuserprofile := *resp.UserProfiles[0]

		if *opsuserprofile.AllowSelfManagement {
			return fmt.Errorf("Unnexpected allowSelfManagement: %t",
				*opsuserprofile.AllowSelfManagement)
		}

		if *opsuserprofile.Name != username {
			return fmt.Errorf("Unnexpected name: %s", *opsuserprofile.Name)
		}

		return nil
	}
}

func testAccCheckUserProfileDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_user_profile" {
			continue
		}

		req := &opsworks.DescribeUserProfilesInput{
			IamUserArns: []*string{aws.String(rs.Primary.Attributes["user_arn"])},
		}
		resp, err := client.DescribeUserProfiles(req)

		if err == nil {
			if len(resp.UserProfiles) > 0 {
				return fmt.Errorf("OpsWorks User Profiles still exist.")
			}
		}

		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}
	return nil
}

func testAccUserProfileCreate(rn string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_user_profile" "user" {
  user_arn     = aws_iam_user.user.arn
  ssh_username = aws_iam_user.user.name
}

resource "aws_iam_user" "user" {
  name = "%s"
  path = "/"
}
`, rn)
}

func testAccUserProfileUpdate(rn, updateRn string) string {
	return fmt.Sprintf(`
resource "aws_opsworks_user_profile" "user" {
  user_arn     = aws_iam_user.new-user.arn
  ssh_username = aws_iam_user.new-user.name
}

resource "aws_iam_user" "user" {
  name = "%s"
  path = "/"
}

resource "aws_iam_user" "new-user" {
  name = "%s"
  path = "/"
}
`, rn, updateRn)
}
