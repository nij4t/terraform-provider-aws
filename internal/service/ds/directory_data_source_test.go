package ds_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/directoryservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

func TestAccDirectoryServiceDirectoryDataSource_nonExistent(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config:      testAccDirectoryDataSourceConfig_NonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func TestAccDirectoryServiceDirectoryDataSource_simpleAD(t *testing.T) {
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test-simple-ad"
	dataSourceName := "data.aws_directory_service_directory.test-simple-ad"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t); acctest.PreCheckDirectoryServiceSimpleDirectory(t) },
		ErrorCheck: acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_SimpleAD(alias, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", directoryservice.DirectoryTypeSimpleAd),
					resource.TestCheckResourceAttr(dataSourceName, "size", "Small"),
					resource.TestCheckResourceAttr(dataSourceName, "name", domainName),
					resource.TestCheckResourceAttr(dataSourceName, "description", "tf-testacc SimpleAD"),
					resource.TestCheckResourceAttr(dataSourceName, "short_name", "corp"),
					resource.TestCheckResourceAttr(dataSourceName, "alias", alias),
					resource.TestCheckResourceAttr(dataSourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.vpc_id", resourceName, "vpc_settings.0.vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.subnet_ids", resourceName, "vpc_settings.0.subnet_ids"),
					resource.TestCheckResourceAttr(dataSourceName, "access_url", fmt.Sprintf("%s.awsapps.com", alias)),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_ip_addresses", resourceName, "dns_ip_addresses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
				),
			},
		},
	})
}

func TestAccDirectoryServiceDirectoryDataSource_microsoftAD(t *testing.T) {
	alias := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_directory_service_directory.test-microsoft-ad"
	dataSourceName := "data.aws_directory_service_directory.test-microsoft-ad"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDirectoryDataSourceConfig_MicrosoftAD(alias, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "type", directoryservice.DirectoryTypeMicrosoftAd),
					resource.TestCheckResourceAttr(dataSourceName, "edition", "Standard"),
					resource.TestCheckResourceAttr(dataSourceName, "name", domainName),
					resource.TestCheckResourceAttr(dataSourceName, "description", "tf-testacc MicrosoftAD"),
					resource.TestCheckResourceAttr(dataSourceName, "short_name", "corp"),
					resource.TestCheckResourceAttr(dataSourceName, "alias", alias),
					resource.TestCheckResourceAttr(dataSourceName, "enable_sso", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "vpc_settings.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.vpc_id", resourceName, "vpc_settings.0.vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_settings.0.subnet_ids", resourceName, "vpc_settings.0.subnet_ids"),
					resource.TestCheckResourceAttr(dataSourceName, "access_url", fmt.Sprintf("%s.awsapps.com", alias)),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_ip_addresses", resourceName, "dns_ip_addresses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_id", resourceName, "security_group_id"),
				),
			},
		},
	})
}

func TestAccDirectoryServiceDirectoryDataSource_connector(t *testing.T) {
	resourceName := "aws_directory_service_directory.connector"
	dataSourceName := "data.aws_directory_service_directory.test-ad-connector"

	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckDirectoryService(t)
			acctest.PreCheckDirectoryServiceSimpleDirectory(t)
		},
		ErrorCheck: acctest.ErrorCheck(t, directoryservice.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceDirectoryServiceDirectoryConfig_connector(domainName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "connect_settings.0.connect_ips", resourceName, "connect_settings.0.connect_ips"),
				),
			},
		},
	})
}

const testAccDirectoryDataSourceConfig_NonExistent = `
data "aws_directory_service_directory" "test" {
  directory_id = "d-abc0123456"
}
`

func testAccDirectoryDataSourceConfig_Prerequisites(adType string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-testacc-%[1]s"
  }
}

resource "aws_subnet" "primary" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-testacc-%[1]s-primary"
  }
}

resource "aws_subnet" "secondary" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "tf-testacc-%[1]s-secondary"
  }
}
`, adType))
}

func testAccDirectoryDataSourceConfig_SimpleAD(alias, domain string) string {
	return acctest.ConfigCompose(testAccDirectoryDataSourceConfig_Prerequisites("simple-ad"), fmt.Sprintf(`
resource "aws_directory_service_directory" "test-simple-ad" {
  type        = "SimpleAD"
  size        = "Small"
  name        = %[2]q
  description = "tf-testacc SimpleAD"
  short_name  = "corp"
  password    = "#S1ncerely"

  alias      = %[1]q
  enable_sso = false

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.primary.id, aws_subnet.secondary.id]
  }
}

data "aws_directory_service_directory" "test-simple-ad" {
  directory_id = aws_directory_service_directory.test-simple-ad.id
}
`, alias, domain))
}

func testAccDirectoryDataSourceConfig_MicrosoftAD(alias, domain string) string {
	return acctest.ConfigCompose(testAccDirectoryDataSourceConfig_Prerequisites("microsoft-ad"), fmt.Sprintf(`
resource "aws_directory_service_directory" "test-microsoft-ad" {
  type        = "MicrosoftAD"
  edition     = "Standard"
  name        = %[2]q
  description = "tf-testacc MicrosoftAD"
  short_name  = "corp"
  password    = "#S1ncerely"

  alias      = %[1]q
  enable_sso = false

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.primary.id, aws_subnet.secondary.id]
  }
}

data "aws_directory_service_directory" "test-microsoft-ad" {
  directory_id = aws_directory_service_directory.test-microsoft-ad.id
}
`, alias, domain))
}

func testAccDataSourceDirectoryServiceDirectoryConfig_connector(domain string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_directory_service_directory" "test" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.foo.id, aws_subnet.test.id]
  }
}

resource "aws_directory_service_directory" "connector" {
  name     = %[1]q
  password = "SuperSecretPassw0rd"
  size     = "Small"
  type     = "ADConnector"

  connect_settings {
    customer_dns_ips  = aws_directory_service_directory.test.dns_ip_addresses
    customer_username = "Administrator"
    vpc_id            = aws_vpc.main.id
    subnet_ids        = [aws_subnet.foo.id, aws_subnet.test.id]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-directory-service-directory-connector"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-directory-service-directory-connector-foo"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "tf-acc-directory-service-directory-connector-test"
  }
}

data "aws_directory_service_directory" "test-ad-connector" {
  directory_id = aws_directory_service_directory.connector.id
}
`, domain))
}
