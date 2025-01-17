package ses

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceTemplate() *schema.Resource {
	return &schema.Resource{
		Create: resourceTemplateCreate,
		Read:   resourceTemplateRead,
		Update: resourceTemplateUpdate,
		Delete: resourceTemplateDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"html": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512000),
			},
			"subject": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"text": {
				Type:         schema.TypeString,
				Optional:     true,
				ValidateFunc: validation.StringLenBetween(0, 512000),
			},
		},
	}
}
func resourceTemplateCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	templateName := d.Get("name").(string)

	template := ses.Template{
		TemplateName: aws.String(templateName),
	}

	if v, ok := d.GetOk("html"); ok {
		template.HtmlPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subject"); ok {
		template.SubjectPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("text"); ok {
		template.TextPart = aws.String(v.(string))
	}

	input := ses.CreateTemplateInput{
		Template: &template,
	}

	log.Printf("[DEBUG] Creating SES template: %#v", input)
	_, err := conn.CreateTemplate(&input)
	if err != nil {
		return fmt.Errorf("Creating SES template failed: %s", err.Error())
	}
	d.SetId(templateName)

	return resourceTemplateRead(d, meta)
}

func resourceTemplateRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn
	input := ses.GetTemplateInput{
		TemplateName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading SES template: %#v", input)
	gto, err := conn.GetTemplate(&input)
	if err != nil {
		if tfawserr.ErrMessageContains(err, ses.ErrCodeTemplateDoesNotExistException, "") {
			log.Printf("[WARN] SES template %q not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return fmt.Errorf("Reading SES template '%s' failed: %s", aws.StringValue(input.TemplateName), err.Error())
	}

	d.Set("html", gto.Template.HtmlPart)
	d.Set("name", gto.Template.TemplateName)
	d.Set("subject", gto.Template.SubjectPart)
	d.Set("text", gto.Template.TextPart)

	arn := arn.ARN{
		Partition: meta.(*conns.AWSClient).Partition,
		Service:   "ses",
		Region:    meta.(*conns.AWSClient).Region,
		AccountID: meta.(*conns.AWSClient).AccountID,
		Resource:  fmt.Sprintf("template/%s", d.Id()),
	}.String()
	d.Set("arn", arn)

	return nil
}

func resourceTemplateUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn

	templateName := d.Id()

	template := ses.Template{
		TemplateName: aws.String(templateName),
	}

	if v, ok := d.GetOk("html"); ok {
		template.HtmlPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("subject"); ok {
		template.SubjectPart = aws.String(v.(string))
	}

	if v, ok := d.GetOk("text"); ok {
		template.TextPart = aws.String(v.(string))
	}

	input := ses.UpdateTemplateInput{
		Template: &template,
	}

	log.Printf("[DEBUG] Update SES template: %#v", input)
	_, err := conn.UpdateTemplate(&input)
	if err != nil {
		return fmt.Errorf("Updating SES template '%s' failed: %s", templateName, err.Error())
	}

	return resourceTemplateRead(d, meta)
}

func resourceTemplateDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SESConn
	input := ses.DeleteTemplateInput{
		TemplateName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Delete SES template: %#v", input)
	_, err := conn.DeleteTemplate(&input)
	if err != nil {
		return fmt.Errorf("Deleting SES template '%s' failed: %s", *input.TemplateName, err.Error())
	}
	return nil
}
