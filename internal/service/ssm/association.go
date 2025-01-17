package ssm

import (
	"fmt"
	"log"
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceAssociation() *schema.Resource {
	//lintignore:R011
	return &schema.Resource{
		Create: resourceAssociationCreate,
		Read:   resourceAssociationRead,
		Update: resourceAssociationUpdate,
		Delete: resourceAssociationDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		MigrateState:  AssociationMigrateState,
		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"apply_only_at_cron_interval": {
				Type:     schema.TypeBool,
				Default:  false,
				Optional: true,
			},
			"association_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"association_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"instance_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"document_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"max_concurrency": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[1-9][0-9]%|[1-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
			},
			"max_errors": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringMatch(regexp.MustCompile(`^([1-9][0-9]*|[0]|[1-9][0-9]%|[0-9]%|100%)$`), "must be a valid number (e.g. 10) or percentage including the percent sign (e.g. 10%)"),
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"parameters": {
				Type:     schema.TypeMap,
				Optional: true,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
			"schedule_expression": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"output_location": {
				Type:     schema.TypeList,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"s3_bucket_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"s3_key_prefix": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"s3_region": {
							Type:         schema.TypeString,
							Optional:     true,
							ValidateFunc: validation.StringLenBetween(3, 20),
						},
					},
				},
			},
			"targets": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				MaxItems: 5,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"key": {
							Type:     schema.TypeString,
							Required: true,
						},
						"values": {
							Type:     schema.TypeList,
							Required: true,
							Elem:     &schema.Schema{Type: schema.TypeString},
						},
					},
				},
			},
			"compliance_severity": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					ssm.ComplianceSeverityUnspecified,
					ssm.ComplianceSeverityLow,
					ssm.ComplianceSeverityMedium,
					ssm.ComplianceSeverityHigh,
					ssm.ComplianceSeverityCritical,
				}, false),
			},
			"automation_target_parameter_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAssociationCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] SSM association create: %s", d.Id())

	associationInput := &ssm.CreateAssociationInput{
		Name: aws.String(d.Get("name").(string)),
	}

	if v, ok := d.GetOk("apply_only_at_cron_interval"); ok {
		associationInput.ApplyOnlyAtCronInterval = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("association_name"); ok {
		associationInput.AssociationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("instance_id"); ok {
		associationInput.InstanceId = aws.String(v.(string))
	}

	if v, ok := d.GetOk("document_version"); ok {
		associationInput.DocumentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_expression"); ok {
		associationInput.ScheduleExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		associationInput.Parameters = expandSSMDocumentParameters(v.(map[string]interface{}))
	}

	if _, ok := d.GetOk("targets"); ok {
		associationInput.Targets = expandTargets(d.Get("targets").([]interface{}))
	}

	if v, ok := d.GetOk("output_location"); ok {
		associationInput.OutputLocation = expandSSMAssociationOutputLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk("compliance_severity"); ok {
		associationInput.ComplianceSeverity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		associationInput.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		associationInput.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automation_target_parameter_name"); ok {
		associationInput.AutomationTargetParameterName = aws.String(v.(string))
	}

	resp, err := conn.CreateAssociation(associationInput)
	if err != nil {
		return fmt.Errorf("Error creating SSM association: %s", err)
	}

	if resp.AssociationDescription == nil {
		return fmt.Errorf("AssociationDescription was nil")
	}

	d.SetId(aws.StringValue(resp.AssociationDescription.AssociationId))
	d.Set("association_id", resp.AssociationDescription.AssociationId)

	return resourceAssociationRead(d, meta)
}

func resourceAssociationRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] Reading SSM Association: %s", d.Id())

	params := &ssm.DescribeAssociationInput{
		AssociationId: aws.String(d.Id()),
	}

	resp, err := conn.DescribeAssociation(params)

	if err != nil {
		if tfawserr.ErrMessageContains(err, ssm.ErrCodeAssociationDoesNotExist, "") {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Error reading SSM association: %s", err)
	}
	if resp.AssociationDescription == nil {
		return fmt.Errorf("AssociationDescription was nil")
	}

	association := resp.AssociationDescription
	d.Set("apply_only_at_cron_interval", association.ApplyOnlyAtCronInterval)
	d.Set("association_name", association.AssociationName)
	d.Set("instance_id", association.InstanceId)
	d.Set("name", association.Name)
	d.Set("association_id", association.AssociationId)
	d.Set("schedule_expression", association.ScheduleExpression)
	d.Set("document_version", association.DocumentVersion)
	d.Set("compliance_severity", association.ComplianceSeverity)
	d.Set("max_concurrency", association.MaxConcurrency)
	d.Set("max_errors", association.MaxErrors)
	d.Set("automation_target_parameter_name", association.AutomationTargetParameterName)

	if err := d.Set("parameters", flattenParameters(association.Parameters)); err != nil {
		return err
	}

	if err := d.Set("targets", flattenTargets(association.Targets)); err != nil {
		return fmt.Errorf("Error setting targets error: %#v", err)
	}

	if err := d.Set("output_location", flattenAssociationOutoutLocation(association.OutputLocation)); err != nil {
		return fmt.Errorf("Error setting output_location error: %#v", err)
	}

	return nil
}

func resourceAssociationUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] SSM Association update: %s", d.Id())

	associationInput := &ssm.UpdateAssociationInput{
		AssociationId: aws.String(d.Get("association_id").(string)),
	}

	if v, ok := d.GetOk("apply_only_at_cron_interval"); ok {
		associationInput.ApplyOnlyAtCronInterval = aws.Bool(v.(bool))
	}

	// AWS creates a new version every time the association is updated, so everything should be passed in the update.
	if v, ok := d.GetOk("association_name"); ok {
		associationInput.AssociationName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("document_version"); ok {
		associationInput.DocumentVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("schedule_expression"); ok {
		associationInput.ScheduleExpression = aws.String(v.(string))
	}

	if v, ok := d.GetOk("parameters"); ok {
		associationInput.Parameters = expandSSMDocumentParameters(v.(map[string]interface{}))
	}

	if _, ok := d.GetOk("targets"); ok {
		associationInput.Targets = expandTargets(d.Get("targets").([]interface{}))
	}

	if v, ok := d.GetOk("output_location"); ok {
		associationInput.OutputLocation = expandSSMAssociationOutputLocation(v.([]interface{}))
	}

	if v, ok := d.GetOk("compliance_severity"); ok {
		associationInput.ComplianceSeverity = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_concurrency"); ok {
		associationInput.MaxConcurrency = aws.String(v.(string))
	}

	if v, ok := d.GetOk("max_errors"); ok {
		associationInput.MaxErrors = aws.String(v.(string))
	}

	if v, ok := d.GetOk("automation_target_parameter_name"); ok {
		associationInput.AutomationTargetParameterName = aws.String(v.(string))
	}

	_, err := conn.UpdateAssociation(associationInput)
	if err != nil {
		return fmt.Errorf("Error updating SSM association: %s", err)
	}

	return resourceAssociationRead(d, meta)
}

func resourceAssociationDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSMConn

	log.Printf("[DEBUG] Deleting SSM Association: %s", d.Id())

	params := &ssm.DeleteAssociationInput{
		AssociationId: aws.String(d.Get("association_id").(string)),
	}

	_, err := conn.DeleteAssociation(params)

	if err != nil {
		return fmt.Errorf("Error deleting SSM association: %s", err)
	}

	return nil
}

