//go:build sweep
// +build sweep

package ecr

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_ecr_repository", &resource.Sweeper{
		Name: "aws_ecr_repository",
		F:    sweepRepositories,
	})
}

func sweepRepositories(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).ECRConn

	var errors error
	err = conn.DescribeRepositoriesPages(&ecr.DescribeRepositoriesInput{}, func(page *ecr.DescribeRepositoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, repository := range page.Repositories {
			repositoryName := aws.StringValue(repository.RepositoryName)
			log.Printf("[INFO] Deleting ECR repository: %s", repositoryName)

			_, err = conn.DeleteRepository(&ecr.DeleteRepositoryInput{
				// We should probably sweep repositories even if there are images.
				Force:          aws.Bool(true),
				RegistryId:     repository.RegistryId,
				RepositoryName: repository.RepositoryName,
			})
			if err != nil {
				if !tfawserr.ErrMessageContains(err, ecr.ErrCodeRepositoryNotFoundException, "") {
					sweeperErr := fmt.Errorf("Error deleting ECR repository (%s): %w", repositoryName, err)
					log.Printf("[ERROR] %s", sweeperErr)
					errors = multierror.Append(errors, sweeperErr)
				}
				continue
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping ECR repository sweep for %s: %s", region, err)
			return nil
		}
		errors = multierror.Append(errors, fmt.Errorf("Error retreiving ECR repositories: %w", err))
	}

	return errors
}
