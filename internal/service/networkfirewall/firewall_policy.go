package networkfirewall

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/flex"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceFirewallPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceFirewallPolicyCreate,
		ReadContext:   resourceFirewallPolicyRead,
		UpdateContext: resourceFirewallPolicyUpdate,
		DeleteContext: resourceFirewallPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"firewall_policy": {
				Type:     schema.TypeList,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"stateful_rule_group_reference": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"resource_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
						"stateless_custom_action": customActionSchema(),
						"stateless_default_actions": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"stateless_fragment_default_actions": {
							Type:     schema.TypeSet,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
						"stateless_rule_group_reference": {
							Type:     schema.TypeSet,
							Optional: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"priority": {
										Type:         schema.TypeInt,
										Required:     true,
										ValidateFunc: validation.IntAtLeast(1),
									},
									"resource_arn": {
										Type:         schema.TypeString,
										Required:     true,
										ValidateFunc: verify.ValidARN,
									},
								},
							},
						},
					},
				},
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"update_token": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))
	name := d.Get("name").(string)
	input := &networkfirewall.CreateFirewallPolicyInput{
		FirewallPolicy:     expandNetworkFirewallFirewallPolicy(d.Get("firewall_policy").([]interface{})),
		FirewallPolicyName: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if len(tags) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating NetworkFirewall Firewall Policy %s", name)

	output, err := conn.CreateFirewallPolicyWithContext(ctx, input)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Firewall Policy (%s): %w", name, err))
	}
	if output == nil || output.FirewallPolicyResponse == nil {
		return diag.FromErr(fmt.Errorf("error creating NetworkFirewall Firewall Policy (%s): empty output", name))
	}

	d.SetId(aws.StringValue(output.FirewallPolicyResponse.FirewallPolicyArn))

	return resourceFirewallPolicyRead(ctx, d, meta)
}

func resourceFirewallPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	log.Printf("[DEBUG] Reading NetworkFirewall Firewall Policy %s", d.Id())

	output, err := FindFirewallPolicy(ctx, conn, d.Id())
	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] NetworkFirewall Firewall Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): %w", d.Id(), err))
	}

	if output == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): empty output", d.Id()))
	}
	if output.FirewallPolicyResponse == nil {
		return diag.FromErr(fmt.Errorf("error reading NetworkFirewall Firewall Policy (%s): empty output.FirewallPolicyResponse", d.Id()))
	}

	resp := output.FirewallPolicyResponse
	policy := output.FirewallPolicy

	d.Set("arn", resp.FirewallPolicyArn)
	d.Set("description", resp.Description)
	d.Set("name", resp.FirewallPolicyName)
	d.Set("update_token", output.UpdateToken)

	if err := d.Set("firewall_policy", flattenNetworkFirewallFirewallPolicy(policy)); err != nil {
		return diag.FromErr(fmt.Errorf("error setting firewall_policy: %w", err))
	}

	tags := KeyValueTags(resp.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags: %w", err))
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return diag.FromErr(fmt.Errorf("error setting tags_all: %w", err))
	}

	return nil
}

func resourceFirewallPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn
	arn := d.Id()

	log.Printf("[DEBUG] Updating NetworkFirewall Firewall Policy %s", arn)

	if d.HasChanges("description", "firewall_policy") {
		input := &networkfirewall.UpdateFirewallPolicyInput{
			FirewallPolicy:    expandNetworkFirewallFirewallPolicy(d.Get("firewall_policy").([]interface{})),
			FirewallPolicyArn: aws.String(arn),
			UpdateToken:       aws.String(d.Get("update_token").(string)),
		}
		// Only pass non-empty description values, else API request returns an InternalServiceError
		if v, ok := d.GetOk("description"); ok {
			input.Description = aws.String(v.(string))
		}
		_, err := conn.UpdateFirewallPolicyWithContext(ctx, input)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall Policy (%s) firewall_policy: %w", arn, err))
		}
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, arn, o, n); err != nil {
			return diag.FromErr(fmt.Errorf("error updating NetworkFirewall Firewall Policy (%s) tags: %w", arn, err))
		}
	}

	return resourceFirewallPolicyRead(ctx, d, meta)
}

func resourceFirewallPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*conns.AWSClient).NetworkFirewallConn

	log.Printf("[DEBUG] Deleting NetworkFirewall Firewall Policy %s", d.Id())

	input := &networkfirewall.DeleteFirewallPolicyInput{
		FirewallPolicyArn: aws.String(d.Id()),
	}

	err := resource.RetryContext(ctx, firewallPolicyTimeout, func() *resource.RetryError {
		_, err := conn.DeleteFirewallPolicyWithContext(ctx, input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, networkfirewall.ErrCodeInvalidOperationException, "Unable to delete the object because it is still in use") {
				return resource.RetryableError(err)
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteFirewallPolicyWithContext(ctx, input)
	}

	if err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error deleting NetworkFirewall Firewall Policy (%s): %w", d.Id(), err))
	}

	if _, err := waitFirewallPolicyDeleted(ctx, conn, d.Id()); err != nil {
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			return nil
		}
		return diag.FromErr(fmt.Errorf("error waiting for NetworkFirewall Firewall Policy (%s) to delete: %w", d.Id(), err))
	}

	return nil
}

