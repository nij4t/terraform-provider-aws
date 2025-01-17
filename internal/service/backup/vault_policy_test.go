package backup_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/backup"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfbackup "github.com/nij4t/terraform-provider-aws/internal/service/backup"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func TestAccBackupVaultPolicy_basic(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Version\":\"2012-10-17\".+"))),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBackupVaultPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("^{\"Version\":\"2012-10-17\".+")),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("backup:ListRecoveryPointsByBackupVault")),
				),
			},
		},
	})
}

func TestAccBackupVaultPolicy_disappears(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVaultPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBackupVaultPolicy_Disappears_vault(t *testing.T) {
	var vault backup.GetBackupVaultAccessPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_backup_vault_policy.test"
	vaultResourceName := "aws_backup_vault.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, backup.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckVaultPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBackupVaultPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVaultPolicyExists(resourceName, &vault),
					acctest.CheckResourceDisappears(acctest.Provider, tfbackup.ResourceVault(), vaultResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVaultPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_backup_vault_policy" {
			continue
		}

		_, err := tfbackup.FindBackupVaultAccessPolicyByName(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Backup Vault Policy %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVaultPolicyExists(name string, vault *backup.GetBackupVaultAccessPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Backup Vault Policy ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).BackupConn

		output, err := tfbackup.FindBackupVaultAccessPolicyByName(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*vault = *output

		return nil
	}
}

func testAccBackupVaultPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "default",
  "Statement": [
    {
      "Sid": "default",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
		"backup:DescribeBackupVault",
		"backup:DeleteBackupVault",
		"backup:PutBackupVaultAccessPolicy",
		"backup:DeleteBackupVaultAccessPolicy",
		"backup:GetBackupVaultAccessPolicy",
		"backup:StartBackupJob",
		"backup:GetBackupVaultNotifications",
		"backup:PutBackupVaultNotifications"
      ],
      "Resource": "${aws_backup_vault.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}

func testAccBackupVaultPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_backup_vault" "test" {
  name = %[1]q
}

resource "aws_backup_vault_policy" "test" {
  backup_vault_name = aws_backup_vault.test.name

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "default",
  "Statement": [
    {
      "Sid": "default",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
		"backup:DescribeBackupVault",
		"backup:DeleteBackupVault",
		"backup:PutBackupVaultAccessPolicy",
		"backup:DeleteBackupVaultAccessPolicy",
		"backup:GetBackupVaultAccessPolicy",
		"backup:StartBackupJob",
		"backup:GetBackupVaultNotifications",
		"backup:PutBackupVaultNotifications",
		"backup:ListRecoveryPointsByBackupVault"
      ],
      "Resource": "${aws_backup_vault.test.arn}"
    }
  ]
}
POLICY
}
`, rName)
}
