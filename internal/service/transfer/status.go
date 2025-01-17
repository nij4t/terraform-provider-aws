package transfer

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

const (
	userStateExists = "exists"
)

func statusServerState(conn *transfer.Transfer, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindServerByID(conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusUserState(conn *transfer.Transfer, serverID, userName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindUserByServerIDAndUserName(conn, serverID, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, userStateExists, nil
	}
}
