package pinpoint

import (
	"errors"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/pinpoint"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceAPNSSandboxChannel() *schema.Resource {
	return &schema.Resource{
		Create: resourceAPNSSandboxChannelUpsert,
		Read:   resourceAPNSSandboxChannelRead,
		Update: resourceAPNSSandboxChannelUpsert,
		Delete: resourceAPNSSandboxChannelDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"application_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bundle_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"certificate": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"default_authentication_method": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"team_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"token_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"token_key_id": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

func resourceAPNSSandboxChannelUpsert(d *schema.ResourceData, meta interface{}) error {
	certificate, certificateOk := d.GetOk("certificate")
	privateKey, privateKeyOk := d.GetOk("private_key")

	bundleId, bundleIdOk := d.GetOk("bundle_id")
	teamId, teamIdOk := d.GetOk("team_id")
	tokenKey, tokenKeyOk := d.GetOk("token_key")
	tokenKeyId, tokenKeyIdOk := d.GetOk("token_key_id")

	if !(certificateOk && privateKeyOk) && !(bundleIdOk && teamIdOk && tokenKeyOk && tokenKeyIdOk) {
		return errors.New("At least one set of credentials is required; either [certificate, private_key] or [bundle_id, team_id, token_key, token_key_id]")
	}

	conn := meta.(*conns.AWSClient).PinpointConn

	applicationId := d.Get("application_id").(string)

	params := &pinpoint.APNSSandboxChannelRequest{}

	params.DefaultAuthenticationMethod = aws.String(d.Get("default_authentication_method").(string))
	params.Enabled = aws.Bool(d.Get("enabled").(bool))

	params.Certificate = aws.String(certificate.(string))
	params.PrivateKey = aws.String(privateKey.(string))

	params.BundleId = aws.String(bundleId.(string))
	params.TeamId = aws.String(teamId.(string))
	params.TokenKey = aws.String(tokenKey.(string))
	params.TokenKeyId = aws.String(tokenKeyId.(string))

	req := pinpoint.UpdateApnsSandboxChannelInput{
		ApplicationId:             aws.String(applicationId),
		APNSSandboxChannelRequest: params,
	}

	_, err := conn.UpdateApnsSandboxChannel(&req)
	if err != nil {
		return fmt.Errorf("error updating Pinpoint APNs Sandbox Channel for Application %s: %s", applicationId, err)
	}

	d.SetId(applicationId)

	return resourceAPNSSandboxChannelRead(d, meta)
}

func resourceAPNSSandboxChannelRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[INFO] Reading Pinpoint APNs Channel for Application %s", d.Id())

	output, err := conn.GetApnsSandboxChannel(&pinpoint.GetApnsSandboxChannelInput{
		ApplicationId: aws.String(d.Id()),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] Pinpoint APNs Sandbox Channel for application %s not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("error getting Pinpoint APNs Sandbox Channel for application %s: %s", d.Id(), err)
	}

	d.Set("application_id", output.APNSSandboxChannelResponse.ApplicationId)
	d.Set("default_authentication_method", output.APNSSandboxChannelResponse.DefaultAuthenticationMethod)
	d.Set("enabled", output.APNSSandboxChannelResponse.Enabled)
	// Sensitive params are not returned

	return nil
}

func resourceAPNSSandboxChannelDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).PinpointConn

	log.Printf("[DEBUG] Deleting Pinpoint APNs Sandbox Channel: %s", d.Id())
	_, err := conn.DeleteApnsSandboxChannel(&pinpoint.DeleteApnsSandboxChannelInput{
		ApplicationId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, pinpoint.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Pinpoint APNs Sandbox Channel for Application %s: %s", d.Id(), err)
	}
	return nil
}
