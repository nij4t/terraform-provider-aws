package wafregional

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfwaf "github.com/nij4t/terraform-provider-aws/internal/service/waf"
)

func ResourceGeoMatchSet() *schema.Resource {
	return &schema.Resource{
		Create: resourceGeoMatchSetCreate,
		Read:   resourceGeoMatchSetRead,
		Update: resourceGeoMatchSetUpdate,
		Delete: resourceGeoMatchSetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"geo_match_constraint": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func resourceGeoMatchSetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	log.Printf("[INFO] Creating WAF Regional Geo Match Set: %s", d.Get("name").(string))

	wr := NewRetryer(conn, region)
	out, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		params := &waf.CreateGeoMatchSetInput{
			ChangeToken: token,
			Name:        aws.String(d.Get("name").(string)),
		}

		return conn.CreateGeoMatchSet(params)
	})
	if err != nil {
		return fmt.Errorf("Failed creating WAF Regional Geo Match Set: %s", err)
	}
	resp := out.(*waf.CreateGeoMatchSetOutput)

	d.SetId(aws.StringValue(resp.GeoMatchSet.GeoMatchSetId))

	return resourceGeoMatchSetUpdate(d, meta)
}

func resourceGeoMatchSetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	log.Printf("[INFO] Reading WAF Regional Geo Match Set: %s", d.Get("name").(string))
	params := &waf.GetGeoMatchSetInput{
		GeoMatchSetId: aws.String(d.Id()),
	}

	resp, err := conn.GetGeoMatchSet(params)

	if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		log.Printf("[WARN] WAF WAF Regional Geo Match Set (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting WAF Regional Geo Match Set (%s): %s", d.Id(), err)
	}

	d.Set("name", resp.GeoMatchSet.Name)
	d.Set("geo_match_constraint", tfwaf.FlattenGeoMatchConstraint(resp.GeoMatchSet.GeoMatchConstraints))

	return nil
}

func resourceGeoMatchSetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	if d.HasChange("geo_match_constraint") {
		o, n := d.GetChange("geo_match_constraint")
		oldConstraints, newConstraints := o.(*schema.Set).List(), n.(*schema.Set).List()

		err := updateGeoMatchSetResourceWR(d.Id(), oldConstraints, newConstraints, conn, region)
		if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
			log.Printf("[WARN] WAF WAF Regional Geo Match Set (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		if err != nil {
			return fmt.Errorf("Failed updating WAF Regional Geo Match Set(%s): %s", d.Id(), err)
		}
	}

	return resourceGeoMatchSetRead(d, meta)
}

func resourceGeoMatchSetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).WAFRegionalConn
	region := meta.(*conns.AWSClient).Region

	oldConstraints := d.Get("geo_match_constraint").(*schema.Set).List()
	if len(oldConstraints) > 0 {
		noConstraints := []interface{}{}
		err := updateGeoMatchSetResourceWR(d.Id(), oldConstraints, noConstraints, conn, region)
		if err != nil {
			return fmt.Errorf("Error updating WAF Regional Geo Match Constraint: %s", err)
		}
	}

	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.DeleteGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(d.Id()),
		}

		return conn.DeleteGeoMatchSet(req)
	})
	if tfawserr.ErrMessageContains(err, wafregional.ErrCodeWAFNonexistentItemException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Failed deleting WAF Regional Geo Match Set(%s): %s", d.Id(), err)
	}

	return nil
}

func updateGeoMatchSetResourceWR(id string, oldConstraints, newConstraints []interface{}, conn *wafregional.WAFRegional, region string) error {
	wr := NewRetryer(conn, region)
	_, err := wr.RetryWithToken(func(token *string) (interface{}, error) {
		req := &waf.UpdateGeoMatchSetInput{
			ChangeToken:   token,
			GeoMatchSetId: aws.String(id),
			Updates:       tfwaf.DiffGeoMatchSetConstraints(oldConstraints, newConstraints),
		}

		log.Printf("[INFO] Updating WAF Regional Geo Match Set constraints: %s", req)
		return conn.UpdateGeoMatchSet(req)
	})
	if err != nil {
		return fmt.Errorf("Failed updating WAF Regional Geo Match Set: %s", err)
	}

	return nil
}
