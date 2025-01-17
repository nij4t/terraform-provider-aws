package networkfirewall

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceResourcePolicyPut,
		ReadContext:   resourceResourcePolicyRead,
		UpdateContext: resourceResourcePolicyPut,
		DeleteContext: resourceResourcePolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"resource_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
		},
	}
}

func resourceResourcePolicyPut(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	resourceArn := d.Get("resource_arn").(string)
	input := &networkfirewall.PutResourcePolicyInput{
		ResourceArn: aws.String(resourceArn),
		Policy:      aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Putting NetworkFirewall Resource Policy for resource: %s", resourceArn)

	_, err := conn.PutResourcePolicyWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error putting NetworkFirewall Resource Policy (for resource: %s): %w", resourceArn, err))
	}

	d.SetId(resourceArn)

	return resourceResourcePolicyRead(ctx, d, meta)
}

func resourceResourcePolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	resourceArn := d.Id()

	log.Printf("[DEBUG] Reading NetworkFirewall Resource Policy for resource: %s", resourceArn)

	policy, err := FindResourcePolicy(ctx, conn, resourceArn)
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Resource Policy (for resource: %s) not found, removing from state", resourceArn)
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Resource Policy (for resource: %s): %w", resourceArn, err))
	}

	if policy == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Resource Policy (for resource: %s): empty output", resourceArn))
	}

	d.Set("policy", policy)
	d.Set("resource_arn", resourceArn)

	return nil
}

func resourceResourcePolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn

	log.Printf("[DEBUG] Deleting NetworkFirewall Resource Policy for resource: %s", d.Id())

	input := &networkfirewall.DeleteResourcePolicyInput{
		ResourceArn: aws.String(d.Id()),
	}

	_, err := conn.DeleteResourcePolicyWithContext(ctx, input)

	if err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting NetworkFirewall Resource Policy (for resource: %s): %w", d.Id(), err))
	}

	return nil
}
