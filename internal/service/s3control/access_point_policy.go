package s3control

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceAccessPointPolicy() *schema.Resource {
	return &schema.Resource{
		Create: resourceAccessPointPolicyCreate,
		Read:   resourceAccessPointPolicyRead,
		Update: resourceAccessPointPolicyUpdate,
		Delete: resourceAccessPointPolicyDelete,

		Importer: &schema.ResourceImporter{
			StateContext: resourceAccessPointPolicyImport,
		},

		Schema: map[string]*schema.Schema{
			"access_point_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"has_public_access_policy": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"policy": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: verify.SuppressEquivalentPolicyDiffs,
			},
		},
	}
}

func resourceAccessPointPolicyCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	resourceID, err := AccessPointCreateResourceID(d.Get("access_point_arn").(string))

	if err != nil {
		return err
	}

	accountID, name, err := AccessPointParseResourceID(resourceID)

	if err != nil {
		return err
	}

	input := &s3control.PutAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Creating S3 Access Point Policy: %s", input)
	_, err = conn.PutAccessPointPolicy(input)

	if err != nil {
		return fmt.Errorf("error creating S3 Access Point (%s) Policy: %w", resourceID, err)
	}

	d.SetId(resourceID)

	return resourceAccessPointPolicyRead(d, meta)
}

func resourceAccessPointPolicyRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, name, err := AccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	policy, status, err := FindAccessPointPolicyAndStatusByAccountIDAndName(conn, accountID, name)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] S3 Access Point Policy (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading S3 Access Point Policy (%s): %w", d.Id(), err)
	}

	d.Set("has_public_access_policy", status.IsPublic)
	d.Set("policy", policy)

	return nil
}

func resourceAccessPointPolicyUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, name, err := AccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	input := &s3control.PutAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
		Policy:    aws.String(d.Get("policy").(string)),
	}

	log.Printf("[DEBUG] Updating S3 Access Point Policy: %s", input)
	_, err = conn.PutAccessPointPolicy(input)

	if err != nil {
		return fmt.Errorf("error updating S3 Access Point Policy (%s): %w", d.Id(), err)
	}

	return resourceAccessPointPolicyRead(d, meta)
}

func resourceAccessPointPolicyDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).S3ControlConn

	accountID, name, err := AccessPointParseResourceID(d.Id())

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Deleting S3 Access Point Policy: %s", d.Id())
	_, err = conn.DeleteAccessPointPolicy(&s3control.DeleteAccessPointPolicyInput{
		AccountId: aws.String(accountID),
		Name:      aws.String(name),
	})

	if tfawserr.ErrCodeEquals(err, errCodeNoSuchAccessPoint, errCodeNoSuchAccessPointPolicy) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting S3 Access Point Policy (%s): %w", d.Id(), err)
	}

	return nil
}

func resourceAccessPointPolicyImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	resourceID, err := AccessPointCreateResourceID(d.Id())

	if err != nil {
		return nil, err
	}

	d.Set("access_point_arn", d.Id())
	d.SetId(resourceID)

	return []*schema.ResourceData{d}, nil
}
