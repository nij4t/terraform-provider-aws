package securityhub_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfsecurityhub "github.com/nij4t/terraform-provider-aws/internal/service/securityhub"
)

func testAccProductSubscription_basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, securityhub.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAccountDestroy,
		Steps: []resource.TestStep{
			{
				// We would like to use an AWS product subscription, but they are
				// all automatically subscribed when enabling Security Hub.
				// This configuration will enable Security Hub, then in a later PreConfig,
				// we will disable an AWS product subscription so we can test (re-)enabling it.
				Config: testAccProductSubscriptionConfig_empty,
				Check:  testAccCheckAccountExists("aws_securityhub_account.example"),
			},
			{
				// AWS product subscriptions happen automatically when enabling Security Hub.
				// Here we attempt to remove one so we can attempt to (re-)enable it.
				PreConfig: func() {
					conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn
					productSubscriptionARN := arn.ARN{
						AccountID: acctest.AccountID(),
						Partition: acctest.Partition(),
						Region:    acctest.Region(),
						Resource:  "product-subscription/aws/guardduty",
						Service:   "securityhub",
					}.String()

					input := &securityhub.DisableImportFindingsForProductInput{
						ProductSubscriptionArn: aws.String(productSubscriptionARN),
					}

					_, err := conn.DisableImportFindingsForProduct(input)
					if err != nil {
						t.Fatalf("error disabling Security Hub Product Subscription for GuardDuty: %s", err)
					}
				},
				Config: testAccProductSubscriptionConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckProductSubscriptionExists("aws_securityhub_product_subscription.example"),
				),
			},
			{
				ResourceName:      "aws_securityhub_product_subscription.example",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Check Destroy - but only target the specific resource (otherwise Security Hub
				// will be disabled and the destroy check will fail)
				Config: testAccProductSubscriptionConfig_empty,
				Check:  testAccCheckProductSubscriptionDestroy,
			},
		},
	})
}

func testAccCheckProductSubscriptionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

		_, productSubscriptionArn, err := tfsecurityhub.ProductSubscriptionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		exists, err := tfsecurityhub.ProductSubscriptionCheckExists(conn, productSubscriptionArn)

		if err != nil {
			return err
		}

		if !exists {
			return fmt.Errorf("Security Hub product subscription %s not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckProductSubscriptionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).SecurityHubConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_securityhub_product_subscription" {
			continue
		}

		_, productSubscriptionArn, err := tfsecurityhub.ProductSubscriptionParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		exists, err := tfsecurityhub.ProductSubscriptionCheckExists(conn, productSubscriptionArn)

		if err != nil {
			return err
		}

		if exists {
			return fmt.Errorf("Security Hub product subscription %s still exists", rs.Primary.ID)
		}
	}

	return nil
}

const testAccProductSubscriptionConfig_empty = `
resource "aws_securityhub_account" "example" {}
`

const testAccProductSubscriptionConfig_basic = `
resource "aws_securityhub_account" "example" {}

data "aws_region" "current" {}

data "aws_partition" "current" {}

resource "aws_securityhub_product_subscription" "example" {
  depends_on  = [aws_securityhub_account.example]
  product_arn = "arn:${data.aws_partition.current.partition}:securityhub:${data.aws_region.current.name}::product/aws/guardduty"
}
`
