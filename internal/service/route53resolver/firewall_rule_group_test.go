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

func TestAccRoute53ResolverFirewallRuleGroup_basic(t *testing.T) {
	var v route53resolver.FirewallRuleGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
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

func TestAccRoute53ResolverFirewallRuleGroup_disappears(t *testing.T) {
	var v route53resolver.FirewallRuleGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupExists(resourceName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tfroute53resolver.ResourceFirewallRuleGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccRoute53ResolverFirewallRuleGroup_tags(t *testing.T) {
	var v route53resolver.FirewallRuleGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_route53_resolver_firewall_rule_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, route53resolver.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckRoute53ResolverFirewallRuleGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53ResolverFirewallRuleGroupConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
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
				Config: testAccRoute53ResolverFirewallRuleGroupConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoute53ResolverFirewallRuleGroupConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53ResolverFirewallRuleGroupExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "share_status", "NOT_SHARED"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckRoute53ResolverFirewallRuleGroupDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_resolver_firewall_rule_group" {
			continue
		}

		// Try to find the resource
		_, err := tfroute53resolver.FindFirewallRuleGroupByID(conn, rs.Primary.ID)
		// Verify the error is what we want
		if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("Route 53 Resolver DNS Firewall rule group still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckRoute53ResolverFirewallRuleGroupExists(n string, v *route53resolver.FirewallRuleGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Route 53 Resolver DNS Firewall rule group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).Route53ResolverConn
		out, err := tfroute53resolver.FindFirewallRuleGroupByID(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccRoute53ResolverFirewallRuleGroupConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccRoute53ResolverFirewallRuleGroupConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccRoute53ResolverFirewallRuleGroupConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_resolver_firewall_rule_group" "test" {
  name = %[1]q
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
