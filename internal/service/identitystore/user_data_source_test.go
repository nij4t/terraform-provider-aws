package identitystore_test

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/identitystore"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccIdentityStoreUserDataSource_userName(t *testing.T) {
	dataSourceName := "data.aws_identitystore_user.test"
	name := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(t)
			testAccPreCheckUserName(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, identitystore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccUserDisplayNameDataSourceConfig(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "user_id"),
					resource.TestCheckResourceAttr(dataSourceName, "user_name", name),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_userID(t *testing.T) {
	dataSourceName := "data.aws_identitystore_user.test"
	name := os.Getenv("AWS_IDENTITY_STORE_USER_NAME")
	userID := os.Getenv("AWS_IDENTITY_STORE_USER_ID")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			testAccPreCheckSSOAdminInstances(t)
			testAccPreCheckUserName(t)
			testAccPreCheckUserID(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, identitystore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccUserUserIDDataSourceConfig(name, userID),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "user_id", userID),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_name"),
				),
			},
		},
	})
}

func TestAccIdentityStoreUserDataSource_nonExistent(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckSSOAdminInstances(t) },
		ErrorCheck:   acctest.ErrorCheck(t, identitystore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config:      testAccUserNonExistentDataSourceConfig,
				ExpectError: regexp.MustCompile(`no Identity Store User found matching criteria`),
			},
		},
	})
}

func testAccPreCheckUserName(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_NAME") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_NAME env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}

func testAccPreCheckUserID(t *testing.T) {
	if os.Getenv("AWS_IDENTITY_STORE_USER_ID") == "" {
		t.Skip("AWS_IDENTITY_STORE_USER_ID env var must be set for AWS Identity Store User acceptance test. " +
			"This is required until ListUsers API returns results without filtering by name.")
	}
}

func testAccUserDisplayNameDataSourceConfig(name string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  filter {
    attribute_path  = "UserName"
    attribute_value = %q
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name)
}

func testAccUserUserIDDataSourceConfig(name, id string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  filter {
    attribute_path  = "UserName"
    attribute_value = %q
  }

  user_id = %q

  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`, name, id)
}

const testAccUserNonExistentDataSourceConfig = `
data "aws_ssoadmin_instances" "test" {}

data "aws_identitystore_user" "test" {
  filter {
    attribute_path  = "UserName"
    attribute_value = "does-not-exist"
  }
  identity_store_id = tolist(data.aws_ssoadmin_instances.test.identity_store_ids)[0]
}
`
