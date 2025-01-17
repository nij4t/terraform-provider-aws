package sns

import (
	"crypto/sha256"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfiam "github.com/nij4t/terraform-provider-aws/internal/service/iam"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func ResourcePlatformApplication() *schema.Resource {
	return &schema.Resource{
		Create: resourcePlatformApplicationCreate,
		Read:   resourcePlatformApplicationRead,
		Update: resourcePlatformApplicationUpdate,
		Delete: resourcePlatformApplicationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"platform": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"platform_credential": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"event_delivery_failure_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_endpoint_created_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_endpoint_deleted_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"event_endpoint_updated_topic_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"failure_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"platform_principal": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
			},
			"success_feedback_role_arn": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"success_feedback_sample_rate": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourcePlatformApplicationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	attributes := make(map[string]*string)
	name := d.Get("name").(string)
	platform := d.Get("platform").(string)

	attributes["PlatformCredential"] = aws.String(d.Get("platform_credential").(string))
	if v, ok := d.GetOk("platform_principal"); ok {
		attributes["PlatformPrincipal"] = aws.String(v.(string))
	}

	req := &sns.CreatePlatformApplicationInput{
		Name:       aws.String(name),
		Platform:   aws.String(platform),
		Attributes: attributes,
	}

	log.Printf("[DEBUG] SNS create application: %s", req)

	output, err := conn.CreatePlatformApplication(req)
	if err != nil {
		return fmt.Errorf("Error creating SNS platform application: %s", err)
	}

	d.SetId(aws.StringValue(output.PlatformApplicationArn))

	return resourcePlatformApplicationUpdate(d, meta)
}

func resourcePlatformApplicationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	attributes := make(map[string]*string)

	if d.HasChange("event_delivery_failure_topic_arn") {
		attributes["EventDeliveryFailure"] = aws.String(d.Get("event_delivery_failure_topic_arn").(string))
	}

	if d.HasChange("event_endpoint_created_topic_arn") {
		attributes["EventEndpointCreated"] = aws.String(d.Get("event_endpoint_created_topic_arn").(string))
	}

	if d.HasChange("event_endpoint_deleted_topic_arn") {
		attributes["EventEndpointDeleted"] = aws.String(d.Get("event_endpoint_deleted_topic_arn").(string))
	}

	if d.HasChange("event_endpoint_updated_topic_arn") {
		attributes["EventEndpointUpdated"] = aws.String(d.Get("event_endpoint_updated_topic_arn").(string))
	}

	if d.HasChange("failure_feedback_role_arn") {
		attributes["FailureFeedbackRoleArn"] = aws.String(d.Get("failure_feedback_role_arn").(string))
	}

	if d.HasChange("success_feedback_role_arn") {
		attributes["SuccessFeedbackRoleArn"] = aws.String(d.Get("success_feedback_role_arn").(string))
	}

	if d.HasChange("success_feedback_sample_rate") {
		attributes["SuccessFeedbackSampleRate"] = aws.String(d.Get("success_feedback_sample_rate").(string))
	}

	if d.HasChanges("platform_credential", "platform_principal") {
		// Prior to version 3.0.0 of the Terraform AWS Provider, the platform_credential and platform_principal
		// attributes were stored in state as SHA256 hashes. If the changes to these two attributes are the only
		// changes and if both of their changes only match updating the state value, then skip the API call.
		oPCRaw, nPCRaw := d.GetChange("platform_credential")
		oPPRaw, nPPRaw := d.GetChange("platform_principal")

		if len(attributes) == 0 && isChangeSha256Removal(oPCRaw, nPCRaw) && isChangeSha256Removal(oPPRaw, nPPRaw) {
			return nil
		}

		attributes["PlatformCredential"] = aws.String(d.Get("platform_credential").(string))
		// If the platform requires a principal it must also be specified, even if it didn't change
		// since credential is stored as a hash, the only way to update principal is to update both
		// as they must be specified together in the request.
		if v, ok := d.GetOk("platform_principal"); ok {
			attributes["PlatformPrincipal"] = aws.String(v.(string))
		}
	}

	// Make API call to update attributes
	req := &sns.SetPlatformApplicationAttributesInput{
		PlatformApplicationArn: aws.String(d.Id()),
		Attributes:             attributes,
	}

	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		_, err := conn.SetPlatformApplicationAttributes(req)
		if err != nil {
			if tfawserr.ErrMessageContains(err, sns.ErrCodeInvalidParameterException, "is not a valid role to allow SNS to write to Cloudwatch Logs") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.SetPlatformApplicationAttributes(req)
	}

	if err != nil {
		return fmt.Errorf("Error updating SNS platform application: %s", err)
	}

	return resourcePlatformApplicationRead(d, meta)
}

func resourcePlatformApplicationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	// There is no SNS Describe/GetPlatformApplication to fetch attributes like name and platform
	// We will use the ID, which should be a platform application ARN, to:
	//  * Validate its an appropriate ARN on import
	//  * Parse out the name and platform
	arn, name, platform, err := DecodePlatformApplicationID(d.Id())
	if err != nil {
		return err
	}

	d.Set("arn", arn)
	d.Set("name", name)
	d.Set("platform", platform)

	input := &sns.GetPlatformApplicationAttributesInput{
		PlatformApplicationArn: aws.String(arn),
	}

	output, err := conn.GetPlatformApplicationAttributes(input)

	if tfawserr.ErrMessageContains(err, sns.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] SNS Platform Application (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting SNS Platform Application (%s) attributes: %w", d.Id(), err)
	}

	if output == nil || output.Attributes == nil {
		return fmt.Errorf("error getting SNS Platform Application (%s) attributes: empty response", d.Id())
	}

	if v, ok := output.Attributes["EventDeliveryFailure"]; ok {
		d.Set("event_delivery_failure_topic_arn", v)
	}

	if v, ok := output.Attributes["EventEndpointCreated"]; ok {
		d.Set("event_endpoint_created_topic_arn", v)
	}

	if v, ok := output.Attributes["EventEndpointDeleted"]; ok {
		d.Set("event_endpoint_deleted_topic_arn", v)
	}

	if v, ok := output.Attributes["EventEndpointUpdated"]; ok {
		d.Set("event_endpoint_updated_topic_arn", v)
	}

	if v, ok := output.Attributes["FailureFeedbackRoleArn"]; ok {
		d.Set("failure_feedback_role_arn", v)
	}

	if v, ok := output.Attributes["PlatformPrincipal"]; ok {
		d.Set("platform_principal", v)
	}

	if v, ok := output.Attributes["SuccessFeedbackRoleArn"]; ok {
		d.Set("success_feedback_role_arn", v)
	}

	if v, ok := output.Attributes["SuccessFeedbackSampleRate"]; ok {
		d.Set("success_feedback_sample_rate", v)
	}

	return nil
}

func resourcePlatformApplicationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SNSConn

	log.Printf("[DEBUG] SNS Delete Application: %s", d.Id())
	_, err := conn.DeletePlatformApplication(&sns.DeletePlatformApplicationInput{
		PlatformApplicationArn: aws.String(d.Id()),
	})
	return err
}

func DecodePlatformApplicationID(input string) (arnS, name, platform string, err error) {
	platformApplicationArn, err := arn.Parse(input)
	if err != nil {
		err = fmt.Errorf(
			"SNS Platform Application ID must be of the form "+
				"arn:PARTITION:sns:REGION:ACCOUNTID:app/PLATFORM/NAME, "+
				"was provided %q and received error: %s", input, err)
		return
	}

	platformApplicationArnResourceParts := strings.Split(platformApplicationArn.Resource, "/")
	if len(platformApplicationArnResourceParts) != 3 || platformApplicationArnResourceParts[0] != "app" {
		err = fmt.Errorf(
			"SNS Platform Application ID must be of the form "+
				"arn:PARTITION:sns:REGION:ACCOUNTID:app/PLATFORM/NAME, "+
				"was provided: %s", input)
		return
	}

	arnS = platformApplicationArn.String()
	name = platformApplicationArnResourceParts[2]
	platform = platformApplicationArnResourceParts[1]
	return
}

func isChangeSha256Removal(oldRaw, newRaw interface{}) bool {
	old, ok := oldRaw.(string)

	if !ok {
		return false
	}

	new, ok := newRaw.(string)

	if !ok {
		return false
	}

	return fmt.Sprintf("%x", sha256.Sum256([]byte(new))) == old
}
