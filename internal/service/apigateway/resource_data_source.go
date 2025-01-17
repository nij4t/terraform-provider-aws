package apigateway

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func DataSourceResource() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceResourceRead,
		Schema: map[string]*schema.Schema{
			"rest_api_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			"path_part": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"parent_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func dataSourceResourceRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).APIGatewayConn

	restApiId := d.Get("rest_api_id").(string)
	target := d.Get("path").(string)
	params := &apigateway.GetResourcesInput{RestApiId: aws.String(restApiId)}

	var match *apigateway.Resource
	log.Printf("[DEBUG] Reading API Gateway Resources: %s", params)
	err := conn.GetResourcesPages(params, func(page *apigateway.GetResourcesOutput, lastPage bool) bool {
		for _, resource := range page.Items {
			if aws.StringValue(resource.Path) == target {
				match = resource
				return false
			}
		}
		return !lastPage
	})
	if err != nil {
		return fmt.Errorf("error describing API Gateway Resources: %w", err)
	}

	if match == nil {
		return fmt.Errorf("no Resources with path %q found for rest api %q", target, restApiId)
	}

	d.SetId(aws.StringValue(match.Id))
	d.Set("path_part", match.PathPart)
	d.Set("parent_id", match.ParentId)

	return nil
}
