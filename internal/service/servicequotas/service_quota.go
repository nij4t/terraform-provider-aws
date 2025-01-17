package servicequotas

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/servicequotas"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceServiceQuota() *schema.Resource {
	return &schema.Resource{
		Create: resourceServiceQuotaCreate,
		Read:   resourceServiceQuotaRead,
		Update: resourceServiceQuotaUpdate,
		Delete: schema.Noop,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"adjustable": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"default_value": {
				Type:     schema.TypeFloat,
				Computed: true,
			},
			"quota_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 128),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"quota_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"request_status": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"service_code": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z]`), "must begin with alphabetic character"),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9-]+$`), "must contain only alphanumeric and hyphen characters"),
				),
			},
			"service_name": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value": {
				Type:     schema.TypeFloat,
				Required: true,
			},
		},
	}
}

func resourceServiceQuotaCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceQuotasConn

	quotaCode := d.Get("quota_code").(string)
	serviceCode := d.Get("service_code").(string)
	value := d.Get("value").(float64)

	d.SetId(fmt.Sprintf("%s/%s", serviceCode, quotaCode))

	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	output, err := conn.GetServiceQuota(input)

	if err != nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): %s", d.Id(), err)
	}

	if output == nil || output.Quota == nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): empty result", d.Id())
	}

	if output.Quota.ErrorReason != nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): %s: %s", d.Id(), aws.StringValue(output.Quota.ErrorReason.ErrorCode), aws.StringValue(output.Quota.ErrorReason.ErrorMessage))
	}

	if output.Quota.Value == nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): empty value", d.Id())
	}

	if value > aws.Float64Value(output.Quota.Value) {
		input := &servicequotas.RequestServiceQuotaIncreaseInput{
			DesiredValue: aws.Float64(value),
			QuotaCode:    aws.String(quotaCode),
			ServiceCode:  aws.String(serviceCode),
		}

		output, err := conn.RequestServiceQuotaIncrease(input)

		if err != nil {
			return fmt.Errorf("error requesting Service Quota (%s) increase: %s", d.Id(), err)
		}

		if output == nil || output.RequestedQuota == nil {
			return fmt.Errorf("error requesting Service Quota (%s) increase: empty result", d.Id())
		}

		d.Set("request_id", output.RequestedQuota.Id)
	}

	return resourceServiceQuotaRead(d, meta)
}

func resourceServiceQuotaRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceQuotasConn

	serviceCode, quotaCode, err := resourceServiceQuotaParseID(d.Id())

	if err != nil {
		return err
	}

	input := &servicequotas.GetServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	output, err := conn.GetServiceQuota(input)

	if tfawserr.ErrMessageContains(err, servicequotas.ErrCodeNoSuchResourceException, "") {
		log.Printf("[WARN] Service Quotas Service Quota (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): %s", d.Id(), err)
	}

	if output == nil || output.Quota == nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): empty result", d.Id())
	}

	if output.Quota.ErrorReason != nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): %s: %s", d.Id(), aws.StringValue(output.Quota.ErrorReason.ErrorCode), aws.StringValue(output.Quota.ErrorReason.ErrorMessage))
	}

	if output.Quota.Value == nil {
		return fmt.Errorf("error getting Service Quotas Service Quota (%s): empty value", d.Id())
	}

	defaultInput := &servicequotas.GetAWSDefaultServiceQuotaInput{
		QuotaCode:   aws.String(quotaCode),
		ServiceCode: aws.String(serviceCode),
	}

	defaultOutput, err := conn.GetAWSDefaultServiceQuota(defaultInput)

	if err != nil {
		return fmt.Errorf("error getting Service Quotas Default Service Quota (%s): %s", d.Id(), err)
	}

	if defaultOutput == nil {
		return fmt.Errorf("error getting Service Quotas Default Service Quota (%s): empty result", d.Id())
	}

	d.Set("adjustable", output.Quota.Adjustable)
	d.Set("arn", output.Quota.QuotaArn)
	d.Set("default_value", defaultOutput.Quota.Value)
	d.Set("quota_code", output.Quota.QuotaCode)
	d.Set("quota_name", output.Quota.QuotaName)
	d.Set("service_code", output.Quota.ServiceCode)
	d.Set("service_name", output.Quota.ServiceName)
	d.Set("value", output.Quota.Value)

	requestID := d.Get("request_id").(string)

	if requestID != "" {
		input := &servicequotas.GetRequestedServiceQuotaChangeInput{
			RequestId: aws.String(requestID),
		}

		output, err := conn.GetRequestedServiceQuotaChange(input)

		if tfawserr.ErrMessageContains(err, servicequotas.ErrCodeNoSuchResourceException, "") {
			d.Set("request_id", "")
			d.Set("request_status", "")
			return nil
		}

		if err != nil {
			return fmt.Errorf("error getting Service Quotas Requested Service Quota Change (%s): %s", requestID, err)
		}

		if output == nil || output.RequestedQuota == nil {
			return fmt.Errorf("error getting Service Quotas Requested Service Quota Change (%s): empty result", requestID)
		}

		requestStatus := aws.StringValue(output.RequestedQuota.Status)
		d.Set("request_status", requestStatus)

		switch requestStatus {
		case servicequotas.RequestStatusApproved, servicequotas.RequestStatusCaseClosed, servicequotas.RequestStatusDenied:
			d.Set("request_id", "")
		case servicequotas.RequestStatusCaseOpened, servicequotas.RequestStatusPending:
			d.Set("value", output.RequestedQuota.DesiredValue)
		}
	}

	return nil
}

func resourceServiceQuotaUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ServiceQuotasConn

	value := d.Get("value").(float64)
	serviceCode, quotaCode, err := resourceServiceQuotaParseID(d.Id())

	if err != nil {
		return err
	}

	input := &servicequotas.RequestServiceQuotaIncreaseInput{
		DesiredValue: aws.Float64(value),
		QuotaCode:    aws.String(quotaCode),
		ServiceCode:  aws.String(serviceCode),
	}

	output, err := conn.RequestServiceQuotaIncrease(input)

	if err != nil {
		return fmt.Errorf("error requesting Service Quota (%s) increase: %s", d.Id(), err)
	}

	if output == nil || output.RequestedQuota == nil {
		return fmt.Errorf("error requesting Service Quota (%s) increase: empty result", d.Id())
	}

	d.Set("request_id", output.RequestedQuota.Id)

	return resourceServiceQuotaRead(d, meta)
}

func resourceServiceQuotaParseID(id string) (string, string, error) {
	parts := strings.SplitN(id, "/", 2)

	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("unexpected format of ID (%s), expected SERVICE-CODE/QUOTA-CODE", id)
	}

	return parts[0], parts[1], nil
}
