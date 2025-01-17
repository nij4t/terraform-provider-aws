package glue_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func testAccGlueDataCatalogEncryptionSettings_basic(t *testing.T) {
	var settings glue.DataCatalogEncryptionSettings

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_glue_data_catalog_encryption_settings.test"
	keyResourceName := "aws_kms_key.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, glue.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDataCatalogEncryptionSettingsNonEncryptedConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDataCatalogEncryptionSettingsEncryptedConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", keyResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "SSE-KMS"),
					resource.TestCheckResourceAttrPair(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", keyResourceName, "arn"),
				),
			},
			{
				Config: testAccDataCatalogEncryptionSettingsNonEncryptedConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDataCatalogEncryptionSettingsExists(resourceName, &settings),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.return_connection_password_encrypted", "false"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.connection_password_encryption.0.aws_kms_key_id", ""),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.catalog_encryption_mode", "DISABLED"),
					resource.TestCheckResourceAttr(resourceName, "data_catalog_encryption_settings.0.encryption_at_rest.0.sse_aws_kms_key_id", ""),
				),
			},
		},
	})
}

func testAccCheckDataCatalogEncryptionSettingsExists(resourceName string, v *glue.DataCatalogEncryptionSettings) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Glue Data Catalog Encryption Settings ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).GlueConn

		output, err := conn.GetDataCatalogEncryptionSettings(&glue.GetDataCatalogEncryptionSettingsInput{
			CatalogId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		*v = *output.DataCatalogEncryptionSettings

		return nil
	}
}

func testAccDataCatalogEncryptionSettingsEncryptedConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q
  policy      = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_glue_data_catalog_encryption_settings" "test" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      aws_kms_key_id                       = aws_kms_key.test.arn
      return_connection_password_encrypted = true
    }

    encryption_at_rest {
      catalog_encryption_mode = "SSE-KMS"
      sse_aws_kms_key_id      = aws_kms_key.test.arn
    }
  }
}
`, rName)
}

func testAccDataCatalogEncryptionSettingsNonEncryptedConfig() string {
	return `
resource "aws_glue_data_catalog_encryption_settings" "test" {
  data_catalog_encryption_settings {
    connection_password_encryption {
      return_connection_password_encrypted = false
    }

    encryption_at_rest {
      catalog_encryption_mode = "DISABLED"
    }
  }
}
`
}
