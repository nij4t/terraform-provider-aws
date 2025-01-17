package efs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/efs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfefs "github.com/nij4t/terraform-provider-aws/internal/service/efs"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func TestAccEFSFileSystemPolicy_basic(t *testing.T) {
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEfsFileSystemPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicyExists(resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccFileSystemPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicyExists(resourceName, &desc),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func TestAccEFSFileSystemPolicy_disappears(t *testing.T) {
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEfsFileSystemPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicyExists(resourceName, &desc),
					acctest.CheckResourceDisappears(acctest.Provider, tfefs.ResourceFileSystemPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEFSFileSystemPolicy_policyBypass(t *testing.T) {
	var desc efs.DescribeFileSystemPolicyOutput
	resourceName := "aws_efs_file_system_policy.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, efs.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEfsFileSystemPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFileSystemPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicyExists(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "false"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"bypass_policy_lockout_safety_check"},
			},
			{
				Config: testAccFileSystemPolicyBypassConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEfsFileSystemPolicyExists(resourceName, &desc),
					resource.TestCheckResourceAttr(resourceName, "bypass_policy_lockout_safety_check", "true"),
				),
			},
		},
	})
}

func testAccCheckEfsFileSystemPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_efs_file_system_policy" {
			continue
		}

		_, err := tfefs.FindFileSystemPolicyByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("EFS File System Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckEfsFileSystemPolicyExists(n string, v *efs.DescribeFileSystemPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EFS File System Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EFSConn

		output, err := tfefs.FindFileSystemPolicyByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccFileSystemPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleSatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": [
                "elasticfilesystem:ClientMount",
                "elasticfilesystem:ClientWrite"
            ],
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName)
}

func testAccFileSystemPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleSatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": "elasticfilesystem:ClientMount",
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName)
}

func testAccFileSystemPolicyBypassConfig(rName string, bypass bool) string {
	return fmt.Sprintf(`
resource "aws_efs_file_system" "test" {
  creation_token = %[1]q
}

resource "aws_efs_file_system_policy" "test" {
  file_system_id = aws_efs_file_system.test.id

  bypass_policy_lockout_safety_check = %[2]t

  policy = <<POLICY
{
    "Version": "2012-10-17",
    "Id": "ExamplePolicy01",
    "Statement": [
        {
            "Sid": "ExampleSatement01",
            "Effect": "Allow",
            "Principal": {
                "AWS": "*"
            },
            "Resource": "${aws_efs_file_system.test.arn}",
            "Action": [
                "elasticfilesystem:ClientMount",
                "elasticfilesystem:ClientWrite"
            ],
            "Condition": {
                "Bool": {
                    "aws:SecureTransport": "true"
                }
            }
        }
    ]
}
POLICY
}
`, rName, bypass)
}
