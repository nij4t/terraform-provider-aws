package appsync

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceFunction() *schema.Resource {
	return &schema.Resource{
		Create: resourceFunctionCreate,
		Read:   resourceFunctionRead,
		Update: resourceFunctionUpdate,
		Delete: resourceFunctionDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"data_source": {
				Type:     schema.TypeString,
				Required: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)
					if !regexp.MustCompile(`[_A-Za-z][_0-9A-Za-z]*`).MatchString(value) {
						errors = append(errors, fmt.Errorf("%q must match [_A-Za-z][_0-9A-Za-z]*", k))
					}
					return
				},
			},
			"request_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"response_mapping_template": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"function_version": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "2018-05-29",
				ValidateFunc: validation.StringInSlice([]string{
					"2018-05-29",
				}, true),
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"function_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceFunctionCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID := d.Get("api_id").(string)

	input := &appsync.CreateFunctionInput{
		ApiId:                  aws.String(apiID),
		DataSourceName:         aws.String(d.Get("data_source").(string)),
		FunctionVersion:        aws.String(d.Get("function_version").(string)),
		Name:                   aws.String(d.Get("name").(string)),
		RequestMappingTemplate: aws.String(d.Get("request_mapping_template").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	resp, err := conn.CreateFunction(input)
	if err != nil {
		return fmt.Errorf("Error creating AppSync Function: %s", err)
	}

	d.SetId(fmt.Sprintf("%s-%s", apiID, aws.StringValue(resp.FunctionConfiguration.FunctionId)))

	return resourceFunctionRead(d, meta)
}

func resourceFunctionRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.GetFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	resp, err := conn.GetFunction(input)
	if tfawserr.ErrMessageContains(err, appsync.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] No such entity found for Appsync Function (%s)", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error getting AppSync Function %s: %s", d.Id(), err)
	}

	d.Set("api_id", apiID)
	d.Set("function_id", functionID)
	d.Set("data_source", resp.FunctionConfiguration.DataSourceName)
	d.Set("description", resp.FunctionConfiguration.Description)
	d.Set("arn", resp.FunctionConfiguration.FunctionArn)
	d.Set("function_version", resp.FunctionConfiguration.FunctionVersion)
	d.Set("name", resp.FunctionConfiguration.Name)
	d.Set("request_mapping_template", resp.FunctionConfiguration.RequestMappingTemplate)
	d.Set("response_mapping_template", resp.FunctionConfiguration.ResponseMappingTemplate)

	return nil
}

func resourceFunctionUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.UpdateFunctionInput{
		ApiId:                  aws.String(apiID),
		DataSourceName:         aws.String(d.Get("data_source").(string)),
		FunctionId:             aws.String(functionID),
		FunctionVersion:        aws.String(d.Get("function_version").(string)),
		Name:                   aws.String(d.Get("name").(string)),
		RequestMappingTemplate: aws.String(d.Get("request_mapping_template").(string)),
	}

	if v, ok := d.GetOk("description"); ok {
		input.Description = aws.String(v.(string))
	}

	if v, ok := d.GetOk("response_mapping_template"); ok {
		input.ResponseMappingTemplate = aws.String(v.(string))
	}

	_, err = conn.UpdateFunction(input)
	if tfawserr.ErrMessageContains(err, appsync.ErrCodeNotFoundException, "") {
		log.Printf("[WARN] No such entity found for Appsync Function (%s)", d.Id())
		d.SetId("")
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error updating AppSync Function %s: %s", d.Id(), err)
	}

	return resourceFunctionRead(d, meta)
}

func resourceFunctionDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).AppSyncConn

	apiID, functionID, err := DecodeFunctionID(d.Id())
	if err != nil {
		return err
	}

	input := &appsync.DeleteFunctionInput{
		ApiId:      aws.String(apiID),
		FunctionId: aws.String(functionID),
	}

	_, err = conn.DeleteFunction(input)
	if tfawserr.ErrMessageContains(err, appsync.ErrCodeNotFoundException, "") {
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting AppSync Function %s: %s", d.Id(), err)
	}

	return nil
}

func DecodeFunctionID(id string) (string, string, error) {
	idParts := strings.SplitN(id, "-", 2)
	if len(idParts) != 2 {
		return "", "", fmt.Errorf("expected ID in format ApiID-FunctionID, received: %s", id)
	}
	return idParts[0], idParts[1], nil
}
