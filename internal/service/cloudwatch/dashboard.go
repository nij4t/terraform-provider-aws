package cloudwatch

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/structure"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceDashboard() *schema.Resource {
	return &schema.Resource{
		Create: resourceDashboardPut,
		Read:   resourceDashboardRead,
		Update: resourceDashboardPut,
		Delete: resourceDashboardDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		// Note that we specify both the `dashboard_body` and
		// the `dashboard_name` as being required, even though
		// according to the REST API documentation both are
		// optional: http://docs.aws.amazon.com/AmazonCloudWatch/latest/APIReference/API_PutDashboard.html#API_PutDashboard_RequestParameters
		Schema: map[string]*schema.Schema{
			"dashboard_arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"dashboard_body": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringIsJSON,
				StateFunc: func(v interface{}) string {
					json, _ := structure.NormalizeJsonString(v)
					return json
				},
				DiffSuppressFunc: verify.SuppressEquivalentJSONDiffs,
			},
			"dashboard_name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validDashboardName,
			},
		},
	}
}

func resourceDashboardRead(d *schema.ResourceData, meta interface{}) error {
	dashboardName := d.Get("dashboard_name").(string)
	log.Printf("[DEBUG] Reading CloudWatch Dashboard: %s", dashboardName)
	conn := meta.(*conns.AWSClient).CloudWatchConn

	params := cloudwatch.GetDashboardInput{
		DashboardName: aws.String(d.Id()),
	}

	resp, err := conn.GetDashboard(&params)
	if err != nil {
		if IsDashboardNotFoundErr(err) {
			log.Printf("[WARN] CloudWatch Dashboard %q not found, removing", dashboardName)
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading dashboard failed: %s", err)
	}

	d.Set("dashboard_arn", resp.DashboardArn)
	d.Set("dashboard_name", resp.DashboardName)
	d.Set("dashboard_body", resp.DashboardBody)
	return nil
}

func resourceDashboardPut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).CloudWatchConn
	params := cloudwatch.PutDashboardInput{
		DashboardBody: aws.String(d.Get("dashboard_body").(string)),
		DashboardName: aws.String(d.Get("dashboard_name").(string)),
	}

	log.Printf("[DEBUG] Putting CloudWatch Dashboard: %#v", params)

	_, err := conn.PutDashboard(&params)
	if err != nil {
		return fmt.Errorf("Putting dashboard failed: %s", err)
	}
	d.SetId(d.Get("dashboard_name").(string))
	log.Println("[INFO] CloudWatch Dashboard put finished")

	return resourceDashboardRead(d, meta)
}

func resourceDashboardDelete(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[INFO] Deleting CloudWatch Dashboard %s", d.Id())
	conn := meta.(*conns.AWSClient).CloudWatchConn
	params := cloudwatch.DeleteDashboardsInput{
		DashboardNames: []*string{aws.String(d.Id())},
	}

	if _, err := conn.DeleteDashboards(&params); err != nil {
		if IsDashboardNotFoundErr(err) {
			return nil
		}
		return fmt.Errorf("Error deleting CloudWatch Dashboard: %s", err)
	}
	log.Printf("[INFO] CloudWatch Dashboard %s deleted", d.Id())

	return nil
}

func IsDashboardNotFoundErr(err error) bool {
	return tfawserr.ErrMessageContains(
		err,
		"ResourceNotFound",
		"does not exist")

}
