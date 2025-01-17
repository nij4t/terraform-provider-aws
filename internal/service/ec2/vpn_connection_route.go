package ec2

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceVPNConnectionRoute() *schema.Resource {
	return &schema.Resource{
		// You can't update a route. You can just delete one and make
		// a new one.
		Create: resourceVPNConnectionRouteCreate,
		Read:   resourceVPNConnectionRouteRead,
		Delete: resourceVPNConnectionRouteDelete,

		Schema: map[string]*schema.Schema{
			"destination_cidr_block": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"vpn_connection_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVPNConnectionRouteCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock := d.Get("destination_cidr_block").(string)
	vpnConnectionId := d.Get("vpn_connection_id").(string)
	createOpts := &ec2.CreateVpnConnectionRouteInput{
		DestinationCidrBlock: aws.String(cidrBlock),
		VpnConnectionId:      aws.String(vpnConnectionId),
	}

	// Create the route.
	log.Printf("[DEBUG] Creating VPN connection route")
	_, err := conn.CreateVpnConnectionRoute(createOpts)
	if err != nil {
		return fmt.Errorf("Error creating VPN connection route: %s", err)
	}

	// Store the ID by the only two data we have available to us.
	d.SetId(fmt.Sprintf("%s:%s", *createOpts.DestinationCidrBlock, *createOpts.VpnConnectionId))

	stateConf := resource.StateChangeConf{
		Pending: []string{"pending"},
		Target:  []string{"available"},
		Timeout: 15 * time.Second,
		Refresh: func() (interface{}, string, error) {
			route, err := findConnectionRoute(conn, cidrBlock, vpnConnectionId)
			if err != nil {
				return 42, "", err
			}
			return route, *route.State, nil
		},
	}
	_, err = stateConf.WaitForState()
	if err != nil {
		return err
	}

	return resourceVPNConnectionRouteRead(d, meta)
}

func resourceVPNConnectionRouteRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock, vpnConnectionId := VPNConnectionRouteParseID(d.Id())

	route, err := findConnectionRoute(conn, cidrBlock, vpnConnectionId)
	if err != nil {
		return err
	}
	if route == nil {
		// Something other than terraform eliminated the route.
		d.SetId("")
	}

	return nil
}

func resourceVPNConnectionRouteDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	cidrBlock := d.Get("destination_cidr_block").(string)
	vpnConnectionId := d.Get("vpn_connection_id").(string)
	_, err := conn.DeleteVpnConnectionRoute(&ec2.DeleteVpnConnectionRouteInput{
		DestinationCidrBlock: aws.String(cidrBlock),
		VpnConnectionId:      aws.String(vpnConnectionId),
	})
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpnConnectionID.NotFound" {
			return nil
		}
		log.Printf("[ERROR] Error deleting VPN connection route: %s", err)
		return err
	}

	stateConf := resource.StateChangeConf{
		Pending: []string{"pending", "available", "deleting"},
		Target:  []string{"deleted"},
		Timeout: 15 * time.Second,
		Refresh: func() (interface{}, string, error) {
			route, err := findConnectionRoute(conn, cidrBlock, vpnConnectionId)
			if err != nil {
				return 42, "", err
			}
			if route == nil {
				return 42, "deleted", nil
			}
			return route, *route.State, nil
		},
	}
	_, err = stateConf.WaitForState()
	return err
}

func findConnectionRoute(conn *ec2.EC2, cidrBlock, vpnConnectionId string) (*ec2.VpnStaticRoute, error) {
	resp, err := conn.DescribeVpnConnections(&ec2.DescribeVpnConnectionsInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("route.destination-cidr-block"),
				Values: []*string{aws.String(cidrBlock)},
			},
			{
				Name:   aws.String("vpn-connection-id"),
				Values: []*string{aws.String(vpnConnectionId)},
			},
		},
	})
	if err != nil {
		if ec2err, ok := err.(awserr.Error); ok && ec2err.Code() == "InvalidVpnConnectionID.NotFound" {
			return nil, nil
		}
		return nil, err
	}
	if resp == nil || len(resp.VpnConnections) == 0 {
		return nil, nil
	}
	vpnConnection := resp.VpnConnections[0]

	for _, r := range vpnConnection.Routes {
		if aws.StringValue(r.DestinationCidrBlock) == cidrBlock && aws.StringValue(r.State) != "deleted" {
			return r, nil
		}
	}
	return nil, nil
}

func VPNConnectionRouteParseID(id string) (string, string) {
	parts := strings.SplitN(id, ":", 2)
	return parts[0], parts[1]
}
