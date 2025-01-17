package ec2

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceSubnet() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceSubnetCreate,
		Read:   resourceSubnetRead,
		Update: resourceSubnetUpdate,
		Delete: resourceSubnetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,

		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(20 * time.Minute),
		},

		SchemaVersion: 1,
		MigrateState:  SubnetMigrateState,

		Schema: map[string]*schema.Schema{
			"vpc_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"cidr_block": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidIPv4CIDRNetworkAddress,
			},

			"ipv6_cidr_block": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: verify.ValidIPv6CIDRNetworkAddress,
			},

			"availability_zone": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone_id"},
			},

			"availability_zone_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"availability_zone"},
			},

			"customer_owned_ipv4_pool": {
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"map_customer_owned_ip_on_launch", "outpost_arn"},
			},

			"map_customer_owned_ip_on_launch": {
				Type:         schema.TypeBool,
				Optional:     true,
				RequiredWith: []string{"customer_owned_ipv4_pool", "outpost_arn"},
			},

			"map_public_ip_on_launch": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"outpost_arn": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},

			"assign_ipv6_address_on_creation": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},

			"ipv6_cidr_block_association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags": tftags.TagsSchema(),

			"tags_all": tftags.TagsSchemaComputed(),

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceSubnetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	createOpts := &ec2.CreateSubnetInput{
		AvailabilityZone:   aws.String(d.Get("availability_zone").(string)),
		AvailabilityZoneId: aws.String(d.Get("availability_zone_id").(string)),
		CidrBlock:          aws.String(d.Get("cidr_block").(string)),
		VpcId:              aws.String(d.Get("vpc_id").(string)),
		TagSpecifications:  ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeSubnet),
	}

	if v, ok := d.GetOk("ipv6_cidr_block"); ok {
		createOpts.Ipv6CidrBlock = aws.String(v.(string))
	}

	if v, ok := d.GetOk("outpost_arn"); ok {
		createOpts.OutpostArn = aws.String(v.(string))
	}

	var err error
	resp, err := conn.CreateSubnet(createOpts)

	if err != nil {
		return fmt.Errorf("error creating subnet: %w", err)
	}

	// Get the ID and store it
	subnet := resp.Subnet
	subnetId := aws.StringValue(subnet.SubnetId)
	d.SetId(subnetId)
	log.Printf("[INFO] Subnet ID: %s", subnetId)

	// Wait for the Subnet to become available
	log.Printf("[DEBUG] Waiting for subnet (%s) to become available", subnetId)
	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.SubnetStatePending},
		Target:  []string{ec2.SubnetStateAvailable},
		Refresh: SubnetStateRefreshFunc(conn, subnetId),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	_, err = stateConf.WaitForState()

	if err != nil {
		return fmt.Errorf("error waiting for subnet (%s) to become ready: %w", d.Id(), err)
	}

	// You cannot modify multiple subnet attributes in the same request,
	// except CustomerOwnedIpv4Pool and MapCustomerOwnedIpOnLaunch.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySubnetAttribute.html

	if d.Get("assign_ipv6_address_on_creation").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error enabling EC2 Subnet (%s) assign IPv6 address on creation: %w", d.Id(), err)
		}
	}

	if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
		input := &ec2.ModifySubnetAttributeInput{
			CustomerOwnedIpv4Pool: aws.String(v.(string)),
			MapCustomerOwnedIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_customer_owned_ip_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error setting EC2 Subnet (%s) customer owned IPv4 pool and map customer owned IP on launch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(conn, d.Id(), d.Get("map_customer_owned_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) map customer owned IP on launch update: %w", d.Id(), err)
		}
	}

	if d.Get("map_public_ip_on_launch").(bool) {
		input := &ec2.ModifySubnetAttributeInput{
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(true),
			},
			SubnetId: aws.String(d.Id()),
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error enabling EC2 Subnet (%s) map public IP on launch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapPublicIPOnLaunchUpdated(conn, d.Id(), d.Get("map_public_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) map public IP on launch update: %w", d.Id(), err)
		}
	}

	return resourceSubnetRead(d, meta)
}

func resourceSubnetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	var subnet *ec2.Subnet

	err := resource.Retry(SubnetPropagationTimeout, func() *resource.RetryError {
		var err error

		subnet, err = FindSubnetByID(conn, d.Id())

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidSubnetID.NotFound") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if d.IsNewResource() && subnet == nil {
			return resource.RetryableError(&resource.NotFoundError{
				LastError: fmt.Errorf("EC2 Subnet (%s) not found", d.Id()),
			})
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		subnet, err = FindSubnetByID(conn, d.Id())
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, "InvalidSubnetID.NotFound") {
		log.Printf("[WARN] EC2 Subnet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Subnet (%s): %w", d.Id(), err)
	}

	if subnet == nil {
		if d.IsNewResource() {
			return fmt.Errorf("error reading EC2 Subnet (%s): not found after creation", d.Id())
		}

		log.Printf("[WARN] EC2 Subnet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	d.Set("vpc_id", subnet.VpcId)
	d.Set("availability_zone", subnet.AvailabilityZone)
	d.Set("availability_zone_id", subnet.AvailabilityZoneId)
	d.Set("cidr_block", subnet.CidrBlock)
	d.Set("customer_owned_ipv4_pool", subnet.CustomerOwnedIpv4Pool)
	d.Set("map_customer_owned_ip_on_launch", subnet.MapCustomerOwnedIpOnLaunch)
	d.Set("map_public_ip_on_launch", subnet.MapPublicIpOnLaunch)
	d.Set("assign_ipv6_address_on_creation", subnet.AssignIpv6AddressOnCreation)
	d.Set("outpost_arn", subnet.OutpostArn)

	// Make sure those values are set, if an IPv6 block exists it'll be set in the loop
	d.Set("ipv6_cidr_block_association_id", "")
	d.Set("ipv6_cidr_block", "")

	for _, a := range subnet.Ipv6CidrBlockAssociationSet {
		if aws.StringValue(a.Ipv6CidrBlockState.State) == ec2.SubnetCidrBlockStateCodeAssociated { //we can only ever have 1 IPv6 block associated at once
			d.Set("ipv6_cidr_block_association_id", a.AssociationId)
			d.Set("ipv6_cidr_block", a.Ipv6CidrBlock)
			break
		}
	}

	d.Set("arn", subnet.SubnetArn)

	tags := KeyValueTags(subnet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	d.Set("owner_id", subnet.OwnerId)

	return nil
}

func resourceSubnetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Subnet (%s) tags: %w", d.Id(), err)
		}
	}

	// You cannot modify multiple subnet attributes in the same request,
	// except CustomerOwnedIpv4Pool and MapCustomerOwnedIpOnLaunch.
	// Reference: https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_ModifySubnetAttribute.html

	if d.HasChanges("customer_owned_ipv4_pool", "map_customer_owned_ip_on_launch") {
		input := &ec2.ModifySubnetAttributeInput{
			MapCustomerOwnedIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_customer_owned_ip_on_launch").(bool)),
			},
			SubnetId: aws.String(d.Id()),
		}

		if v, ok := d.GetOk("customer_owned_ipv4_pool"); ok {
			input.CustomerOwnedIpv4Pool = aws.String(v.(string))
		}

		if _, err := conn.ModifySubnetAttribute(input); err != nil {
			return fmt.Errorf("error updating EC2 Subnet (%s) customer owned IPv4 pool and map customer owned IP on launch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapCustomerOwnedIPOnLaunchUpdated(conn, d.Id(), d.Get("map_customer_owned_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) map customer owned IP on launch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("map_public_ip_on_launch") {
		modifyOpts := &ec2.ModifySubnetAttributeInput{
			SubnetId: aws.String(d.Id()),
			MapPublicIpOnLaunch: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("map_public_ip_on_launch").(bool)),
			},
		}

		_, err := conn.ModifySubnetAttribute(modifyOpts)

		if err != nil {
			return fmt.Errorf("error updating EC2 Subnet (%s) map public IP on launch: %w", d.Id(), err)
		}

		if _, err := WaitSubnetMapPublicIPOnLaunchUpdated(conn, d.Id(), d.Get("map_public_ip_on_launch").(bool)); err != nil {
			return fmt.Errorf("error waiting for EC2 Subnet (%s) map public IP on launch update: %w", d.Id(), err)
		}
	}

	if d.HasChange("ipv6_cidr_block") {
		// We need to handle that we disassociate the IPv6 CIDR block before we try to associate the new one
		// This could be an issue as, we could error out when we try to add the new one
		// We may need to roll back the state and reattach the old one if this is the case

		newIpv6 := d.Get("ipv6_cidr_block").(string)

		if v, ok := d.GetOk("ipv6_cidr_block_association_id"); ok {

			ipv6AssignOnCreate := d.Get("assign_ipv6_address_on_creation").(bool)

			if !ipv6AssignOnCreate {
				modifyOpts := &ec2.ModifySubnetAttributeInput{
					SubnetId: aws.String(d.Id()),
					AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
						Value: aws.Bool(false),
					},
				}

				log.Printf("[DEBUG] Subnet modify attributes: %#v", modifyOpts)

				_, err := conn.ModifySubnetAttribute(modifyOpts)

				if err != nil {
					return fmt.Errorf("error modifying EC2 Subnet (%s) attribute: %w", d.Id(), err)
				}
			}
			//Firstly we have to disassociate the old IPv6 CIDR Block
			disassociateOps := &ec2.DisassociateSubnetCidrBlockInput{
				AssociationId: aws.String(v.(string)),
			}

			_, err := conn.DisassociateSubnetCidrBlock(disassociateOps)
			if err != nil {
				return err
			}

			// Wait for the CIDR to become disassociated
			log.Printf("[DEBUG] Waiting for IPv6 CIDR (%s) to become disassociated", d.Id())
			stateConf := &resource.StateChangeConf{
				Pending: []string{ec2.SubnetCidrBlockStateCodeDisassociating, ec2.SubnetCidrBlockStateCodeAssociated},
				Target:  []string{ec2.SubnetCidrBlockStateCodeDisassociated},
				Refresh: SubnetIpv6CidrStateRefreshFunc(conn, d.Id(), d.Get("ipv6_cidr_block_association_id").(string)),
				Timeout: 3 * time.Minute,
			}
			if _, err := stateConf.WaitForState(); err != nil {
				return fmt.Errorf("Error waiting for IPv6 CIDR (%s) to become disassociated: %w", d.Id(), err)
			}
		}

		if newIpv6 != "" {
			//Now we need to try to associate the new CIDR block
			associatesOpts := &ec2.AssociateSubnetCidrBlockInput{
				SubnetId:      aws.String(d.Id()),
				Ipv6CidrBlock: aws.String(newIpv6),
			}

			resp, err := conn.AssociateSubnetCidrBlock(associatesOpts)
			if err != nil {
				//The big question here is, do we want to try to reassociate the old one??
				//If we have a failure here, then we may be in a situation that we have nothing associated
				return fmt.Errorf("error associating EC2 Subnet (%s) CIDR block: %w", d.Id(), err)
			}

			// Wait for the CIDR to become associated
			log.Printf(
				"[DEBUG] Waiting for IPv6 CIDR (%s) to become associated",
				d.Id())
			stateConf := &resource.StateChangeConf{
				Pending: []string{ec2.SubnetCidrBlockStateCodeAssociating, ec2.SubnetCidrBlockStateCodeDisassociated},
				Target:  []string{ec2.SubnetCidrBlockStateCodeAssociated},
				Refresh: SubnetIpv6CidrStateRefreshFunc(conn, d.Id(), aws.StringValue(resp.Ipv6CidrBlockAssociation.AssociationId)),
				Timeout: 3 * time.Minute,
			}
			if _, err := stateConf.WaitForState(); err != nil {
				return fmt.Errorf(
					"Error waiting for IPv6 CIDR (%s) to become associated: %w",
					d.Id(), err)
			}

		}
	}

	if d.HasChange("assign_ipv6_address_on_creation") {
		modifyOpts := &ec2.ModifySubnetAttributeInput{
			SubnetId: aws.String(d.Id()),
			AssignIpv6AddressOnCreation: &ec2.AttributeBooleanValue{
				Value: aws.Bool(d.Get("assign_ipv6_address_on_creation").(bool)),
			},
		}

		log.Printf("[DEBUG] Subnet modify attributes: %#v", modifyOpts)

		_, err := conn.ModifySubnetAttribute(modifyOpts)

		if err != nil {
			return err
		}
	}

	return resourceSubnetRead(d, meta)
}

func resourceSubnetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	log.Printf("[INFO] Deleting subnet: %s", d.Id())

	if err := deleteLingeringLambdaENIs(conn, "subnet-id", d.Id(), d.Timeout(schema.TimeoutDelete)); err != nil {
		return fmt.Errorf("error deleting Lambda ENIs using subnet (%s): %w", d.Id(), err)
	}

	req := &ec2.DeleteSubnetInput{
		SubnetId: aws.String(d.Id()),
	}

	wait := resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"destroyed"},
		Timeout:    d.Timeout(schema.TimeoutDelete),
		MinTimeout: 1 * time.Second,
		Refresh: func() (interface{}, string, error) {
			_, err := conn.DeleteSubnet(req)
			if err != nil {
				if tfawserr.ErrMessageContains(err, "DependencyViolation", "") {
					// There is some pending operation, so just retry
					// in a bit.
					return 42, "pending", nil
				}
				if tfawserr.ErrMessageContains(err, "InvalidSubnetID.NotFound", "") {
					return 42, "destroyed", nil
				}

				return 42, "failure", err
			}

			return 42, "destroyed", nil
		},
	}

	if _, err := wait.WaitForState(); err != nil {
		return fmt.Errorf("error deleting subnet (%s): %w", d.Id(), err)
	}

	return nil
}

