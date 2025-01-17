package ec2

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceCustomerGateway() *schema.Resource {
	return &schema.Resource{
		Create: resourceCustomerGatewayCreate,
		Read:   resourceCustomerGatewayRead,
		Update: resourceCustomerGatewayUpdate,
		Delete: resourceCustomerGatewayDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"bgp_asn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: valid4ByteASN,
			},

			"device_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
			},

			"ip_address": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.Any(
					validation.StringIsEmpty,
					validation.IsIPv4Address,
				),
			},

			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.GatewayTypeIpsec1,
				}, false),
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),

			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceCustomerGatewayCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	ipAddress := d.Get("ip_address").(string)
	vpnType := d.Get("type").(string)
	bgpAsn := d.Get("bgp_asn").(string)
	deviceName := d.Get("device_name").(string)

	alreadyExists, err := resourceCustomerGatewayExists(vpnType, ipAddress, bgpAsn, deviceName, conn)
	if err != nil {
		return err
	}

	if alreadyExists {
		return fmt.Errorf("An existing customer gateway for IpAddress: %s, VpnType: %s, BGP ASN: %s has been found", ipAddress, vpnType, bgpAsn)
	}

	i64BgpAsn, err := strconv.ParseInt(bgpAsn, 10, 64)
	if err != nil {
		return err
	}

	createOpts := &ec2.CreateCustomerGatewayInput{
		BgpAsn:            aws.Int64(i64BgpAsn),
		PublicIp:          aws.String(ipAddress),
		Type:              aws.String(vpnType),
		TagSpecifications: ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeCustomerGateway),
	}

	if len(deviceName) != 0 {
		createOpts.DeviceName = aws.String(deviceName)
	}

	// Create the Customer Gateway.
	log.Printf("[DEBUG] Creating customer gateway")
	resp, err := conn.CreateCustomerGateway(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating customer gateway: %s", err)
	}

	// Store the ID
	customerGateway := resp.CustomerGateway
	cgId := aws.StringValue(customerGateway.CustomerGatewayId)
	d.SetId(cgId)
	log.Printf("[INFO] Customer gateway ID: %s", cgId)

	// Wait for the CustomerGateway to be available.
	stateConf := &resource.StateChangeConf{
		Pending:    []string{"pending"},
		Target:     []string{"available"},
		Refresh:    customerGatewayRefreshFunc(conn, cgId),
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	_, stateErr := stateConf.WaitForState()

	if stateErr != nil {
		return fmt.Errorf(
			"Error waiting for customer gateway (%s) to become ready: %s", cgId, err)
	}

	return resourceCustomerGatewayRead(d, meta)
}

func customerGatewayRefreshFunc(conn *ec2.EC2, gatewayId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gatewayFilter := &ec2.Filter{
			Name:   aws.String("customer-gateway-id"),
			Values: []*string{aws.String(gatewayId)},
		}

		resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
			Filters: []*ec2.Filter{gatewayFilter},
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidCustomerGatewayID.NotFound", "") {
				resp = nil
			} else {
				log.Printf("Error on CustomerGatewayRefresh: %s", err)
				return nil, "", err
			}
		}

		if resp == nil || len(resp.CustomerGateways) == 0 {
			// handle consistency issues
			return nil, "", nil
		}

		gateway := resp.CustomerGateways[0]
		return gateway, *gateway.State, nil
	}
}

func resourceCustomerGatewayExists(vpnType, ipAddress, bgpAsn, deviceName string, conn *ec2.EC2) (bool, error) {
	filters := []*ec2.Filter{
		{
			Name:   aws.String("ip-address"),
			Values: []*string{aws.String(ipAddress)},
		},
		{
			Name:   aws.String("type"),
			Values: []*string{aws.String(vpnType)},
		},
		{
			Name:   aws.String("bgp-asn"),
			Values: []*string{aws.String(bgpAsn)},
		},
	}

	if len(deviceName) != 0 {
		filters = append(filters, &ec2.Filter{
			Name:   aws.String("device-name"),
			Values: []*string{aws.String(deviceName)},
		})
	}

	resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
		Filters: filters,
	})
	if err != nil {
		return false, err
	}

	if len(resp.CustomerGateways) > 0 && aws.StringValue(resp.CustomerGateways[0].State) != "deleted" {
		return true, nil
	}

	return false, nil
}

func resourceCustomerGatewayRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	gatewayFilter := &ec2.Filter{
		Name:   aws.String("customer-gateway-id"),
		Values: []*string{aws.String(d.Id())},
	}

	resp, err := conn.DescribeCustomerGateways(&ec2.DescribeCustomerGatewaysInput{
		Filters: []*ec2.Filter{gatewayFilter},
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidCustomerGatewayID.NotFound", "") {
			log.Printf("[WARN] Customer Gateway (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		} else {
			log.Printf("[ERROR] Error finding CustomerGateway: %s", err)
			return err
		}
	}

	if len(resp.CustomerGateways) != 1 {
		return fmt.Errorf("Error finding CustomerGateway: %s", d.Id())
	}

	if aws.StringValue(resp.CustomerGateways[0].State) == "deleted" {
		log.Printf("[INFO] Customer Gateway is in `deleted` state: %s", d.Id())
		d.SetId("")
		return nil
	}

	customerGateway := resp.CustomerGateways[0]
	d.Set("bgp_asn", customerGateway.BgpAsn)
	d.Set("ip_address", customerGateway.IpAddress)
	d.Set("type", customerGateway.Type)
	d.Set("device_name", customerGateway.DeviceName)

	tags := KeyValueTags(customerGateway.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   ec2.ServiceName,
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("customer-gateway/%s", d.Id()),
	}.String()

	d.Set("arn", arn)

	return nil
}

func resourceCustomerGatewayUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating EC2 Customer Gateway (%s) tags: %s", d.Id(), err)
		}
	}

	return resourceCustomerGatewayRead(d, meta)
}

func resourceCustomerGatewayDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	_, err := conn.DeleteCustomerGateway(&ec2.DeleteCustomerGatewayInput{
		CustomerGatewayId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, "InvalidCustomerGatewayID.NotFound", "") {
			return nil
		} else {
			return fmt.Errorf("[ERROR] Error deleting CustomerGateway: %s", err)
		}
	}

	gatewayFilter := &ec2.Filter{
		Name:   aws.String("customer-gateway-id"),
		Values: []*string{aws.String(d.Id())},
	}

	input := &ec2.DescribeCustomerGatewaysInput{
		Filters: []*ec2.Filter{gatewayFilter},
	}
	err = resource.Retry(5*time.Minute, func() *resource.RetryError {
		resp, err := conn.DescribeCustomerGateways(input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidCustomerGatewayID.NotFound", "") {
				return nil
			}
			return resource.NonRetryableError(err)
		}

		err = checkGatewayDeleteResponse(resp, d.Id())
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		var resp *ec2.DescribeCustomerGatewaysOutput
		resp, err = conn.DescribeCustomerGateways(input)

		if err != nil {
			return checkGatewayDeleteResponse(resp, d.Id())
		}
	}

	if err != nil {
		return fmt.Errorf("Error deleting customer gateway: %s", err)
	}
	return nil

}

func checkGatewayDeleteResponse(resp *ec2.DescribeCustomerGatewaysOutput, id string) error {
	if len(resp.CustomerGateways) != 1 {
		return fmt.Errorf("Error finding CustomerGateway for delete: %s", id)
	}

	cgState := aws.StringValue(resp.CustomerGateways[0].State)
	switch cgState {
	case "pending", "available", "deleting":
		return fmt.Errorf("Gateway (%s) in state (%s), retrying", id, cgState)
	case "deleted":
		return nil
	default:
		return fmt.Errorf("Unrecognized state (%s) for Customer Gateway delete on (%s)", *resp.CustomerGateways[0].State, id)
	}
}
