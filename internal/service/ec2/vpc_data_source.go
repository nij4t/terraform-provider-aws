package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
)

func DataSourceVPC() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVPCRead,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"cidr_block_associations": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"association_id": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"state": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},

			"default": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},

			"dhcp_options_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"enable_dns_hostnames": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"enable_dns_support": {
				Type:     schema.TypeBool,
				Computed: true,
			},

			"filter": CustomFiltersSchema(),

			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"instance_tenancy": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv6_cidr_block": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"ipv6_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"main_route_table_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"state": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"tags": tftags.TagsSchemaComputed(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceVPCRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	req := &ec2.DescribeVpcsInput{}

	var id string
	if cid, ok := d.GetOk("id"); ok {
		id = cid.(string)
	}

	if id != "" {
		req.VpcIds = []*string{aws.String(id)}
	}

	// We specify "default" as boolean, but EC2 filters want
	// it to be serialized as a string. Note that setting it to
	// "false" here does not actually filter by it *not* being
	// the default, because Terraform can't distinguish between
	// "false" and "not set".
	isDefaultStr := ""
	if d.Get("default").(bool) {
		isDefaultStr = "true"
	}

	req.Filters = BuildAttributeFilterList(
		map[string]string{
			"cidr":            d.Get("cidr_block").(string),
			"dhcp-options-id": d.Get("dhcp_options_id").(string),
			"isDefault":       isDefaultStr,
			"state":           d.Get("state").(string),
		},
	)

	if tags, tagsOk := d.GetOk("tags"); tagsOk {
		req.Filters = append(req.Filters, BuildTagFilterList(
			Tags(tftags.New(tags.(map[string]interface{}))),
		)...)
	}

	req.Filters = append(req.Filters, BuildCustomFilterList(
		d.Get("filter").(*schema.Set),
	)...)
	if len(req.Filters) == 0 {
		// Don't send an empty filters list; the EC2 API won't accept it.
		req.Filters = nil
	}

	log.Printf("[DEBUG] Reading AWS VPC: %s", req)
	resp, err := conn.DescribeVpcs(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.Vpcs) == 0 {
		return fmt.Errorf("no matching VPC found")
	}
	if len(resp.Vpcs) > 1 {
		return fmt.Errorf("multiple VPCs matched; use additional constraints to reduce matches to a single VPC")
	}

	vpc := resp.Vpcs[0]

	d.SetId(aws.StringValue(vpc.VpcId))
	d.Set("cidr_block", vpc.CidrBlock)
	d.Set("dhcp_options_id", vpc.DhcpOptionsId)
	d.Set("instance_tenancy", vpc.InstanceTenancy)
	d.Set("default", vpc.IsDefault)
	d.Set("state", vpc.State)

	if err := d.Set("tags", KeyValueTags(vpc.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	d.Set("owner_id", vpc.OwnerId)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: aws.StringValue(vpc.OwnerId),
		Resource:  fmt.Sprintf("vpc/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	cidrAssociations := []interface{}{}
	for _, associationSet := range vpc.CidrBlockAssociationSet {
		association := map[string]interface{}{
			"association_id": aws.StringValue(associationSet.AssociationId),
			"cidr_block":     aws.StringValue(associationSet.CidrBlock),
			"state":          aws.StringValue(associationSet.CidrBlockState.State),
		}
		cidrAssociations = append(cidrAssociations, association)
	}
	if err := d.Set("cidr_block_associations", cidrAssociations); err != nil {
		return fmt.Errorf("error setting cidr_block_associations: %w", err)
	}

	if vpc.Ipv6CidrBlockAssociationSet != nil {
		d.Set("ipv6_association_id", vpc.Ipv6CidrBlockAssociationSet[0].AssociationId)
		d.Set("ipv6_cidr_block", vpc.Ipv6CidrBlockAssociationSet[0].Ipv6CidrBlock)
	}

	enableDnsHostnames, err := FindVPCAttribute(conn, aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsHostnames)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) Attribute (%s): %w", aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsHostnames, err)
	}

	d.Set("enable_dns_hostnames", enableDnsHostnames)

	enableDnsSupport, err := FindVPCAttribute(conn, aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsSupport)

	if err != nil {
		return fmt.Errorf("error reading EC2 VPC (%s) Attribute (%s): %w", aws.StringValue(vpc.VpcId), ec2.VpcAttributeNameEnableDnsSupport, err)
	}

	d.Set("enable_dns_support", enableDnsSupport)

	routeTableId, err := resourceVPCSetMainRouteTable(conn, aws.StringValue(vpc.VpcId))
	if err != nil {
		log.Printf("[WARN] Unable to set Main Route Table: %s", err)
	}
	d.Set("main_route_table_id", routeTableId)

	return nil
}
