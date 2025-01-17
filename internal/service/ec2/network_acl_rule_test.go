package ec2_test

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfec2 "github.com/nij4t/terraform-provider-aws/internal/service/ec2"
)

func TestAccEC2NetworkACLRule_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.baz"),
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.qux"),
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.wibble"),
				),
			},
			{
				ResourceName:      "aws_network_acl_rule.baz",
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc("aws_network_acl_rule.baz", "tcp"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_network_acl_rule.qux",
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc("aws_network_acl_rule.qux", "icmp"),
				ImportStateVerify: true,
			},
			{
				ResourceName:      "aws_network_acl_rule.wibble",
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc("aws_network_acl_rule.wibble", "icmp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_disappears(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.baz"),
					testAccCheckNetworkACLRuleDelete("aws_network_acl_rule.baz"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_Disappears_ingressEgressSameNumber(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIngressEgressSameNumberMissing,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.baz"),
					testAccCheckNetworkACLRuleDelete("aws_network_acl_rule.baz"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_Disappears_networkACL(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_network_acl.bar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLExists(resourceName, &networkAcl),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceNetworkACL(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_missingParam(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccNetworkACLRuleMissingParam,
				ExpectError: regexp.MustCompile("Either `cidr_block` or `ipv6_cidr_block` must be defined"),
			},
		},
	})
}

func TestAccEC2NetworkACLRule_ipv6(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIPv6Config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists("aws_network_acl_rule.baz"),
				),
			},
			{
				ResourceName:      "aws_network_acl_rule.baz",
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc("aws_network_acl_rule.baz", "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_ipv6ICMP(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIPv6ICMPConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetworkACLRuleExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "58"),
				ImportStateVerify: true,
			},
		},
	})
}

// Reference: https://github.com/nij4t/terraform-provider-aws/issues/6710
func TestAccEC2NetworkACLRule_ipv6VPCAssignGeneratedIPv6CIDRBlockUpdate(t *testing.T) {
	var vpc ec2.Vpc
	vpcResourceName := "aws_vpc.test"
	resourceName := "aws_network_acl_rule.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkACLRuleIPv6VPCNotAssignGeneratedIpv6CIDRBlockUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(vpcResourceName, &vpc),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", "false"),
					resource.TestCheckResourceAttr(vpcResourceName, "ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccNetworkACLRuleIPv6VPCAssignGeneratedIpv6CIDRBlockUpdateConfig(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckVPCExists(vpcResourceName, &vpc),
					resource.TestCheckResourceAttr(vpcResourceName, "assign_generated_ipv6_cidr_block", "true"),
					resource.TestMatchResourceAttr(vpcResourceName, "ipv6_cidr_block", regexp.MustCompile(`/56$`)),
					testAccCheckNetworkACLRuleExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccNetworkACLRuleImportStateIdFunc(resourceName, "tcp"),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_allProtocol(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccNetworkACLRuleAllProtocolConfig,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testAccNetworkACLRuleAllProtocolNoRealUpdateConfig,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccEC2NetworkACLRule_tcpProtocol(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckNetworkACLRuleDestroy,
		Steps: []resource.TestStep{
			{
				Config:             testAccNetworkACLRuleTCPProtocolConfig,
				ExpectNonEmptyPlan: false,
			},
			{
				Config:             testAccNetworkACLRuleTCPProtocolNoRealUpdateConfig,
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestNetworkACLRule_validateICMPArgumentValue(t *testing.T) {
	type testCases struct {
		Value    string
		ErrCount int
	}

	invalidCases := []testCases{
		{
			Value:    "",
			ErrCount: 1,
		},
		{
			Value:    "not-a-number",
			ErrCount: 1,
		},
		{
			Value:    "1.0",
			ErrCount: 1,
		},
	}

	for _, tc := range invalidCases {
		_, errors := tfec2.ValidICMPArgumentValue(tc.Value, "icmp_type")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q to trigger a validation error.", tc.Value)
		}
	}

	validCases := []testCases{
		{
			Value:    "0",
			ErrCount: 0,
		},
		{
			Value:    "-1",
			ErrCount: 0,
		},
		{
			Value:    "1",
			ErrCount: 0,
		},
	}

	for _, tc := range validCases {
		_, errors := tfec2.ValidICMPArgumentValue(tc.Value, "icmp_type")
		if len(errors) != tc.ErrCount {
			t.Fatalf("Expected %q not to trigger a validation error.", tc.Value)
		}
	}

}

func testAccCheckNetworkACLRuleDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		if rs.Type != "aws_network_acl_rule" {
			continue
		}

		req := &ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeNetworkAcls(req)
		if err == nil {
			if len(resp.NetworkAcls) > 0 && *resp.NetworkAcls[0].NetworkAclId == rs.Primary.ID {
				networkAcl := resp.NetworkAcls[0]
				if networkAcl.Entries != nil {
					return fmt.Errorf("Network ACL Entries still exist")
				}
			}
		}

		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidNetworkAclID.NotFound" {
			return err
		}
	}

	return nil
}

func testAccCheckNetworkACLRuleExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ACL Rule Id is set")
		}

		req := &ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.Attributes["network_acl_id"])},
		}
		resp, err := conn.DescribeNetworkAcls(req)
		if err != nil {
			return err
		}
		if len(resp.NetworkAcls) != 1 {
			return fmt.Errorf("Network ACL not found")
		}
		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])
		if err != nil {
			return err
		}
		ruleNo, err := strconv.ParseInt(rs.Primary.Attributes["rule_number"], 10, 64)
		if err != nil {
			return err
		}
		for _, e := range resp.NetworkAcls[0].Entries {
			if *e.RuleNumber == ruleNo && *e.Egress == egress {
				return nil
			}
		}
		return fmt.Errorf("Entry not found: %s", resp.NetworkAcls[0])
	}
}