func expandNetworkFirewallStatefulRuleGroupReferences(l []interface{}) []*networkfirewall.StatefulRuleGroupReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	references := make([]*networkfirewall.StatefulRuleGroupReference, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		reference := &networkfirewall.StatefulRuleGroupReference{}
		if v, ok := tfMap["resource_arn"].(string); ok && v != "" {
			reference.ResourceArn = aws.String(v)
		}
		references = append(references, reference)
	}
	return references
}

func expandNetworkFirewallStatelessRuleGroupReferences(l []interface{}) []*networkfirewall.StatelessRuleGroupReference {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	references := make([]*networkfirewall.StatelessRuleGroupReference, 0, len(l))
	for _, tfMapRaw := range l {
		tfMap, ok := tfMapRaw.(map[string]interface{})
		if !ok {
			continue
		}
		reference := &networkfirewall.StatelessRuleGroupReference{}
		if v, ok := tfMap["priority"].(int); ok && v > 0 {
			reference.Priority = aws.Int64(int64(v))
		}
		if v, ok := tfMap["resource_arn"].(string); ok && v != "" {
			reference.ResourceArn = aws.String(v)
		}
		references = append(references, reference)
	}
	return references
}

func expandNetworkFirewallFirewallPolicy(l []interface{}) *networkfirewall.FirewallPolicy {
	if len(l) == 0 || l[0] == nil {
		return nil
	}
	lRaw := l[0].(map[string]interface{})
	policy := &networkfirewall.FirewallPolicy{
		StatelessDefaultActions:         flex.ExpandStringSet(lRaw["stateless_default_actions"].(*schema.Set)),
		StatelessFragmentDefaultActions: flex.ExpandStringSet(lRaw["stateless_fragment_default_actions"].(*schema.Set)),
	}

	if v, ok := lRaw["stateful_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatefulRuleGroupReferences = expandNetworkFirewallStatefulRuleGroupReferences(v.List())
	}

	if v, ok := lRaw["stateless_custom_action"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatelessCustomActions = expandNetworkFirewallCustomActions(v.List())
	}

	if v, ok := lRaw["stateless_rule_group_reference"].(*schema.Set); ok && v.Len() > 0 {
		policy.StatelessRuleGroupReferences = expandNetworkFirewallStatelessRuleGroupReferences(v.List())
	}

	return policy
}

func flattenNetworkFirewallFirewallPolicy(policy *networkfirewall.FirewallPolicy) []interface{} {
	if policy == nil {
		return []interface{}{}
	}
	p := map[string]interface{}{}
	if policy.StatefulRuleGroupReferences != nil {
		p["stateful_rule_group_reference"] = flattenNetworkFirewallPolicyStatefulRuleGroupReference(policy.StatefulRuleGroupReferences)
	}
	if policy.StatelessCustomActions != nil {
		p["stateless_custom_action"] = flattenNetworkFirewallCustomActions(policy.StatelessCustomActions)
	}
	if policy.StatelessDefaultActions != nil {
		p["stateless_default_actions"] = flex.FlattenStringSet(policy.StatelessDefaultActions)
	}
	if policy.StatelessFragmentDefaultActions != nil {
		p["stateless_fragment_default_actions"] = flex.FlattenStringSet(policy.StatelessFragmentDefaultActions)
	}
	if policy.StatelessRuleGroupReferences != nil {
		p["stateless_rule_group_reference"] = flattenNetworkFirewallPolicyStatelessRuleGroupReference(policy.StatelessRuleGroupReferences)
	}

	return []interface{}{p}
}

func flattenNetworkFirewallPolicyStatefulRuleGroupReference(l []*networkfirewall.StatefulRuleGroupReference) []interface{} {
	references := make([]interface{}, 0, len(l))
	for _, ref := range l {
		reference := map[string]interface{}{
			"resource_arn": aws.StringValue(ref.ResourceArn),
		}
		references = append(references, reference)
	}

	return references
}

func flattenNetworkFirewallPolicyStatelessRuleGroupReference(l []*networkfirewall.StatelessRuleGroupReference) []interface{} {
	references := make([]interface{}, 0, len(l))
	for _, ref := range l {
		reference := map[string]interface{}{
			"priority":     int(aws.Int64Value(ref.Priority)),
			"resource_arn": aws.StringValue(ref.ResourceArn),
		}
		references = append(references, reference)
	}
	return references
}
