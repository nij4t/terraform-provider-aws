package ses

import (
	"fmt"
	"log"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func ResourceDomainIdentityVerification() *schema.Resource {
	return &schema.Resource{
		Create: resourceDomainIdentityVerificationCreate,
		Read:   resourceDomainIdentityVerificationRead,
		Delete: resourceDomainIdentityVerificationDelete,

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"domain": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringDoesNotMatch(regexp.MustCompile(`\.$`), "cannot end with a period"),
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},
	}
}

func getIdentityVerificationAttributes(conn *ses.SES, domainName string) (*ses.IdentityVerificationAttributes, error) {
	input := &ses.GetIdentityVerificationAttributesInput{
		Identities: []*string{
			aws.String(domainName),
		},
	}

	response, err := conn.GetIdentityVerificationAttributes(input)
	if err != nil {
		return nil, fmt.Errorf("Error getting identity verification attributes: %s", err)
	}

	return response.VerificationAttributes[domainName], nil
}

func resourceDomainIdentityVerificationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn
	domainName := d.Get("domain").(string)
	err := resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		att, err := getIdentityVerificationAttributes(conn, domainName)
		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error getting identity verification attributes: %s", err))
		}

		if att == nil {
			return resource.NonRetryableError(fmt.Errorf("SES Domain Identity %s not found in AWS", domainName))
		}

		if aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
			return resource.RetryableError(fmt.Errorf("Expected domain verification Success, but was in state %s", aws.StringValue(att.VerificationStatus)))
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		var att *ses.IdentityVerificationAttributes
		att, err = getIdentityVerificationAttributes(conn, domainName)

		if att != nil && aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
			return fmt.Errorf("Expected domain verification Success, but was in state %s", aws.StringValue(att.VerificationStatus))
		}
	}
	if err != nil {
		return fmt.Errorf("Error creating SES domain identity verification: %s", err)
	}

	log.Printf("[INFO] Domain verification successful for %s", domainName)
	d.SetId(domainName)
	return resourceDomainIdentityVerificationRead(d, meta)
}

func resourceDomainIdentityVerificationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	domainName := d.Id()
	d.Set("domain", domainName)

	att, err := getIdentityVerificationAttributes(conn, domainName)
	if err != nil {
		log.Printf("[WARN] Error fetching identity verification attributes for %s: %s", d.Id(), err)
		return err
	}

	if att == nil {
		log.Printf("[WARN] Domain not listed in response when fetching verification attributes for %s", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(att.VerificationStatus) != ses.VerificationStatusSuccess {
		log.Printf("[WARN] Expected domain verification Success, but was %s, tainting verification", aws.StringValue(att.VerificationStatus))
		d.SetId("")
		return nil
	}

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("identity/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceDomainIdentityVerificationDelete(d *schema.ResourceData, meta interface{}) error {
	// No need to do anything, domain identity will be deleted when aws_ses_domain_identity is deleted
	return nil
}
