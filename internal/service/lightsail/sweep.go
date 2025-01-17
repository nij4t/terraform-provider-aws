//go:build sweep
// +build sweep

package lightsail

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_lightsail_instance", &resource.Sweeper{
		Name: "aws_lightsail_instance",
		F:    sweepInstances,
	})

	resource.AddTestSweepers("aws_lightsail_static_ip", &resource.Sweeper{
		Name: "aws_lightsail_static_ip",
		F:    sweepStaticIPs,
	})
}

func sweepInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LightsailConn

	input := &lightsail.GetInstancesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetInstances(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Lightsail Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving Lightsail Instances: %s", err)
		}

		for _, instance := range output.Instances {
			name := aws.StringValue(instance.Name)
			input := &lightsail.DeleteInstanceInput{
				InstanceName: instance.Name,
			}

			log.Printf("[INFO] Deleting Lightsail Instance: %s", name)
			_, err := conn.DeleteInstance(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Lightsail Instance (%s): %s", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		if aws.StringValue(output.NextPageToken) == "" {
			break
		}

		input.PageToken = output.NextPageToken
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepStaticIPs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).LightsailConn

	input := &lightsail.GetStaticIpsInput{}

	for {
		output, err := conn.GetStaticIps(input)
		if err != nil {
			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping Lightsail Static IP sweep for %s: %s", region, err)
				return nil
			}
			return fmt.Errorf("Error retrieving Lightsail Static IPs: %s", err)
		}

		if len(output.StaticIps) == 0 {
			log.Print("[DEBUG] No Lightsail Static IPs to sweep")
			return nil
		}

		for _, staticIp := range output.StaticIps {
			name := aws.StringValue(staticIp.Name)

			log.Printf("[INFO] Deleting Lightsail Static IP %s", name)
			_, err := conn.ReleaseStaticIp(&lightsail.ReleaseStaticIpInput{
				StaticIpName: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting Lightsail Static IP %s: %s", name, err)
			}
		}

		if output.NextPageToken == nil {
			break
		}
		input.PageToken = output.NextPageToken
	}

	return nil
}