// SubnetStateRefreshFunc returns a resource.StateRefreshFunc that is used to watch a Subnet.
func SubnetStateRefreshFunc(conn *ec2.EC2, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := conn.DescribeSubnets(&ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(id)},
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidSubnetID.NotFound", "") {
				resp = nil
			} else {
				log.Printf("Error on SubnetStateRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		subnet := resp.Subnets[0]
		return subnet, aws.StringValue(subnet.State), nil
	}
}

func SubnetIpv6CidrStateRefreshFunc(conn *ec2.EC2, id string, associationId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		opts := &ec2.DescribeSubnetsInput{
			SubnetIds: []*string{aws.String(id)},
		}
		resp, err := conn.DescribeSubnets(opts)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidSubnetID.NotFound", "") {
				resp = nil
			} else {
				log.Printf("Error on SubnetIpv6CidrStateRefreshFunc: %s", err)
				return nil, "", err
			}
		}

		if resp == nil {
			// Sometimes AWS just has consistency issues and doesn't see
			// our instance yet. Return an empty state.
			return nil, "", nil
		}

		if resp.Subnets[0].Ipv6CidrBlockAssociationSet == nil {
			return nil, "", nil
		}

		for _, association := range resp.Subnets[0].Ipv6CidrBlockAssociationSet {
			if aws.StringValue(association.AssociationId) == associationId {
				return association, aws.StringValue(association.Ipv6CidrBlockState.State), nil
			}
		}

		return nil, "", nil
	}
}
