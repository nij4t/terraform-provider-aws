package events

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfiam "github.com/nij4t/terraform-provider-aws/internal/service/iam"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceBusPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceBusPolicyCreate,
		Read:   resourceBusPolicyRead,
		Update: resourceBusPolicyUpdate,
		Delete: resourceBusPolicyDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				d.Set("event_bus_name", d.Id())
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"event_bus_name": {
				Type:         schema.TypeString,
				Optional:     true,
				ForceNew:     true,
				ValidateFunc: validBusNameOrARN,
				Default:      DefaultEventBusName,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceBusPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName := d.Get("event_bus_name").(string)
	policy := d.Get("policy").(string)

	input := eventbridge.PutPermissionInput{
		EventBusName: aws.String(eventBusName),
		Policy:       aws.String(policy),
	}

	log.Printf("[DEBUG] Creating EventBridge policy: %s", input)
	_, err := conn.PutPermission(&input)
	if err != nil {
		return fmt.Errorf("Creating EventBridge policy failed: %w", err)
	}

	d.SetId(eventBusName)

	return resourceBusPolicyRead(d, meta)
}

// See also: https://docs.aws.amazon.com/eventbridge/latest/APIReference/API_DescribeEventBus.html
func resourceBusPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName := d.Id()

	input := eventbridge.DescribeEventBusInput{
		Name: aws.String(eventBusName),
	}
	var output *eventbridge.DescribeEventBusOutput
	var err error
	var policy *string

	// Especially with concurrent PutPermission calls there can be a slight delay
	err = resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		log.Printf("[DEBUG] Reading EventBridge bus: %s", input)
		output, err = conn.DescribeEventBus(&input)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("reading EventBridge permission (%s) failed: %w", d.Id(), err))
		}

		policy, err = getEventBusPolicy(output)
		if err != nil {
			return resource.RetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.DescribeEventBus(&input)
		if output != nil {
			policy, err = getEventBusPolicy(output)
		}
	}

	if tfresource.NotFound(err) {
		log.Printf("[WARN] Policy on {%s} EventBus not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error reading policy from EventBridge Bus (%s): %w", d.Id(), err)
	}

	busName := aws.StringValue(output.Name)
	if busName == "" {
		busName = DefaultEventBusName
	}
	d.Set("event_bus_name", busName)

	d.Set("policy", policy)

	return nil
}

func getEventBusPolicy(output *eventbridge.DescribeEventBusOutput) (*string, error) {
	if output == nil || output.Policy == nil {
		return nil, &resource.NotFoundError{
			Message:      fmt.Sprintf("Policy for EventBridge Bus %s not found", *output.Name),
			LastResponse: output,
		}
	}

	return output.Policy, nil
}

func resourceBusPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName := d.Id()

	input := eventbridge.PutPermissionInput{
		EventBusName: aws.String(eventBusName),
		Policy:       aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Update EventBridge Bus policy: %s", input)
	_, err := conn.PutPermission(&input)
	if tfawserr.ErrMessageContains(err, eventbridge.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] EventBridge Bus %q not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("error updating policy for EventBridge Bus (%s): %w", d.Id(), err)
	}

	return resourceBusPolicyRead(d, meta)
}

func resourceBusPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EventsConn

	eventBusName := d.Id()
	removeAllPermissions := true

	input := eventbridge.RemovePermissionInput{
		EventBusName:         aws.String(eventBusName),
		RemoveAllPermissions: &removeAllPermissions,
	}

	log.Printf("[DEBUG] Delete EventBridge Bus Policy: %s", input)
	_, err := conn.RemovePermission(&input)
	if tfawserr.ErrMessageContains(err, eventbridge.ErrCodeResourceNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("error deleting policy for EventBridge Bus (%s): %w", d.Id(), err)
	}
	return nil
}
