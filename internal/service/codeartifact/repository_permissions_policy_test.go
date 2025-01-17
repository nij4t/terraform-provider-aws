package codeartifact_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/codeartifact"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfcodeartifact "github.com/nij4t/terraform-provider-aws/internal/service/codeartifact"
)

func testAccCodeArtifactRepositoryPermissionsPolicy_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_repository.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRepositoryPermissionsPolicyUpdatedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_repository.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:ListRepositoriesInDomain")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
				),
			},
		},
	})
}

func testAccCodeArtifactRepositoryPermissionsPolicy_owner(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyOwnerConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "resource_arn", "aws_codeartifact_repository.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain", rName),
					resource.TestMatchResourceAttr(resourceName, "policy_document", regexp.MustCompile("codeartifact:CreateRepository")),
					resource.TestCheckResourceAttrPair(resourceName, "domain_owner", "aws_codeartifact_domain.test", "owner"),
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

func testAccCodeArtifactRepositoryPermissionsPolicy_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodeartifact.ResourceRepositoryPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCodeArtifactRepositoryPermissionsPolicy_Disappears_domain(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_codeartifact_repository_permissions_policy.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(codeartifact.EndpointsID, t) },
		ErrorCheck:   acctest.ErrorCheck(t, codeartifact.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRepositoryPermissionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRepositoryPermissionsPolicyBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRepositoryPermissionsExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfcodeartifact.ResourceRepositoryPermissionsPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckRepositoryPermissionsExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no CodeArtifact domain set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn

		domainOwner, domainName, repoName, err := tfcodeartifact.DecodeRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.GetRepositoryPermissionsPolicy(&codeartifact.GetRepositoryPermissionsPolicyInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
			Repository:  aws.String(repoName),
		})

		return err
	}
}

func testAccCheckRepositoryPermissionsDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_codeartifact_repository_permissions_policy" {
			continue
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).CodeArtifactConn

		domainOwner, domainName, repoName, err := tfcodeartifact.DecodeRepositoryID(rs.Primary.ID)
		if err != nil {
			return err
		}

		resp, err := conn.GetRepositoryPermissionsPolicy(&codeartifact.GetRepositoryPermissionsPolicyInput{
			Domain:      aws.String(domainName),
			DomainOwner: aws.String(domainOwner),
			Repository:  aws.String(repoName),
		})

		if err == nil {
			if aws.StringValue(resp.Policy.ResourceArn) == rs.Primary.ID {
				return fmt.Errorf("CodeArtifact Domain %s still exists", rs.Primary.ID)
			}
		}

		if tfawserr.ErrMessageContains(err, codeartifact.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		return err
	}

	return nil
}

func testAccRepositoryPermissionsPolicyBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPermissionsPolicyOwnerConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  domain_owner    = aws_codeartifact_domain.test.owner
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": "codeartifact:CreateRepository",
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}

func testAccRepositoryPermissionsPolicyUpdatedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
}

resource "aws_codeartifact_domain" "test" {
  domain         = %[1]q
  encryption_key = aws_kms_key.test.arn
}

resource "aws_codeartifact_repository" "test" {
  repository = %[1]q
  domain     = aws_codeartifact_domain.test.domain
}

resource "aws_codeartifact_repository_permissions_policy" "test" {
  domain          = aws_codeartifact_domain.test.domain
  repository      = aws_codeartifact_repository.test.repository
  policy_document = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Action": [
 				"codeartifact:CreateRepository",
				"codeartifact:ListRepositoriesInDomain"
			],
            "Effect": "Allow",
            "Principal": "*",
            "Resource": "${aws_codeartifact_domain.test.arn}"
        }
    ]
}
EOF
}
`, rName)
}
