package neptune_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccNeptuneOrderableDBInstanceDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_neptune_orderable_db_instance.test"
	engine := "neptune"
	engineVersion := "1.0.2.2"
	licenseModel := "amazon-license"
	class := "db.t3.medium"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckOrderableDBInstance(t) },
		ErrorCheck:   acctest.ErrorCheck(t, neptune.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableDBInstanceBasicDataSourceConfig(class, engine, engineVersion, licenseModel),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", class),
				),
			},
		},
	})
}

func TestAccNeptuneOrderableDBInstanceDataSource_preferred(t *testing.T) {
	dataSourceName := "data.aws_neptune_orderable_db_instance.test"
	engine := "neptune"
	engineVersion := "1.0.3.0"
	licenseModel := "amazon-license"
	preferredOption := "db.r4.2xlarge"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheckOrderableDBInstance(t) },
		ErrorCheck:   acctest.ErrorCheck(t, neptune.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccOrderableDBInstancePreferredDataSourceConfig(engine, engineVersion, licenseModel, preferredOption),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "engine", engine),
					resource.TestCheckResourceAttr(dataSourceName, "engine_version", engineVersion),
					resource.TestCheckResourceAttr(dataSourceName, "license_model", licenseModel),
					resource.TestCheckResourceAttr(dataSourceName, "instance_class", preferredOption),
				),
			},
		},
	})
}

func testAccPreCheckOrderableDBInstance(t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).NeptuneConn

	input := &neptune.DescribeOrderableDBInstanceOptionsInput{
		Engine: aws.String("mysql"),
	}

	_, err := conn.DescribeOrderableDBInstanceOptions(input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccOrderableDBInstanceBasicDataSourceConfig(class, engine, version, license string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  instance_class = %q
  engine         = %q
  engine_version = %q
  license_model  = %q
}
`, class, engine, version, license)
}

func testAccOrderableDBInstancePreferredDataSourceConfig(engine, version, license, preferredOption string) string {
	return fmt.Sprintf(`
data "aws_neptune_orderable_db_instance" "test" {
  engine         = %q
  engine_version = %q
  license_model  = %q

  preferred_instance_classes = [
    "db.xyz.xlarge",
    %q,
    "db.t3.small",
  ]
}
`, engine, version, license, preferredOption)
}
