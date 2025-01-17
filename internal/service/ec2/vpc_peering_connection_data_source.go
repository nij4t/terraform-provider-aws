package ec2

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
)

func DataSourceVPCPeeringConnection() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVPCPeeringConnectionRead,

		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_vpc_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_owner_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_cidr_block": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"peer_cidr_block_set": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"cidr_block": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
			"peer_region": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"accepter": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			"requester": {
				Type:     schema.TypeMap,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeBool},
			},
			"filter": CustomFiltersSchema(),
			"tags":   tftags.TagsSchemaComputed(),
		},
	}
}

func dataSourceVPCPeeringConnectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading VPC Peering Connections.")

	req := &ec2.DescribeVpcPeeringConnectionsInput{}

	if id, ok := d.GetOk("id"); ok {
		req.VpcPeeringConnectionIds = aws.StringSlice([]string{id.(string)})
	}

	req.Filters = BuildAttributeFilterList(
		map[string]string{
			"status-code":                   d.Get("status").(string),
			"requester-vpc-info.vpc-id":     d.Get("vpc_id").(string),
			"requester-vpc-info.owner-id":   d.Get("owner_id").(string),
			"requester-vpc-info.cidr-block": d.Get("cidr_block").(string),
			"accepter-vpc-info.vpc-id":      d.Get("peer_vpc_id").(string),
			"accepter-vpc-info.owner-id":    d.Get("peer_owner_id").(string),
			"accepter-vpc-info.cidr-block":  d.Get("peer_cidr_block").(string),
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

	log.Printf("[DEBUG] Reading VPC Peering Connection: %s", req)
	resp, err := conn.DescribeVpcPeeringConnections(req)
	if err != nil {
		return err
	}
	if resp == nil || len(resp.VpcPeeringConnections) == 0 {
		return fmt.Errorf("no matching VPC peering connection found")
	}
	if len(resp.VpcPeeringConnections) > 1 {
		return fmt.Errorf("multiple VPC peering connections matched; use additional constraints to reduce matches to a single VPC peering connection")
	}

	pcx := resp.VpcPeeringConnections[0]

	d.SetId(aws.StringValue(pcx.VpcPeeringConnectionId))
	d.Set("status", pcx.Status.Code)
	d.Set("vpc_id", pcx.RequesterVpcInfo.VpcId)
	d.Set("owner_id", pcx.RequesterVpcInfo.OwnerId)
	d.Set("cidr_block", pcx.RequesterVpcInfo.CidrBlock)
	cidrBlockSet := []interface{}{}
	for _, associationSet := range pcx.RequesterVpcInfo.CidrBlockSet {
		association := map[string]interface{}{
			"cidr_block": aws.StringValue(associationSet.CidrBlock),
		}
		cidrBlockSet = append(cidrBlockSet, association)
	}
	if err := d.Set("cidr_block_set", cidrBlockSet); err != nil {
		return fmt.Errorf("error setting cidr_block_set: %w", err)
	}
	d.Set("region", pcx.RequesterVpcInfo.Region)
	d.Set("peer_vpc_id", pcx.AccepterVpcInfo.VpcId)
	d.Set("peer_owner_id", pcx.AccepterVpcInfo.OwnerId)
	d.Set("peer_cidr_block", pcx.AccepterVpcInfo.CidrBlock)
	peerCidrBlockSet := []interface{}{}
	for _, associationSet := range pcx.AccepterVpcInfo.CidrBlockSet {
		association := map[string]interface{}{
			"cidr_block": aws.StringValue(associationSet.CidrBlock),
		}
		peerCidrBlockSet = append(peerCidrBlockSet, association)
	}
	if err := d.Set("peer_cidr_block_set", peerCidrBlockSet); err != nil {
		return fmt.Errorf("error setting peer_cidr_block_set: %w", err)
	}
	d.Set("peer_region", pcx.AccepterVpcInfo.Region)
	if err := d.Set("tags", KeyValueTags(pcx.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if pcx.AccepterVpcInfo.PeeringOptions != nil {
		if err := d.Set("accepter", flattenVPCPeeringConnectionOptions(pcx.AccepterVpcInfo.PeeringOptions)[0]); err != nil {
			return err
		}
	}

	if pcx.RequesterVpcInfo.PeeringOptions != nil {
		if err := d.Set("requester", flattenVPCPeeringConnectionOptions(pcx.RequesterVpcInfo.PeeringOptions)[0]); err != nil {
			return err
		}
	}

	return nil
}
