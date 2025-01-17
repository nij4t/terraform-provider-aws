package lambda_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccLambdaCodeSigningConfigDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningBasicDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfigDataSource_policyID(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningurePolicyDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policies", resourceName, "policies"),
					resource.TestCheckResourceAttrPair(dataSourceName, "config_id", resourceName, "config_id"),
				),
			},
		},
	})
}

func TestAccLambdaCodeSigningConfigDataSource_description(t *testing.T) {
	dataSourceName := "data.aws_lambda_code_signing_config.test"
	resourceName := "aws_lambda_code_signing_config.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, lambda.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccCodeSigningureDescriptionDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "allowed_publishers.0.signing_profile_version_arns.#", resourceName, "allowed_publishers.0.signing_profile_version_arns.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
				),
			},
		},
	})
}

const testAccCodeSigningBasicDataSourceConfig = `
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn
    ]
  }
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`

const testAccCodeSigningurePolicyDataSourceConfig = `
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn
    ]
  }

  policies {
    untrusted_artifact_on_deployment = "Warn"
  }
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`

const testAccCodeSigningureDescriptionDataSourceConfig = `
resource "aws_signer_signing_profile" "test" {
  platform_id = "AWSLambda-SHA384-ECDSA"
}

resource "aws_lambda_code_signing_config" "test" {
  allowed_publishers {
    signing_profile_version_arns = [
      aws_signer_signing_profile.test.version_arn
    ]
  }

  description = "Code Signing Config for app A"
}

data "aws_lambda_code_signing_config" "test" {
  arn = aws_lambda_code_signing_config.test.arn
}
`
