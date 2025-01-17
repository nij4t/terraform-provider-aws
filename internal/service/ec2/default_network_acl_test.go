package ec2_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

var defaultEgressAcl = &ec2.NetworkAclEntry{
	CidrBlock:  aws.String("0.0.0.0/0"),
	Egress:     aws.Bool(true),
	Protocol:   aws.String("-1"),
	RuleAction: aws.String("allow"),
	RuleNumber: aws.Int64(100),
}
var ipv6IngressAcl = &ec2.NetworkAclEntry{
	Ipv6CidrBlock: aws.String("::/0"),
	Egress:        aws.Bool(false),
	Protocol:      aws.String("-1"),
	RuleAction:    aws.String("allow"),
	RuleNumber:    aws.Int64(101),
}

func TestAccEC2DefaultNetworkACL_basic(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultNetworkConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`network-acl/acl-.+`)),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{}, 0, 2),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
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

func TestAccEC2DefaultNetworkACL_basicIPv6VPC(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultNetworkConfig_basicIPv6VPC,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{}, 0, 4),
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

func TestAccEC2DefaultNetworkACL_Deny_ingress(t *testing.T) {
	// TestAccEC2DefaultNetworkACL_Deny_ingress will deny all Ingress rules, but
	// not Egress. We then expect there to be 3 rules, 2 AWS defaults and 1
	// additional Egress.
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultNetworkConfig_deny_ingress,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{defaultEgressAcl}, 0, 2),
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

func TestAccEC2DefaultNetworkACL_withIPv6Ingress(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultNetworkConfig_includingIPv6Rule,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{ipv6IngressAcl}, 0, 2),
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

func TestAccEC2DefaultNetworkACL_subnetRemoval(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultNetworkConfig_Subnets,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{}, 2, 2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Here the Subnets have been removed from the Default Network ACL Config,
			// but have not been reassigned. The result is that the Subnets are still
			// there, and we have a non-empty plan
			{
				Config: testAccDefaultNetworkConfig_Subnets_remove,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{}, 2, 2),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2DefaultNetworkACL_subnetReassign(t *testing.T) {
	var networkAcl ec2.NetworkAcl
	resourceName := "aws_default_network_acl.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDefaultNetworkACLDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultNetworkConfig_Subnets,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{}, 2, 2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},

			// Here we've reassigned the subnets to a different ACL.
			// Without any otherwise association between the `aws_network_acl` and
			// `aws_default_network_acl` resources, we cannot guarantee that the
			// reassignment of the two subnets to the `aws_network_acl` will happen
			// before the update/read on the `aws_default_network_acl` resource.
			// Because of this, there could be a non-empty plan if a READ is done on
			// the default before the reassignment occurs on the other resource.
			//
			// For the sake of testing, here we introduce a depends_on attribute from
			// the default resource to the other acl resource, to ensure the latter's
			// update occurs first, and the former's READ will correctly read zero
			// subnets
			{
				Config: testAccDefaultNetworkConfig_Subnets_move,
				Check: resource.ComposeTestCheckFunc(
					testAccGetDefaultNetworkACL(resourceName, &networkAcl),
					testAccCheckDefaultACLAttributes(&networkAcl, []*ec2.NetworkAclEntry{}, 0, 2),
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

func testAccCheckDefaultNetworkACLDestroy(s *terraform.State) error {
	// We can't destroy this resource; it comes and goes with the VPC itself.
	return nil
}

func testAccCheckDefaultACLAttributes(acl *ec2.NetworkAcl, rules []*ec2.NetworkAclEntry, subnetCount int, hiddenRuleCount int) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		aclEntriesCount := len(acl.Entries)
		ruleCount := len(rules)

		// Default ACL has hidden rules we can't do anything about
		ruleCount = ruleCount + hiddenRuleCount

		if ruleCount != aclEntriesCount {
			return fmt.Errorf("Expected (%d) Rules, got (%d)", ruleCount, aclEntriesCount)
		}

		if len(acl.Associations) != subnetCount {
			return fmt.Errorf("Expected (%d) Subnets, got (%d)", subnetCount, len(acl.Associations))
		}

		return nil
	}
}

func testAccGetDefaultNetworkACL(n string, networkAcl *ec2.NetworkAcl) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Network ACL is set")
		}
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		resp, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
			NetworkAclIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.NetworkAcls) > 0 &&
			aws.StringValue(resp.NetworkAcls[0].NetworkAclId) == rs.Primary.ID {
			*networkAcl = *resp.NetworkAcls[0]
			return nil
		}

		return fmt.Errorf("Network Acls not found")
	}
}

const testAccDefaultNetworkConfig_basic = `
resource "aws_vpc" "tftestvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-basic"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.tftestvpc.default_network_acl_id

  tags = {
    Name = "tf-acc-default-acl-basic"
  }
}
`

const testAccDefaultNetworkConfig_includingIPv6Rule = `
resource "aws_vpc" "tftestvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-including-ipv6-rule"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.tftestvpc.default_network_acl_id

  ingress {
    protocol        = -1
    rule_no         = 101
    action          = "allow"
    ipv6_cidr_block = "::/0"
    from_port       = 0
    to_port         = 0
  }

  tags = {
    Name = "tf-acc-default-acl-basic-including-ipv6-rule"
  }
}
`

const testAccDefaultNetworkConfig_deny_ingress = `
resource "aws_vpc" "tftestvpc" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-deny-ingress"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.tftestvpc.default_network_acl_id

  egress {
    protocol   = -1
    rule_no    = 100
    action     = "allow"
    cidr_block = "0.0.0.0/0"
    from_port  = 0
    to_port    = 0
  }

  tags = {
    Name = "tf-acc-default-acl-deny-ingress"
  }
}
`

const testAccDefaultNetworkConfig_Subnets = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-subnets"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-network-acl-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-network-acl-two"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-acl-subnets"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.foo.default_network_acl_id

  subnet_ids = [aws_subnet.one.id, aws_subnet.two.id]

  tags = {
    Name = "tf-acc-default-acl-subnets"
  }
}
`

const testAccDefaultNetworkConfig_Subnets_remove = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-subnets-remove"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-network-acl-subnets-remove-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-network-acl-subnets-remove-two"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-acl-subnets-remove"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.foo.default_network_acl_id

  tags = {
    Name = "tf-acc-default-acl-subnets-remove"
  }
}
`

const testAccDefaultNetworkConfig_Subnets_move = `
resource "aws_vpc" "foo" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-default-network-acl-subnets-move"
  }
}

resource "aws_subnet" "one" {
  cidr_block = "10.1.111.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-network-acl-subnets-move-one"
  }
}

resource "aws_subnet" "two" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.foo.id

  tags = {
    Name = "tf-acc-default-network-acl-subnets-move-two"
  }
}

resource "aws_network_acl" "bar" {
  vpc_id = aws_vpc.foo.id

  subnet_ids = [aws_subnet.one.id, aws_subnet.two.id]

  tags = {
    Name = "tf-acc-default-acl-subnets-move"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.foo.default_network_acl_id

  depends_on = [aws_network_acl.bar]

  tags = {
    Name = "tf-acc-default-acl-subnets-move"
  }
}
`

const testAccDefaultNetworkConfig_basicIPv6VPC = `
resource "aws_vpc" "tftestvpc" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = "terraform-testacc-default-network-acl-basic-ipv6-vpc"
  }
}

resource "aws_default_network_acl" "default" {
  default_network_acl_id = aws_vpc.tftestvpc.default_network_acl_id

  tags = {
    Name = "tf-acc-default-acl-subnets-basic-ipv6-vpc"
  }
}
`
