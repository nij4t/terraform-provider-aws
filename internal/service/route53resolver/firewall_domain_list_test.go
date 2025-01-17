package route53resolver_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfroute53resolver "github.com/nij4t/terraform-provider-aws/internal/service/route53resolver"
)

func TestAccRoute53ResolverFirewallDomainList_basic(t *testing.T) {
	var v route53resolver.FirewallDomainList
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_domain_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallDomainListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallDomainListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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

func TestAccRoute53ResolverFirewallDomainList_domains(t *testing.T) {
	var v route53resolver.FirewallDomainList
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_domain_list.test"

	domainName1 := acctest.RandomFQDomainName()
	domainName2 := acctest.RandomFQDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallDomainListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallDomainListConfigDomains(rName, domainName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "domains.*", domainName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53ResolverFirewallDomainListConfigDomains(rName, domainName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "domains.*", domainName2),
				),
			},
			{
				Config: testAccRoute53ResolverFirewallDomainListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "domains.#", "0"),
				),
			},
		},
	})
}

func TestAccRoute53ResolverFirewallDomainList_disappears(t *testing.T) {
	var v route53resolver.FirewallDomainList
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_domain_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallDomainListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallDomainListConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceFirewallDomainList(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverFirewallDomainList_tags(t *testing.T) {
	var v route53resolver.FirewallDomainList
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_domain_list.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallDomainListDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallDomainListConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
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
				Config: testAccRoute53ResolverFirewallDomainListConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoute53ResolverFirewallDomainListConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallDomainListExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverFirewallDomainListDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_firewall_domain_list" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindFirewallDomainListByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver DNS Firewall domain list still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverFirewallDomainListExists(n string, v *route53resolver.FirewallDomainList) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver DNS Firewall domain list ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindFirewallDomainListByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccRoute53ResolverFirewallDomainListConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRoute53ResolverFirewallDomainListConfigDomains(rName, domain string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name    = %[1]q
  domains = [%[2]q]
}
`, rName, domain)
}

func testAccRoute53ResolverFirewallDomainListConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRoute53ResolverFirewallDomainListConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_domain_list" "test" {
  name = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
