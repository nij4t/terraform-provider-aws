package kafka

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func statusClusterState(conn *kafka.Kafka, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusClusterOperationState(conn *kafka.Kafka, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterOperationByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.OperationState), nil
	}
}

func statusConfigurationState(conn *kafka.Kafka, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConfigurationByARN(conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}
