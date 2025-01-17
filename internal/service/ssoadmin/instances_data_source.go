package ssoadmin

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func DataSourceInstances() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceInstancesRead,

		Schema: map[string]*schema.Schema{
			"arns": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},

			"identity_store_ids": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem:     &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceInstancesRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).SSOAdminConn

	var instances []*ssoadmin.InstanceMetadata
	var arns, ids []string

	err := conn.ListInstancesPages(&ssoadmin.ListInstancesInput{}, func(page *ssoadmin.ListInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		instances = append(instances, page.Instances...)

		return !lastPage
	})

	if err != nil {
		return fmt.Errorf("error reading SSO Instances: %w", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("error reading SSO Instance: no instances found")
	}

	for _, instance := range instances {
		arns = append(arns, aws.StringValue(instance.InstanceArn))
		ids = append(ids, aws.StringValue(instance.IdentityStoreId))
	}

	d.SetId(meta.(*conns.AWSClient).Region)
	if err := d.Set("arns", arns); err != nil {
		return fmt.Errorf("error setting arns: %w", err)
	}
	if err := d.Set("identity_store_ids", ids); err != nil {
		return fmt.Errorf("error setting identity_store_ids: %w", err)
	}

	return nil
}
