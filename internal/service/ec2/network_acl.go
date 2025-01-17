package ec2

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/create"
	"github.com/nij4t/terraform-provider-aws/internal/flex"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceNetworkACL() *schema.Resource {

	return &schema.Resource{
		Create: resourceNetworkACLCreate,
		Read:   resourceNetworkACLRead,
		Delete: resourceNetworkACLDelete,
		Update: resourceNetworkACLUpdate,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"subnet_ids": {
				Type:     schema.TypeSet,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"ingress": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
						"rule_no": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 32766),
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							ValidateFunc: validation.StringInSlice([]string{
								ec2.RuleActionAllow,
								ec2.RuleActionDeny,
							}, true),
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validation.IsCIDR,
							),
						},
						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validation.IsCIDR,
							),
						},
						"icmp_type": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"icmp_code": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Set: resourceNetworkACLEntryHash,
			},
			"egress": {
				Type:       schema.TypeSet,
				Optional:   true,
				Computed:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"from_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
						"to_port": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IsPortNumberOrZero,
						},
						"rule_no": {
							Type:         schema.TypeInt,
							Required:     true,
							ValidateFunc: validation.IntBetween(1, 32766),
						},
						"action": {
							Type:     schema.TypeString,
							Required: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								return strings.EqualFold(old, new)
							},
							ValidateFunc: validation.StringInSlice([]string{
								ec2.RuleActionAllow,
								ec2.RuleActionDeny,
							}, true),
						},
						"protocol": {
							Type:     schema.TypeString,
							Required: true,
						},
						"cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validation.IsCIDR,
							),
						},
						"ipv6_cidr_block": {
							Type:     schema.TypeString,
							Optional: true,
							ValidateFunc: validation.Any(
								validation.StringIsEmpty,
								validation.IsCIDR,
							),
						},
						"icmp_type": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"icmp_code": {
							Type:     schema.TypeInt,
							Optional: true,
						},
					},
				},
				Set: resourceNetworkACLEntryHash,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceNetworkACLCreate(d *schema.ResourceData, meta interface{}) error {

	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	// Create the Network Acl
	createOpts := &ec2.CreateNetworkAclInput{
		VpcId:             aws.String(d.Get("vpc_id").(string)),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeNetworkAcl),
	}

	log.Printf("[DEBUG] Network Acl create config: %#v", createOpts)
	resp, err := conn.CreateNetworkAcl(createOpts)

	if err != nil {
		return fmt.Errorf("error creating EC2 Network ACL: %w", err)
	}

	if resp == nil || resp.NetworkAcl == nil {
		return fmt.Errorf("error creating EC2 Network ACL: empty response")
	}

	d.SetId(aws.StringValue(resp.NetworkAcl.NetworkAclId))

	if v, ok := d.GetOk("egress"); ok && v.(*schema.Set).Len() > 0 {
		err := updateNetworkAclEntries(d, "egress", conn)

		if err != nil {
			return fmt.Errorf("error updating EC2 Network ACL (%s) Egress Entries: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("ingress"); ok && v.(*schema.Set).Len() > 0 {
		err := updateNetworkAclEntries(d, "ingress", conn)

		if err != nil {
			return fmt.Errorf("error updating EC2 Network ACL (%s) Ingress Entries: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("subnet_ids"); ok && v.(*schema.Set).Len() > 0 {
		for _, subnetIDRaw := range v.(*schema.Set).List() {
			subnetID, ok := subnetIDRaw.(string)

			if !ok {
				continue
			}

			association, err := findNetworkAclAssociation(subnetID, conn)

			if err != nil {
				return fmt.Errorf("error finding existing EC2 Network ACL association for Subnet (%s): %w", subnetID, err)
			}

			if association == nil {
				return fmt.Errorf("error finding existing EC2 Network ACL association for Subnet (%s): empty response", subnetID)
			}

			input := &ec2.ReplaceNetworkAclAssociationInput{
				AssociationId: association.NetworkAclAssociationId,
				NetworkAclId:  aws.String(d.Id()),
			}

			_, err = conn.ReplaceNetworkAclAssociation(input)

			if err != nil {
				return fmt.Errorf("error replacing existing EC2 Network ACL association for Subnet (%s): %w", subnetID, err)
			}
		}
	}

	return resourceNetworkACLRead(d, meta)
}

func resourceNetworkACLRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var networkAcl *ec2.NetworkAcl

	err := resource.Retry(NetworkACLPropagationTimeout, func() *resource.RetryError {
		var err error

		networkAcl, err = FindNetworkACLByID(conn, d.Id())

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidNetworkAclID.NotFound") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && networkAcl == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("EC2 Network ACL (%s) not found", d.Id()),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		networkAcl, err = FindNetworkACLByID(conn, d.Id())
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidNetworkAclID.NotFound") {
		log.Printf("[WARN] EC2 Network ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Network ACL (%s): %w", d.Id(), err)
	}

	if networkAcl == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 Network ACL (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] EC2 Network ACL (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var ingressEntries []*ec2.NetworkAclEntry
	var egressEntries []*ec2.NetworkAclEntry

	// separate the ingress and egress rules
	for _, e := range networkAcl.Entries {
		// Skip the default rules added by AWS. They can be neither
		// configured or deleted by users.
		if aws.Int64Value(e.RuleNumber) == defaultACLRuleNumberIPv4 ||
			aws.Int64Value(e.RuleNumber) == defaultACLRuleNumberIPv6 {
			continue
		}

		if aws.BoolValue(e.Egress) {
			egressEntries = append(egressEntries, e)
		} else {
			ingressEntries = append(ingressEntries, e)
		}
	}

	d.Set("vpc_id", networkAcl.VpcId)

	tags := KeyValueTags(networkAcl.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("owner_id", networkAcl.OwnerId)

	var subnetIds []*string
	for _, a := range networkAcl.Associations {
		subnetIds = append(subnetIds, a.SubnetId)
	}
	if err := d.Set("subnet_ids", flex.FlattenStringSet(subnetIds)); err != nil {
		return err
	}

	if err := d.Set("ingress", networkAclEntriesToMapList(ingressEntries)); err != nil {
		return err
	}
	if err := d.Set("egress", networkAclEntriesToMapList(egressEntries)); err != nil {
		return err
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(networkAcl.OwnerId),
		Resource:  fmt.Sprintf("network-acl/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceNetworkACLUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("ingress") {
		err := updateNetworkAclEntries(d, "ingress", conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("egress") {
		err := updateNetworkAclEntries(d, "egress", conn)
		if err != nil {
			return err
		}
	}

	if d.HasChange("subnet_ids") {
		o, n := d.GetChange("subnet_ids")
		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		remove := os.Difference(ns).List()
		add := ns.Difference(os).List()

		if len(remove) > 0 {
			// A Network ACL is required for each subnet. In order to disassociate a
			// subnet from this ACL, we must associate it with the default ACL.
			defaultAcl, err := GetDefaultNetworkACL(d.Get("vpc_id").(string), conn)
			if err != nil {
				return fmt.Errorf("Failed to find Default ACL for VPC %s", d.Get("vpc_id").(string))
			}
			for _, r := range remove {
				association, err := findNetworkAclAssociation(r.(string), conn)
				if err != nil {
					if tfresource.NotFound(err) {
						// Subnet has been deleted.
						continue
					}
					return fmt.Errorf("Failed to find acl association: acl %s with subnet %s: %s", d.Id(), r, err)
				}
				log.Printf("[DEBUG] Replacing Network Acl Association (%s) with Default Network ACL ID (%s)", *association.NetworkAclAssociationId, *defaultAcl.NetworkAclId)
				_, err = conn.ReplaceNetworkAclAssociation(&ec2.ReplaceNetworkAclAssociationInput{
					AssociationId: association.NetworkAclAssociationId,
					NetworkAclId:  defaultAcl.NetworkAclId,
				})
				if err != nil {
					if tfawserr.ErrMessageContains(err, "InvalidAssociationID.NotFound", "") {
						continue
					}
					return fmt.Errorf("Error Replacing Default Network Acl Association: %s", err)
				}
			}
		}

		if len(add) > 0 {
			for _, a := range add {
				association, err := findNetworkAclAssociation(a.(string), conn)
				if err != nil {
					return fmt.Errorf("Failed to find acl association: acl %s with subnet %s: %s", d.Id(), a, err)
				}
				_, err = conn.ReplaceNetworkAclAssociation(&ec2.ReplaceNetworkAclAssociationInput{
					AssociationId: association.NetworkAclAssociationId,
					NetworkAclId:  aws.String(d.Id()),
				})
				if err != nil {
					return err
				}
			}
		}

	}

	if d.HasChange("tags_all") && !d.IsNewResource() {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Network ACL (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceNetworkACLRead(d, meta)
}

func updateNetworkAclEntries(d *schema.ResourceData, entryType string, conn *ec2.EC2) error {
	if d.HasChange(entryType) {
		o, n := d.GetChange(entryType)

		if o == nil {
			o = new(schema.Set)
		}
		if n == nil {
			n = new(schema.Set)
		}

		os := o.(*schema.Set)
		ns := n.(*schema.Set)

		toBeDeleted, err := ExpandNetworkACLEntries(os.Difference(ns).List(), entryType)
		if err != nil {
			return err
		}
		for _, remove := range toBeDeleted {
			// AWS includes default rules with all network ACLs that can be
			// neither modified nor destroyed. They have a custom rule
			// number that is out of bounds for any other rule. If we
			// encounter it, just continue. There's no work to be done.
			if aws.Int64Value(remove.RuleNumber) == defaultACLRuleNumberIPv4 ||
				aws.Int64Value(remove.RuleNumber) == defaultACLRuleNumberIPv6 {
				continue
			}

			// Delete old Acl
			log.Printf("[DEBUG] Destroying Network ACL Entry number (%d)", int(aws.Int64Value(remove.RuleNumber)))
			_, err := conn.DeleteNetworkAclEntry(&ec2.DeleteNetworkAclEntryInput{
				NetworkAclId: aws.String(d.Id()),
				RuleNumber:   remove.RuleNumber,
				Egress:       remove.Egress,
			})
			if err != nil {
				return fmt.Errorf("Error deleting %s entry: %s", entryType, err)
			}
		}

		toBeCreated, err := ExpandNetworkACLEntries(ns.Difference(os).List(), entryType)
		if err != nil {
			return err
		}
		for _, add := range toBeCreated {
			// Protocol -1 rules don't store ports in AWS. Thus, they'll always
			// hash differently when being read out of the API. Force the user
			// to set from_port and to_port to 0 for these rules, to keep the
			// hashing consistent.
			if aws.StringValue(add.Protocol) == "-1" {
				to := aws.Int64Value(add.PortRange.To)
				from := aws.Int64Value(add.PortRange.From)
				expected := &ExpectedPortPair{
					to_port:   0,
					from_port: 0,
				}
				if ok := ValidPorts(to, from, *expected); !ok {
					return fmt.Errorf(
						"to_port (%d) and from_port (%d) must both be 0 to use the the 'all' \"-1\" protocol!",
						to, from)
				}
			}

			// AWS mutates the CIDR block into a network implied by the IP and
			// mask provided. This results in hashing inconsistencies between
			// the local config file and the state returned by the API. Error
			// if the user provides a CIDR block with an inappropriate mask
			if cidrBlock := aws.StringValue(add.CidrBlock); cidrBlock != "" {
				if err := verify.ValidateIPv4CIDRBlock(cidrBlock); err != nil {
					return err
				}
			}
			if ipv6CidrBlock := aws.StringValue(add.Ipv6CidrBlock); ipv6CidrBlock != "" {
				if err := verify.ValidateIPv6CIDRBlock(ipv6CidrBlock); err != nil {
					return err
				}
			}

			createOpts := &ec2.CreateNetworkAclEntryInput{
				NetworkAclId: aws.String(d.Id()),
				Egress:       add.Egress,
				PortRange:    add.PortRange,
				Protocol:     add.Protocol,
				RuleAction:   add.RuleAction,
				RuleNumber:   add.RuleNumber,
				IcmpTypeCode: add.IcmpTypeCode,
			}

			if add.CidrBlock != nil && aws.StringValue(add.CidrBlock) != "" {
				createOpts.CidrBlock = add.CidrBlock
			}

			if add.Ipv6CidrBlock != nil && aws.StringValue(add.Ipv6CidrBlock) != "" {
				createOpts.Ipv6CidrBlock = add.Ipv6CidrBlock
			}

			// Add new Acl entry
			_, connErr := conn.CreateNetworkAclEntry(createOpts)
			if connErr != nil {
				return fmt.Errorf("Error creating %s entry: %s", entryType, connErr)
			}
		}
	}
	return nil
}

func resourceNetworkACLDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting Network Acl: %s", d.Id())
	input := &ec2.DeleteNetworkAclInput{
		NetworkAclId: aws.String(d.Id()),
	}
	err := resource.Retry(5*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteNetworkAcl(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidNetworkAclID.NotFound", "") {
				return nil
			}
			if tfawserr.ErrMessageContains(err, "DependencyViolation", "") {
				err = cleanUpDependencyViolations(d, conn)
				if err != nil {
					return resource.NonRetryableError(err)
				}
				return resource.RetryableError(fmt.Errorf("Dependencies found and cleaned up, retrying"))
			}

			return resource.NonRetryableError(err)

		}
		log.Printf("[Info] Deleted network ACL %s successfully", d.Id())
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.DeleteNetworkAcl(input)
		if err != nil && tfawserr.ErrMessageContains(err, "InvalidNetworkAclID.NotFound", "") {
			return nil
		}
		err = cleanUpDependencyViolations(d, conn)
		if err != nil {
			// This seems excessive but is probably the best way to make sure it's actually deleted
			_, err = conn.DeleteNetworkAcl(input)
			if err != nil && tfawserr.ErrMessageContains(err, "InvalidNetworkAclID.NotFound", "") {
				return nil
			}
		}
	}
	if err != nil {
		return fmt.Errorf("Error destroying Network ACL (%s): %s", d.Id(), err)
	}
	return nil
}

func cleanUpDependencyViolations(d *schema.ResourceData, conn *ec2.EC2) error {
	// In case of dependency violation, we remove the association between subnet and network acl.
	// This means the subnet is attached to default acl of vpc.
	var associations []*ec2.NetworkAclAssociation
	if v, ok := d.GetOk("subnet_ids"); ok {
		ids := v.(*schema.Set).List()
		for _, i := range ids {
			a, err := findNetworkAclAssociation(i.(string), conn)
			if err != nil {
				if tfresource.NotFound(err) {
					continue
				}
				return fmt.Errorf("Error finding network ACL association: %s", err)
			}
			associations = append(associations, a)
		}
	}

	log.Printf("[DEBUG] Replacing network associations for Network ACL (%s): %s", d.Id(), associations)
	defaultAcl, err := GetDefaultNetworkACL(d.Get("vpc_id").(string), conn)
	if err != nil {
		return fmt.Errorf("Error getting default network ACL: %s", err)
	}

	for _, a := range associations {
		log.Printf("DEBUG] Replacing Network Acl Association (%s) with Default Network ACL ID (%s)",
			aws.StringValue(a.NetworkAclAssociationId), aws.StringValue(defaultAcl.NetworkAclId))
		_, replaceErr := conn.ReplaceNetworkAclAssociation(&ec2.ReplaceNetworkAclAssociationInput{
			AssociationId: a.NetworkAclAssociationId,
			NetworkAclId:  defaultAcl.NetworkAclId,
		})
		if replaceErr != nil {
			// It's possible that during an attempt to replace this
			// association, the Subnet in question has already been moved to
			// another ACL. This can happen if you're destroying a network acl
			// and simultaneously re-associating it's subnet(s) with another
			// ACL; Terraform may have already re-associated the subnet(s) by
			// the time we attempt to destroy them, even between the time we
			// list them and then try to destroy them. In this case, the
			// association we're trying to replace will no longer exist and
			// this call will fail. Here we trap that error and fail
			// gracefully; the association we tried to replace gone, we trust
			// someone else has taken ownership.
			if tfawserr.ErrMessageContains(replaceErr, "InvalidAssociationID.NotFound", "") {
				log.Printf("[WARN] Network Association (%s) no longer found; Network Association likely updated or removed externally, removing from state", aws.StringValue(a.NetworkAclAssociationId))
				continue
			}
			log.Printf("[ERR] Non retry-able error in replacing associations for Network ACL (%s): %s", d.Id(), replaceErr)
			return fmt.Errorf("Error replacing network ACL associations: %s", err)
		}
	}
	return nil
}

func resourceNetworkACLEntryHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%d-", m["from_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["to_port"].(int)))
	buf.WriteString(fmt.Sprintf("%d-", m["rule_no"].(int)))
	buf.WriteString(fmt.Sprintf("%s-", strings.ToLower(m["action"].(string))))

	// The AWS network ACL API only speaks protocol numbers, and that's
	// all we store. Never hash a protocol name.
	protocol := m["protocol"].(string)
	if _, err := strconv.Atoi(m["protocol"].(string)); err != nil {
		// We're a protocol name. Look up the number.
		buf.WriteString(fmt.Sprintf("%d-", protocolIntegers()[protocol]))
	} else {
		// We're a protocol number. Pass the value through.
		buf.WriteString(fmt.Sprintf("%s-", protocol))
	}

	if v, ok := m["cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["ipv6_cidr_block"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["ssl_certificate_id"]; ok {
		buf.WriteString(fmt.Sprintf("%s-", v.(string)))
	}

	if v, ok := m["icmp_type"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}
	if v, ok := m["icmp_code"]; ok {
		buf.WriteString(fmt.Sprintf("%d-", v.(int)))
	}

	return create.StringHashcode(buf.String())
}

func GetDefaultNetworkACL(vpcId string, conn *ec2.EC2) (defaultAcl *ec2.NetworkAcl, err error) {
	resp, err := conn.DescribeNetworkAcls(&ec2.DescribeNetworkAclsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("default"),
				Values: []*string{aws.String("true")},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{aws.String(vpcId)},
			},
		},
	})

	if err != nil {
		return nil, err
	}
	return resp.NetworkAcls[0], nil
}

func findNetworkAclAssociation(subnetId string, conn *ec2.EC2) (networkAclAssociation *ec2.NetworkAclAssociation, err error) {
	req := &ec2.DescribeNetworkAclsInput{}
	req.Filters = BuildAttributeFilterList(
		map[string]string{
			"association.subnet-id": subnetId,
		},
	)
	resp, err := conn.DescribeNetworkAcls(req)
	if err != nil {
		return nil, err
	}

	if len(resp.NetworkAcls) > 0 {
		for _, association := range resp.NetworkAcls[0].Associations {
			if aws.StringValue(association.SubnetId) == subnetId {
				return association, nil
			}
		}
	}

	return nil, &resource.NotFoundError{
		LastRequest:  req,
		LastResponse: resp,
		Message:      fmt.Sprintf("could not find association for subnet: %s ", subnetId),
	}
}

// networkAclEntriesToMapList turns ingress/egress rules read from AWS into a list
// of maps.
func networkAclEntriesToMapList(networkAcls []*ec2.NetworkAclEntry) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(networkAcls))
	for _, entry := range networkAcls {
		acl := make(map[string]interface{})
		acl["rule_no"] = aws.Int64Value(entry.RuleNumber)
		acl["action"] = aws.StringValue(entry.RuleAction)
		if entry.CidrBlock != nil {
			acl["cidr_block"] = aws.StringValue(entry.CidrBlock)
		}
		if entry.Ipv6CidrBlock != nil {
			acl["ipv6_cidr_block"] = aws.StringValue(entry.Ipv6CidrBlock)
		}
		// The AWS network ACL API only speaks protocol numbers, and
		// that's all we record.
		if _, err := strconv.Atoi(aws.StringValue(entry.Protocol)); err != nil {
			// We're a protocol name. Look up the number.
			acl["protocol"] = protocolIntegers()[aws.StringValue(entry.Protocol)]
		} else {
			// We're a protocol number. Pass through.
			acl["protocol"] = aws.StringValue(entry.Protocol)
		}

		acl["protocol"] = aws.StringValue(entry.Protocol)
		if entry.PortRange != nil {
			acl["from_port"] = aws.Int64Value(entry.PortRange.From)
			acl["to_port"] = aws.Int64Value(entry.PortRange.To)
		}

		if entry.IcmpTypeCode != nil {
			acl["icmp_type"] = aws.Int64Value(entry.IcmpTypeCode.Type)
			acl["icmp_code"] = aws.Int64Value(entry.IcmpTypeCode.Code)
		}

		result = append(result, acl)
	}

	return result
}