func expandSSMDocumentParameters(params map[string]interface{}) map[string][]*string {
	var docParams = make(map[string][]*string)
	for k, v := range params {
		values := make([]*string, 1)
		values[0] = aws.String(v.(string))
		docParams[k] = values
	}

	return docParams
}

func expandSSMAssociationOutputLocation(config []interface{}) *ssm.InstanceAssociationOutputLocation {
	if config == nil {
		return nil
	}

	//We only allow 1 Item so we can grab the first in the list only
	locationConfig := config[0].(map[string]interface{})

	S3OutputLocation := &ssm.S3OutputLocation{
		OutputS3BucketName: aws.String(locationConfig["s3_bucket_name"].(string)),
	}

	if v, ok := locationConfig["s3_key_prefix"]; ok {
		S3OutputLocation.OutputS3KeyPrefix = aws.String(v.(string))
	}

	if v, ok := locationConfig["s3_region"].(string); ok && v != "" {
		S3OutputLocation.OutputS3Region = aws.String(v)
	}

	return &ssm.InstanceAssociationOutputLocation{
		S3Location: S3OutputLocation,
	}
}

func flattenAssociationOutoutLocation(location *ssm.InstanceAssociationOutputLocation) []map[string]interface{} {
	if location == nil {
		return nil
	}

	result := make([]map[string]interface{}, 0)
	item := make(map[string]interface{})

	item["s3_bucket_name"] = *location.S3Location.OutputS3BucketName

	if location.S3Location.OutputS3KeyPrefix != nil {
		item["s3_key_prefix"] = *location.S3Location.OutputS3KeyPrefix
	}

	if location.S3Location.OutputS3Region != nil {
		item["s3_region"] = *location.S3Location.OutputS3Region
	}

	result = append(result, item)

	return result
}
