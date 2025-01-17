package backup

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/flex"
	tfiam "github.com/nij4t/terraform-provider-aws/internal/service/iam"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceSelection() *schema.Resource {
	return &schema.Resource{
		Create: resourceSelectionCreate,
		Read:   resourceSelectionRead,
		Delete: resourceSelectionDelete,
		Importer: &schema.ResourceImporter{
			State: resourceSelectionImportState,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 50),
					validation.StringMatch(regexp.MustCompile(`^[a-zA-Z0-9\-\_\.]+$`), "must contain only alphanumeric, hyphen, underscore, and period characters"),
				),
			},
			"plan_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"iam_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: verify.ValidARN,
			},
			"selection_tag": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								backup.ConditionTypeStringequals,
							}, false),
						},
						"key": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
						"value": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
						},
					},
				},
			},
			"resources": {
				Type:     schema.TypeSet,
				Optional: true,
				ForceNew: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func resourceSelectionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	selection := &backup.Selection{
		IamRoleArn:    aws.String(d.Get("iam_role_arn").(string)),
		ListOfTags:    expandBackupConditionTags(d.Get("selection_tag").(*schema.Set).List()),
		Resources:     flex.ExpandStringSet(d.Get("resources").(*schema.Set)),
		SelectionName: aws.String(d.Get("name").(string)),
	}

	input := &backup.CreateBackupSelectionInput{
		BackupPlanId:    aws.String(d.Get("plan_id").(string)),
		BackupSelection: selection,
	}

	// Retry for IAM eventual consistency
	var output *backup.CreateBackupSelectionOutput
	err := resource.Retry(tfiam.PropagationTimeout, func() *resource.RetryError {
		var err error
		output, err = conn.CreateBackupSelection(input)

		// Retry on the following error:
		// InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX cannot be assumed by AWS Backup
		if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "cannot be assumed") {
			log.Printf("[DEBUG] Received %s, retrying create backup selection.", err)
			return resource.RetryableError(err)
		}

		// Retry on the following error:
		// InvalidParameterValueException: IAM Role arn:aws:iam::123456789012:role/XXX is not authorized to call tag:GetResources
		if tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "is not authorized to call") {
			log.Printf("[DEBUG] Received %s, retrying create backup selection.", err)
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		output, err = conn.CreateBackupSelection(input)
	}

	if err != nil {
		return fmt.Errorf("error creating Backup Selection: %s", err)
	}

	d.SetId(aws.StringValue(output.SelectionId))

	return resourceSelectionRead(d, meta)
}

func resourceSelectionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	input := &backup.GetBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	}

	var resp *backup.GetBackupSelectionOutput

	err := resource.Retry(propagationTimeout, func() *resource.RetryError {
		var err error

		resp, err = conn.GetBackupSelection(input)

		if d.IsNewResource() && tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
			return resource.RetryableError(err)
		}

		if d.IsNewResource() && tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "Cannot find Backup plan") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		resp, err = conn.GetBackupSelection(input)
	}

	if !d.IsNewResource() && tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
		log.Printf("[WARN] Backup Selection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if !d.IsNewResource() && tfawserr.ErrMessageContains(err, backup.ErrCodeInvalidParameterValueException, "Cannot find Backup plan") {
		log.Printf("[WARN] Backup Selection (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading Backup Selection (%s): %w", d.Id(), err)
	}

	if resp == nil {
		return fmt.Errorf("error reading Backup Selection (%s): empty response", d.Id())
	}

	d.Set("plan_id", resp.BackupPlanId)
	d.Set("name", resp.BackupSelection.SelectionName)
	d.Set("iam_role_arn", resp.BackupSelection.IamRoleArn)

	if resp.BackupSelection.ListOfTags != nil {
		tags := make([]map[string]interface{}, 0)

		for _, r := range resp.BackupSelection.ListOfTags {
			m := make(map[string]interface{})

			m["type"] = aws.StringValue(r.ConditionType)
			m["key"] = aws.StringValue(r.ConditionKey)
			m["value"] = aws.StringValue(r.ConditionValue)

			tags = append(tags, m)
		}

		if err := d.Set("selection_tag", tags); err != nil {
			return fmt.Errorf("error setting selection tag: %s", err)
		}
	}
	if resp.BackupSelection.Resources != nil {
		if err := d.Set("resources", aws.StringValueSlice(resp.BackupSelection.Resources)); err != nil {
			return fmt.Errorf("error setting resources: %s", err)
		}
	}

	return nil
}

func resourceSelectionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).BackupConn

	input := &backup.DeleteBackupSelectionInput{
		BackupPlanId: aws.String(d.Get("plan_id").(string)),
		SelectionId:  aws.String(d.Id()),
	}

	_, err := conn.DeleteBackupSelection(input)
	if err != nil {
		return fmt.Errorf("error deleting Backup Selection: %s", err)
	}

	return nil
}

func resourceSelectionImportState(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	idParts := strings.Split(d.Id(), "|")
	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		return nil, fmt.Errorf("unexpected format of ID (%q), expected <plan-id>|<selection-id>", d.Id())
	}

	planID := idParts[0]
	selectionID := idParts[1]

	d.Set("plan_id", planID)
	d.SetId(selectionID)

	return []*schema.ResourceData{d}, nil
}

func expandBackupConditionTags(tagList []interface{}) []*backup.Condition {
	conditions := []*backup.Condition{}

	for _, i := range tagList {
		item := i.(map[string]interface{})
		tag := &backup.Condition{}

		tag.ConditionType = aws.String(item["type"].(string))
		tag.ConditionKey = aws.String(item["key"].(string))
		tag.ConditionValue = aws.String(item["value"].(string))

		conditions = append(conditions, tag)
	}

	return conditions
}
