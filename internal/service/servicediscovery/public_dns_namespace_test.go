package servicediscovery_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfservicediscovery "github.com/nij4t/terraform-provider-aws/internal/service/servicediscovery"
)

func TestAccServiceDiscoveryPublicDNSNamespace_basic(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPublicDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "servicediscovery", regexp.MustCompile(`namespace/.+`)),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "hosted_zone"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccServiceDiscoveryPublicDNSNamespace_disappears(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPublicDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfservicediscovery.ResourcePublicDNSNamespace(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccServiceDiscoveryPublicDNSNamespace_description(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPublicDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfigDescription(rName, "test"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
				),
			},
		},
	})
}

func TestAccServiceDiscoveryPublicDNSNamespace_tags(t *testing.T) {
	resourceName := "aws_service_discovery_public_dns_namespace.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheck(t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckPublicDNSNamespaceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(resourceName),
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
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccServiceDiscoveryPublicDnsNamespaceConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPublicDNSNamespaceExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckPublicDNSNamespaceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_public_dns_namespace" {
			continue
		}

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, servicediscovery.ErrCodeNamespaceNotFound, "") {
				return nil
			}
			return err
		}
	}
	return nil
}

func testAccCheckPublicDNSNamespaceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ServiceDiscoveryConn

		input := &servicediscovery.GetNamespaceInput{
			Id: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetNamespace(input)
		return err
	}
}

func testAccServiceDiscoveryPublicDnsNamespaceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.tf"
}
`, rName)
}

func testAccServiceDiscoveryPublicDnsNamespaceConfigDescription(rName, description string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  description = %[1]q
  name        = "%[2]s.tf"
}
`, description, rName)
}

func testAccServiceDiscoveryPublicDnsNamespaceConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.tf"

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccServiceDiscoveryPublicDnsNamespaceConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "test" {
  name = "%[1]s.tf"

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
