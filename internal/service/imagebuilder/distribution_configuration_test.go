package imagebuilder_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfimagebuilder "github.com/nij4t/terraform-provider-aws/internal/service/imagebuilder"
)

func TestAccImageBuilderDistributionConfiguration_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "imagebuilder", fmt.Sprintf("distribution-configuration/%s", rName)),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_created"),
					resource.TestCheckResourceAttr(resourceName, "date_updated", ""),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccImageBuilderDistributionConfiguration_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfimagebuilder.ResourceDistributionConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "description1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "description", "description2"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_distribution(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.FactoriesMultipleRegion(nil, 2),
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistribution2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "2"),
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

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_amiTags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationAMITagsConfig(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":               "1",
						"ami_distribution_configuration.0.ami_tags.%":    "1",
						"ami_distribution_configuration.0.ami_tags.key1": "value1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationAMITagsConfig(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":               "1",
						"ami_distribution_configuration.0.ami_tags.%":    "1",
						"ami_distribution_configuration.0.ami_tags.key2": "value2",
					}),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_description(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationDescriptionConfig(rName, "description1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":             "1",
						"ami_distribution_configuration.0.description": "description1",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationDescriptionConfig(rName, "description2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":             "1",
						"ami_distribution_configuration.0.description": "description2",
					}),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_kmsKeyID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	kmsKeyResourceName := "aws_kms_key.test"
	kmsKeyResourceName2 := "aws_kms_key.test2"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationKMSKeyId1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.kms_key_id", kmsKeyResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationKMSKeyId2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.ami_distribution_configuration.0.kms_key_id", kmsKeyResourceName2, "arn"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistributionLaunchPermission_userGroups(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationLaunchPermissionUserGroupsConfig(rName, "all"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_groups.*", "all"),
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

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistributionLaunchPermission_userIDs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationLaunchPermissionUserIDsConfig(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_ids.*", "111111111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationLaunchPermissionUserIDsConfig(rName, "222222222222"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.launch_permission.0.user_ids.*", "222222222222"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_name(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationNameConfig(rName, "name1-{{ imagebuilder:buildDate }}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":      "1",
						"ami_distribution_configuration.0.name": "name1-{{ imagebuilder:buildDate }}",
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationNameConfig(rName, "name2-{{ imagebuilder:buildDate }}"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "distribution.*", map[string]string{
						"ami_distribution_configuration.#":      "1",
						"ami_distribution_configuration.0.name": "name2-{{ imagebuilder:buildDate }}",
					}),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_DistributionAMIDistribution_targetAccountIDs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationTargetAccountIDsConfig(rName, "111111111111"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.target_account_ids.*", "111111111111"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionAMIDistributionConfigurationTargetAccountIDsConfig(rName, "222222222222"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "distribution.*.ami_distribution_configuration.0.target_account_ids.*", "222222222222"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_Distribution_licenseARNs(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	licenseConfigurationResourceName := "aws_licensemanager_license_configuration.test"
	licenseConfigurationResourceName2 := "aws_licensemanager_license_configuration.test2"
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationDistributionLicenseConfigurationARNs1Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.license_configuration_arns.*", licenseConfigurationResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDistributionConfigurationDistributionLicenseConfigurationARNs2Config(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					acctest.CheckResourceAttrRFC3339(resourceName, "date_updated"),
					resource.TestCheckResourceAttr(resourceName, "distribution.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "distribution.*.license_configuration_arns.*", licenseConfigurationResourceName2, "id"),
				),
			},
		},
	})
}

func TestAccImageBuilderDistributionConfiguration_tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_imagebuilder_distribution_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, imagebuilder.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckDistributionConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDistributionConfigurationTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
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
				Config: testAccDistributionConfigurationTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDistributionConfigurationTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDistributionConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckDistributionConfigurationDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_distribution_configuration" {
			continue
		}

		input := &imagebuilder.GetDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.GetDistributionConfiguration(input)

		if tfawserr.ErrCodeEquals(err, imagebuilder.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", rs.Primary.ID, err)
		}

		if output != nil {
			return fmt.Errorf("Image Builder Distribution Configuration (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckDistributionConfigurationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("resource not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ImageBuilderConn

		input := &imagebuilder.GetDistributionConfigurationInput{
			DistributionConfigurationArn: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetDistributionConfiguration(input)

		if err != nil {
			return fmt.Errorf("error getting Image Builder Distribution Configuration (%s): %w", rs.Primary.ID, err)
		}

		return nil
	}
}

func testAccDistributionConfigurationDescriptionConfig(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  description = %[2]q
  name        = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccDistributionConfigurationDistribution2Config(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_region" "alternate" {
  provider = awsalternate
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.alternate.name
  }
}
`, rName))
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationAMITagsConfig(rName string, amiTagKey string, amiTagValue string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      ami_tags = {
        %[2]q = %[3]q
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, amiTagKey, amiTagValue)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationDescriptionConfig(rName string, description string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      description = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, description)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationKMSKeyId1Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      kms_key_id = aws_kms_key.test.arn
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationKMSKeyId2Config(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test2" {
  deletion_window_in_days = 7
}

data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      kms_key_id = aws_kms_key.test2.arn
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationLaunchPermissionUserGroupsConfig(rName string, userGroup string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      launch_permission {
        user_groups = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, userGroup)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationLaunchPermissionUserIDsConfig(rName string, userId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      launch_permission {
        user_ids = [%[2]q]
      }
    }

    region = data.aws_region.current.name
  }
}
`, rName, userId)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationNameConfig(rName string, name string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = %[2]q
    }

    region = data.aws_region.current.name
  }
}
`, rName, name)
}

func testAccDistributionConfigurationDistributionAMIDistributionConfigurationTargetAccountIDsConfig(rName string, targetAccountId string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      target_account_ids = [%[2]q]
    }

    region = data.aws_region.current.name
  }
}
`, rName, targetAccountId)
}

func testAccDistributionConfigurationDistributionLicenseConfigurationARNs1Config(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_licensemanager_license_configuration" "test" {
  name                  = %[1]q
  license_counting_type = "Socket"
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    license_configuration_arns = [aws_licensemanager_license_configuration.test.id]
    region                     = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationDistributionLicenseConfigurationARNs2Config(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_licensemanager_license_configuration" "test2" {
  name                  = %[1]q
  license_counting_type = "Socket"
}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    license_configuration_arns = [aws_licensemanager_license_configuration.test2.id]
    region                     = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationNameConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }
}
`, rName)
}

func testAccDistributionConfigurationTags1Config(rName string, tagKey1 string, tagValue1 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDistributionConfigurationTags2Config(rName string, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_imagebuilder_distribution_configuration" "test" {
  name = %[1]q

  distribution {
    ami_distribution_configuration {
      name = "{{ imagebuilder:buildDate }}"
    }

    region = data.aws_region.current.name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
