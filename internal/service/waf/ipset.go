package waf

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

// WAF requires UpdateIPSet operations be split into batches of 1000 Updates
const ipSetUpdatesLimit = 1000

func ResourceIPSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPSetCreate,
		Read:   resourceIPSetRead,
		Update: resourceIPSetUpdate,
		Delete: resourceIPSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"ip_set_descriptors": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								waf.IPSetDescriptorTypeIpv4,
								waf.IPSetDescriptorTypeIpv6,
							}, false),
						},
						"value": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validation.IsCIDR,
						},
					},
				},
			},
		},
	}
}

func resourceIPSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn

	wr := NewRetryer(conn)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateIPSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}
		return conn.CreateIPSet(params)
	})
	if err != nil {
		return err
	}
	resp := out.(*waf.CreateIPSetOutput)
	d.SetId(aws.StringValue(resp.IPSet.IPSetId))

	if v, ok := d.GetOk("ip_set_descriptors"); ok && v.(*schema.Set).Len() > 0 {

		err := updateWafIpSetDescriptors(d.Id(), nil, v.(*schema.Set).List(), conn)
		if err != nil {
			return fmt.Errorf("Error Setting IP Descriptors: %s", err)
		}
	}

	return resourceIPSetRead(d, meta)
}

func resourceIPSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn

	params := &waf.GetIPSetInput{
		IPSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetIPSet(params)
	if err != nil {
		if tfawserr.ErrMessageContains(err, waf.ErrCodeNonexistentItemException, "") {
			log.Printf("[WARN] WAF IPSet (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}

		return err
	}

	var descriptors []map[string]interface{}

	for _, descriptor := range resp.IPSet.IPSetDescriptors {
		d := map[string]interface{}{
			"type":  aws.StringValue(descriptor.Type),
			"value": aws.StringValue(descriptor.Value),
		}
		descriptors = append(descriptors, d)
	}

	d.Set("ip_set_descriptors", descriptors)

	d.Set("name", resp.IPSet.Name)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "waf",
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("ipset/%s", d.Id()),
	}
	d.Set("arn", arn.String())

	return nil
}

func resourceIPSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn

	if d.HasChange("ip_set_descriptors") {
		o, n := d.GetChange("ip_set_descriptors")
		oldD, newD := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateWafIpSetDescriptors(d.Id(), oldD, newD, conn)
		if err != nil {
			return fmt.Errorf("Error Updating WAF IPSet: %s", err)
		}
	}

	return resourceIPSetRead(d, meta)
}

func resourceIPSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFConn

	oldDescriptors := d.Get("ip_set_descriptors").(*schema.Set).List()

	if len(oldDescriptors) > 0 {
		err := updateWafIpSetDescriptors(d.Id(), oldDescriptors, nil, conn)
		if err != nil {
			return fmt.Errorf("Error Deleting IPSetDescriptors: %s", err)
		}
	}

	wr := NewRetryer(conn)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteIPSetInput{
			ChangeToken: token,
			IPSetId:     aws.String(d.Id()),
		}
		log.Printf("[INFO] Deleting WAF IPSet")
		return conn.DeleteIPSet(req)
	})
	if err != nil {
		return fmt.Errorf("Error Deleting WAF IPSet: %s", err)
	}

	return nil
}

func updateWafIpSetDescriptors(id string, oldD, newD []interface{}, conn *waf.WAF) error {
	for _, ipSetUpdates := range DiffIPSetDescriptors(oldD, newD) {
		wr := NewRetryer(conn)
		_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
			req := &waf.UpdateIPSetInput{
				ChangeToken: token,
				IPSetId:     aws.String(id),
				Updates:     ipSetUpdates,
			}
			log.Printf("[INFO] Updating IPSet descriptors: %s", req)
			return conn.UpdateIPSet(req)
		})
		if err != nil {
			return fmt.Errorf("Error Updating WAF IPSet: %s", err)
		}
	}

	return nil
}

func DiffIPSetDescriptors(oldD, newD []interface{}) [][]*waf.IPSetUpdate {
	updates := make([]*waf.IPSetUpdate, 0, ipSetUpdatesLimit)
	updatesBatches := make([][]*waf.IPSetUpdate, 0)

	for _, od := range oldD {
		descriptor := od.(map[string]interface{})

		if idx, contains := sliceContainsMap(newD, descriptor); contains {
			newD = append(newD[:idx], newD[idx+1:]...)
			continue
		}

		if len(updates) == ipSetUpdatesLimit {
			updatesBatches = append(updatesBatches, updates)
			updates = make([]*waf.IPSetUpdate, 0, ipSetUpdatesLimit)
		}

		updates = append(updates, &waf.IPSetUpdate{
			Action: aws.String(waf.ChangeActionDelete),
			IPSetDescriptor: &waf.IPSetDescriptor{
				Type:  aws.String(descriptor["type"].(string)),
				Value: aws.String(descriptor["value"].(string)),
			},
		})
	}

	for _, nd := range newD {
		descriptor := nd.(map[string]interface{})

		if len(updates) == ipSetUpdatesLimit {
			updatesBatches = append(updatesBatches, updates)
			updates = make([]*waf.IPSetUpdate, 0, ipSetUpdatesLimit)
		}

		updates = append(updates, &waf.IPSetUpdate{
			Action: aws.String(waf.ChangeActionInsert),
			IPSetDescriptor: &waf.IPSetDescriptor{
				Type:  aws.String(descriptor["type"].(string)),
				Value: aws.String(descriptor["value"].(string)),
			},
		})
	}
	updatesBatches = append(updatesBatches, updates)
	return updatesBatches
}
