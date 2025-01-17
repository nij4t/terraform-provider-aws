//go:build sweep
// +build sweep

package secretsmanager

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_secretsmanager_secret_policy", &resource.Sweeper{
		Name: "aws_secretsmanager_secret_policy",
		F:    sweepSecretPolicies,
	})

	resource.AddTestSweepers("aws_secretsmanager_secret", &resource.Sweeper{
		Name: "aws_secretsmanager_secret",
		F:    sweepSecrets,
	})
}

func sweepSecretPolicies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SecretsManagerConn

	err = conn.ListSecretsPages(&secretsmanager.ListSecretsInput{}, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		if len(page.SecretList) == 0 {
			log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			return true
		}

		for _, secret := range page.SecretList {
			name := aws.StringValue(secret.Name)

			log.Printf("[INFO] Deleting Secrets Manager Secret Policy: %s", name)
			input := &secretsmanager.DeleteResourcePolicyInput{
				SecretId: aws.String(name),
			}

			_, err := conn.DeleteResourcePolicy(input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret Policy (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %w", err)
	}
	return nil
}

func sweepSecrets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).SecretsManagerConn

	err = conn.ListSecretsPages(&secretsmanager.ListSecretsInput{}, func(page *secretsmanager.ListSecretsOutput, lastPage bool) bool {
		if len(page.SecretList) == 0 {
			log.Print("[DEBUG] No Secrets Manager Secrets to sweep")
			return true
		}

		for _, secret := range page.SecretList {
			name := aws.StringValue(secret.Name)

			log.Printf("[INFO] Deleting Secrets Manager Secret: %s", name)
			input := &secretsmanager.DeleteSecretInput{
				ForceDeleteWithoutRecovery: aws.Bool(true),
				SecretId:                   aws.String(name),
			}

			_, err := conn.DeleteSecret(input)
			if err != nil {
				if tfawserr.ErrMessageContains(err, secretsmanager.ErrCodeResourceNotFoundException, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Secrets Manager Secret (%s): %s", name, err)
			}
		}

		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Secrets Manager Secret sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Secrets Manager Secrets: %s", err)
	}
	return nil
}
