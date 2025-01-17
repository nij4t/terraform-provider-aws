package acm

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func ResourceCertificateValidation() *schema.Resource {
	return &schema.Resource{
		Create: resourceCertificateValidationCreate,
		Read:   resourceCertificateValidationRead,
		Delete: resourceCertificateValidationDelete,

		Schema: map[string]*schema.Schema{
			"certificate_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"validation_record_fqdns": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Set:      schema.HashString,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(45 * time.Minute),
		},
	}
}

func resourceCertificateValidationCreate(d *schema.ResourceData, meta interface{}) error {
	certificate_arn := d.Get("certificate_arn").(string)

	conn := meta.(*conns.AWSClient).ACMConn
	params := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(certificate_arn),
	}

	resp, err := conn.DescribeCertificate(params)

	if err != nil {
		return fmt.Errorf("Error describing certificate: %w", err)
	}

	if resp == nil || resp.Certificate == nil {
		return fmt.Errorf("Error describing certificate: empty output")
	}

	if aws.StringValue(resp.Certificate.Type) != acm.CertificateTypeAmazonIssued {
		return fmt.Errorf("Certificate %s has type %s, no validation necessary", aws.StringValue(resp.Certificate.CertificateArn), aws.StringValue(resp.Certificate.Status))
	}

	if validation_record_fqdns, ok := d.GetOk("validation_record_fqdns"); ok {
		err := resourceCertificateCheckValidationRecords(validation_record_fqdns.(*schema.Set).List(), resp.Certificate, conn)
		if err != nil {
			return err
		}
	} else {
		log.Printf("[INFO] No validation_record_fqdns set, skipping check")
	}

	err = resource.Retry(d.Timeout(schema.TimeoutCreate), func() *resource.RetryError {
		resp, err := conn.DescribeCertificate(params)

		if err != nil {
			return resource.NonRetryableError(fmt.Errorf("Error describing certificate: %w", err))
		}

		if aws.StringValue(resp.Certificate.Status) != acm.CertificateStatusIssued {
			return resource.RetryableError(fmt.Errorf("Expected certificate to be issued but was in state %s", aws.StringValue(resp.Certificate.Status)))
		}

		log.Printf("[INFO] ACM Certificate validation for %s done, certificate was issued", certificate_arn)
		if err := resourceCertificateValidationRead(d, meta); err != nil {
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if tfresource.TimedOut(err) {
		resp, err = conn.DescribeCertificate(params)
		if aws.StringValue(resp.Certificate.Status) != acm.CertificateStatusIssued {
			return fmt.Errorf("Expected certificate to be issued but was in state %s", aws.StringValue(resp.Certificate.Status))
		}
	}
	if err != nil {
		return fmt.Errorf("Error describing created certificate: %w", err)
	}
	return nil
}

func resourceCertificateCheckValidationRecords(validationRecordFqdns []interface{}, cert *acm.CertificateDetail, conn *acm.ACM) error {
	expectedFqdns := make(map[string]*acm.DomainValidation)

	if len(cert.DomainValidationOptions) == 0 {
		input := &acm.DescribeCertificateInput{
			CertificateArn: cert.CertificateArn,
		}
		var err error
		var output *acm.DescribeCertificateOutput
		err = resource.Retry(1*time.Minute, func() *resource.RetryError {
			log.Printf("[DEBUG] Certificate domain validation options empty for %s, retrying", aws.StringValue(cert.CertificateArn))
			output, err = conn.DescribeCertificate(input)
			if err != nil {
				return resource.NonRetryableError(err)
			}
			if len(output.Certificate.DomainValidationOptions) == 0 {
				return resource.RetryableError(fmt.Errorf("Certificate domain validation options empty for %s", aws.StringValue(cert.CertificateArn)))
			}
			cert = output.Certificate
			return nil
		})
		if tfresource.TimedOut(err) {
			output, err = conn.DescribeCertificate(input)
			if err != nil {
				return fmt.Errorf("Error describing ACM certificate: %w", err)
			}
			if len(output.Certificate.DomainValidationOptions) == 0 {
				return fmt.Errorf("Certificate domain validation options empty for %s", aws.StringValue(cert.CertificateArn))
			}
		}
		if err != nil {
			return fmt.Errorf("Error checking certificate domain validation options: %w", err)
		}
		if output == nil || output.Certificate == nil {
			return fmt.Errorf("Error checking certificate domain validation options: empty output")
		}

		cert = output.Certificate
	}
	for _, v := range cert.DomainValidationOptions {
		if v.ValidationMethod != nil {
			if aws.StringValue(v.ValidationMethod) != acm.ValidationMethodDns {
				return fmt.Errorf("validation_record_fqdns is only valid for DNS validation")
			}
			if v.ResourceRecord != nil && aws.StringValue(v.ResourceRecord.Name) != "" {
				newExpectedFqdn := strings.TrimSuffix(aws.StringValue(v.ResourceRecord.Name), ".")
				expectedFqdns[newExpectedFqdn] = v
			}
		} else if len(v.ValidationEmails) > 0 {
			// ACM API sometimes is not sending ValidationMethod for EMAIL validation
			return fmt.Errorf("validation_record_fqdns is only valid for DNS validation")
		}
	}

	for _, v := range validationRecordFqdns {
		delete(expectedFqdns, strings.TrimSuffix(v.(string), "."))
	}

	if len(expectedFqdns) > 0 {
		var errors error
		for expectedFqdn, domainValidation := range expectedFqdns {
			errors = multierror.Append(errors, fmt.Errorf("missing %s DNS validation record: %s", aws.StringValue(domainValidation.DomainName), expectedFqdn))
		}
		return errors
	}

	return nil
}

func resourceCertificateValidationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ACMConn

	params := &acm.DescribeCertificateInput{
		CertificateArn: aws.String(d.Get("certificate_arn").(string)),
	}

	resp, err := conn.DescribeCertificate(params)

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, acm.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] ACM Certificate (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing ACM Certificate (%s): %w", d.Id(), err)
	}

	if resp == nil || resp.Certificate == nil {
		return fmt.Errorf("error describing ACM Certificate (%s): empty response", d.Id())
	}

	if status := aws.StringValue(resp.Certificate.Status); status != acm.CertificateStatusIssued {
		if d.IsNewResource() {
			return fmt.Errorf("ACM Certificate (%s) status not issued: %s", d.Id(), status)
		}

		log.Printf("[WARN] ACM Certificate (%s) status not issued (%s), removing from state", d.Id(), status)
		d.SetId("")
		return nil
	}

	d.SetId(aws.TimeValue(resp.Certificate.IssuedAt).String())

	return nil
}

func resourceCertificateValidationDelete(d *schema.ResourceData, meta interface{}) error {
	// No need to do anything, certificate will be deleted when acm_certificate is deleted
	return nil
}
