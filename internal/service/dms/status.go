package dms

import (
	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func statusEndpoint(conn *dms.DatabaseMigrationService, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindEndpointByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
