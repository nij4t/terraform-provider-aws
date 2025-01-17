package apigateway

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func ResourceGatewayResponse() *schema.Resource {
	return &schema.Resource{
		Create: resourceGatewayResponsePut,
		Read:   resourceGatewayResponseRead,
		Update: resourceGatewayResponsePut,
		Delete: resourceGatewayResponseDelete,
		Importer: &schema.ResourceImporter{
			State: func(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
				idParts := strings.Split(d.Id(), "/")
				if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
					return nil, fmt.Errorf("Unexpected format of ID (%q), expected REST-API-ID/RESPONSE-TYPE", d.Id())
				}
				restApiID := idParts[0]
				responseType := idParts[1]
				d.Set("response_type", responseType)
				d.Set("rest_api_id", restApiID)
				d.SetId(fmt.Sprintf("aggr-%s-%s", restApiID, responseType))
				return []*schema.ResourceData{d}, nil
			},
		},

		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"response_type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"status_code": {
				Type:     schema.TypeString,
				Optional: true,
			},

			"response_templates": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},

			"response_parameters": {
				Type:     schema.TypeMap,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Optional: true,
			},
		},
	}
}

func resourceGatewayResponsePut(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	templates := make(map[string]string)
	if kv, ok := d.GetOk("response_templates"); ok {
		for k, v := range kv.(map[string]interface{}) {
			templates[k] = v.(string)
		}
	}

	parameters := make(map[string]string)
	if kv, ok := d.GetOk("response_parameters"); ok {
		for k, v := range kv.(map[string]interface{}) {
			parameters[k] = v.(string)
		}
	}

	input := apigateway.PutGatewayResponseInput{
		RestApiId:          aws.String(d.Get("rest_api_id").(string)),
		ResponseType:       aws.String(d.Get("response_type").(string)),
		ResponseTemplates:  aws.StringMap(templates),
		ResponseParameters: aws.StringMap(parameters),
	}

	if v, ok := d.GetOk("status_code"); ok {
		input.StatusCode = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Putting API Gateway Gateway Response: %s", input)

	_, err := conn.PutGatewayResponse(&input)
	if err != nil {
		return fmt.Errorf("Error putting API Gateway Gateway Response: %s", err)
	}

	d.SetId(fmt.Sprintf("aggr-%s-%s", d.Get("rest_api_id").(string), d.Get("response_type").(string)))
	log.Printf("[DEBUG] API Gateway Gateway Response put (%q)", d.Id())

	return resourceGatewayResponseRead(d, meta)
}

func resourceGatewayResponseRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	log.Printf("[DEBUG] Reading API Gateway Gateway Response %s", d.Id())
	gatewayResponse, err := conn.GetGatewayResponse(&apigateway.GetGatewayResponseInput{
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
		ResponseType: aws.String(d.Get("response_type").(string)),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
			log.Printf("[WARN] API Gateway Gateway Response (%s) not found, removing from state", d.Id())
			d.SetId("")
			return nil
		}
		return err
	}

	log.Printf("[DEBUG] Received API Gateway Gateway Response: %s", gatewayResponse)

	d.Set("response_type", gatewayResponse.ResponseType)
	d.Set("status_code", gatewayResponse.StatusCode)
	d.Set("response_templates", aws.StringValueMap(gatewayResponse.ResponseTemplates))
	d.Set("response_parameters", aws.StringValueMap(gatewayResponse.ResponseParameters))

	return nil
}

func resourceGatewayResponseDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn
	log.Printf("[DEBUG] Deleting API Gateway Gateway Response: %s", d.Id())

	_, err := conn.DeleteGatewayResponse(&apigateway.DeleteGatewayResponseInput{
		RestApiId:    aws.String(d.Get("rest_api_id").(string)),
		ResponseType: aws.String(d.Get("response_type").(string)),
	})

	if tfawserr.ErrMessageContains(err, apigateway.ErrCodeNotFoundException, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error deleting API Gateway gateway response: %s", err)
	}
	return nil
}
