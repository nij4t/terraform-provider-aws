package cloudfront_test

import (
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
)

// testAccCloudfrontRegionProviderConfig is the Terraform provider configuration for CloudFront region testing
//
// Testing CloudFront assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccCloudfrontRegionProviderConfig() string {
	switch acctest.Partition() {
	case endpoints.AwsPartitionID:
		return acctest.ConfigRegionalProvider(endpoints.UsEast1RegionID)
	case endpoints.AwsCnPartitionID:
		return acctest.ConfigRegionalProvider(endpoints.CnNorthwest1RegionID)
	default:
		return acctest.ConfigRegionalProvider(acctest.Region())
	}
}
