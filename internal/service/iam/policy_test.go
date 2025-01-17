package iam_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfiam "github.com/nij4t/terraform-provider-aws/internal/service/iam"
)

func TestAccIAMPolicy_basic(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"
	expectedPolicyText := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
`
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					acctest.CheckResourceAttrGlobalARN(resourceName, "arn", "iam", fmt.Sprintf("policy/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttr(resourceName, "policy", expectedPolicyText),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "policy_id"),
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

func TestAccIAMPolicy_description(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
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

func TestAccIAMPolicy_tags(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
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
				Config: testAccPolicyTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccPolicyTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccIAMPolicy_disappears(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					acctest.CheckResourceDisappears(acctest.Provider, tfiam.ResourcePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIAMPolicy_namePrefix(t *testing.T) {
	var out iam.GetPolicyOutput
	namePrefix := "tf-acc-test-"
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyNamePrefixConfig(namePrefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile(fmt.Sprintf("^%s", namePrefix))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccIAMPolicy_path(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyPathConfig(rName, "/path1/"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "path", "/path1/"),
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

func TestAccIAMPolicy_policy(t *testing.T) {
	var out iam.GetPolicyOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iam_policy.test"
	policy1 := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:Describe*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"
	policy2 := "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Action\":[\"ec2:*\"],\"Effect\":\"Allow\",\"Resource\":\"*\"}]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, iam.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccPolicyPolicyConfig(rName, "not-json"),
				ExpectError: regexp.MustCompile("invalid JSON"),
			},
			{
				Config: testAccPolicyPolicyConfig(rName, policy1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "policy", policy1),
				),
			},
			{
				Config: testAccPolicyPolicyConfig(rName, policy2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPolicyExists(resourceName, &out),
					resource.TestCheckResourceAttr(resourceName, "policy", policy2),
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

func testAccCheckPolicyExists(resource string, res *iam.GetPolicyOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resource]
		if !ok {
			return fmt.Errorf("Not found: %s", resource)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Policy name is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

		resp, err := conn.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(rs.Primary.Attributes["arn"]),
		})
		if err != nil {
			return err
		}

		*res = *resp

		return nil
	}
}

func testAccCheckPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).IAMConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iam_policy" {
			continue
		}

		_, err := conn.GetPolicy(&iam.GetPolicyInput{
			PolicyArn: aws.String(rs.Primary.ID),
		})

		if tfawserr.ErrMessageContains(err, iam.ErrCodeNoSuchEntityException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IAM Policy (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccPolicyDescriptionConfig(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  description = %q
  name        = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, description, rName)
}

func testAccPolicyNameConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName)
}

func testAccPolicyNamePrefixConfig(namePrefix string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name_prefix = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, namePrefix)
}

func testAccPolicyPathConfig(rName, path string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q
  path = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}
`, rName, path)
}

func testAccPolicyPolicyConfig(rName, policy string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name   = %q
  policy = %q
}
`, rName, policy)
}

func testAccPolicyTags1Config(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccPolicyTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test" {
  name = %q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
