package route53resolver

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53resolver"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceFirewallRuleGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceFirewallRuleGroupCreate,
		Read:   resourceFirewallRuleGroupRead,
		Update: resourceFirewallRuleGroupUpdate,
		Delete: resourceFirewallRuleGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validResolverName,
			},

			"owner_id": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"share_status": {
				Type:     schema.TypeString,
				Computed: true,
			},

			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
		},

		CustomizeDiff: verify.SetTagsDiff,
	}
}

func resourceFirewallRuleGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &route53resolver.CreateFirewallRuleGroupInput{
		CreatorRequestId: aws.String(resource.PrefixedUniqueId("tf-r53-resolver-firewall-rule-group-")),
		Name:             aws.String(d.Get("name").(string)),
	}
	if v, ok := d.GetOk("tags"); ok && len(v.(map[string]interface{})) > 0 {
		input.Tags = Tags(tags.IgnoreAWS())
	}

	log.Printf("[DEBUG] Creating Route 53 Resolver DNS Firewall rule group: %#v", input)
	output, err := conn.CreateFirewallRuleGroup(input)
	if err != nil {
		return fmt.Errorf("error creating Route 53 Resolver DNS Firewall rule group: %w", err)
	}

	d.SetId(aws.StringValue(output.FirewallRuleGroup.Id))

	return resourceFirewallRuleGroupRead(d, meta)
}

func resourceFirewallRuleGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	ruleGroup, err := FindFirewallRuleGroupByID(conn, d.Id())

	if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		log.Printf("[WARN] Route53 Resolver DNS Firewall rule group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Route 53 Resolver DNS Firewall rule group (%s): %w", d.Id(), err)
	}

	if ruleGroup == nil {
		log.Printf("[WARN] Route 53 Resolver DNS Firewall rule group (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	arn := aws.StringValue(ruleGroup.Arn)
	d.Set("arn", arn)
	d.Set("id", ruleGroup.Id)
	d.Set("name", ruleGroup.Name)
	d.Set("owner_id", ruleGroup.OwnerId)
	d.Set("share_status", ruleGroup.ShareStatus)

	tags, err := ListTags(conn, arn)
	if err != nil {
		return fmt.Errorf("error listing tags for Route53 Resolver DNS Firewall rule group (%s): %w", arn, err)
	}

	tags = tags.IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceFirewallRuleGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")
		if err := UpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
			return fmt.Errorf("error updating Route53 Resolver DNS Firewall rule group (%s) tags: %w", d.Get("arn").(string), err)
		}
	}

	return resourceFirewallRuleGroupRead(d, meta)
}

func resourceFirewallRuleGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).Route53ResolverConn

	_, err := conn.DeleteFirewallRuleGroup(&route53resolver.DeleteFirewallRuleGroupInput{
		FirewallRuleGroupId: aws.String(d.Id()),
	})

	if tfawserr.ErrMessageContains(err, route53resolver.ErrCodeResourceNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting Route 53 Resolver DNS Firewall rule group (%s): %w", d.Id(), err)
	}

	return nil
}