func testAccCheckNetworkACLRuleDelete(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ACL Rule Id is set")
		}

		egress, err := strconv.ParseBool(rs.Primary.Attributes["egress"])
		if err != nil {
			return err
		}
		ruleNo, err := strconv.ParseInt(rs.Primary.Attributes["rule_number"], 10, 64)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn
		_, err = conn.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
			NetworkAclId: aws.String(rs.Primary.Attributes["network_acl_id"]),
			RuleNumber:   aws.Int64(ruleNo),
			Egress:       aws.Bool(egress),
		})
		if err != nil {
			return fmt.Errorf("Error deleting Network ACL Rule (%s) in testAccCheckNetworkACLRuleDelete: %s", rs.Primary.ID, err)
		}

		return nil
	}
}

const testAccNetworkACLRuleBasicConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-basic"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-basic"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "qux" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 300
  protocol       = "icmp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  icmp_type      = 0
  icmp_code      = -1
}

resource "aws_network_acl_rule" "wibble" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 400
  protocol       = "icmp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  icmp_type      = -1
  icmp_code      = -1
}
`

const testAccNetworkACLRuleMissingParam = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-missing-param"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-missing-param"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 200
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleAllProtocolNoRealUpdateConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-all-proto-no-real-upd"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-no-real-update"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "all"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleTCPProtocolNoRealUpdateConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "testAccNetworkACLRuleTCPProtocolNoRealUpdateConfig"
  }
}
resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}
resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleAllProtocolConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-proto"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-all-protocol"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "-1"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleTCPProtocolConfig = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "testAccNetworkACLRuleTCPProtocolConfig"
  }
}
resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id
}
resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 150
  egress         = false
  protocol       = "6"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

const testAccNetworkACLRuleIPv6Config = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-ipv6"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-ipv6"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id  = aws_network_acl.bar.id
  rule_number     = 150
  egress          = false
  protocol        = "tcp"
  rule_action     = "allow"
  ipv6_cidr_block = "::/0"
  from_port       = 22
  to_port         = 22
}
`

const testAccNetworkACLRuleIngressEgressSameNumberMissing = `
resource "aws_vpc" "foo" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = "terraform-testacc-network-acl-rule-ingress-egress-same-number-missing"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-acl-rule-basic"
  }
}

resource "aws_network_acl_rule" "baz" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 100
  egress         = false
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}

resource "aws_network_acl_rule" "qux" {
  network_acl_id = aws_network_acl.bar.id
  rule_number    = 100
  egress         = true
  protocol       = "tcp"
  rule_action    = "allow"
  cidr_block     = "0.0.0.0/0"
  from_port      = 22
  to_port        = 22
}
`

func testAccNetworkACLRuleIPv6ICMPConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.3.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %q
  }
}

resource "aws_network_acl_rule" "test" {
  icmp_code       = -1
  icmp_type       = -1
  ipv6_cidr_block = "::/0"
  network_acl_id  = aws_network_acl.test.id
  protocol        = 58
  rule_action     = "allow"
  rule_number     = 150
}
`, rName, rName)
}

func testAccNetworkACLRuleIPv6VPCAssignGeneratedIpv6CIDRBlockUpdateConfig() string {
	return `
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-enabled"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-enabled"
  }
}

resource "aws_network_acl_rule" "test" {
  from_port       = 22
  ipv6_cidr_block = aws_vpc.test.ipv6_cidr_block
  network_acl_id  = aws_network_acl.test.id
  protocol        = "tcp"
  rule_action     = "allow"
  rule_number     = 150
  to_port         = 22
}
`
}

func testAccNetworkACLRuleIPv6VPCNotAssignGeneratedIpv6CIDRBlockUpdateConfig() string {
	return `
resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = false
  cidr_block                       = "10.3.0.0/16"

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-not-enabled"
  }
}

resource "aws_network_acl" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-network-acl-rule-ipv6-not-enabled"
  }
}`
}

func testAccNetworkACLRuleImportStateIdFunc(resourceName, resourceProtocol string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}
		networkAclId := rs.Primary.Attributes["network_acl_id"]
		ruleNumber := rs.Primary.Attributes["rule_number"]
		protocol := rs.Primary.Attributes["protocol"]
		// Ensure the resource's ID will be determined from the original protocol value set in the resource's config
		if protocol != resourceProtocol {
			protocol = resourceProtocol
		}
		egress := rs.Primary.Attributes["egress"]
		return fmt.Sprintf("%s:%s:%s:%s", networkAclId, ruleNumber, protocol, egress), nil
	}
}
