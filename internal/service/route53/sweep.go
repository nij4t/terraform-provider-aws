//go:build sweep
// +build sweep

package route53

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_route53_health_check", &resource.Sweeper{
		Name: "aws_route53_health_check",
		F:    sweepHealthchecks,
	})

	resource.AddTestSweepers("aws_route53_key_signing_key", &resource.Sweeper{
		Name: "aws_route53_key_signing_key",
		F:    sweepKeySigningKeys,
	})

	resource.AddTestSweepers("aws_route53_query_log", &resource.Sweeper{
		Name: "aws_route53_query_log",
		F:    sweepQueryLogs,
	})

	resource.AddTestSweepers("aws_route53_zone", &resource.Sweeper{
		Name: "aws_route53_zone",
		Dependencies: []string{
			"aws_service_discovery_http_namespace",
			"aws_service_discovery_public_dns_namespace",
			"aws_service_discovery_private_dns_namespace",
			"aws_elb",
			"aws_route53_key_signing_key",
		},
		F: sweepZones,
	})
}

func sweepHealthchecks(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).Route53Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &route53.ListHealthChecksInput{}

	err = conn.ListHealthChecksPages(input, func(page *route53.ListHealthChecksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, detail := range page.HealthChecks {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			r := ResourceHealthCheck()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Route53 Health Checks for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestratorContext(context.Background(), sweepResources, 0*time.Minute, 1*time.Minute, 10*time.Second, 18*time.Second, 10*time.Minute); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Route53 Health Checks for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Health Checks sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepKeySigningKeys(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).Route53Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &route53.ListHostedZonesInput{}

	err = conn.ListHostedZonesPages(input, func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

	MAIN:
		for _, detail := range page.HostedZones {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.StringValue(detail.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone (%s): %s", domain, id)
					continue MAIN
				}
			}

			dnsInput := &route53.GetDNSSECInput{
				HostedZoneId: detail.Id,
			}

			output, err := conn.GetDNSSEC(dnsInput)

			if tfawserr.ErrMessageContains(err, route53.ErrCodeInvalidArgument, "private hosted zones") {
				continue
			}

			if err != nil {
				errs = multierror.Append(errs, fmt.Errorf("error getting Route53 DNS SEC for %s: %w", region, err))
			}

			for _, dns := range output.KeySigningKeys {
				r := ResourceKeySigningKey()
				d := r.Data(nil)
				d.SetId(id)
				d.Set("hosted_zone_id", id)
				d.Set("name", dns.Name)
				d.Set("status", dns.Status)

				sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
			}

		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error getting Route53 Key-Signing Keys for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestratorContext(context.Background(), sweepResources, 0*time.Millisecond, 1*time.Minute, 30*time.Second, 30*time.Second, 10*time.Minute); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Route53 Key-Signing Keys for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Key-Signing Keys sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepQueryLogs(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).Route53Conn
	var sweeperErrs *multierror.Error

	err = conn.ListQueryLoggingConfigsPages(&route53.ListQueryLoggingConfigsInput{}, func(page *route53.ListQueryLoggingConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, queryLoggingConfig := range page.QueryLoggingConfigs {
			id := aws.StringValue(queryLoggingConfig.Id)

			r := ResourceQueryLog()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Route53 query logging configuration (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	// In unsupported AWS partitions, the API may return an error even the SDK cannot handle.
	// Reference: https://github.com/aws/aws-sdk-go/issues/3313
	if sweep.SkipSweepError(err) || tfawserr.ErrMessageContains(err, "SerializationError", "failed to unmarshal error message") || tfawserr.ErrMessageContains(err, "AccessDeniedException", "Unable to determine service/operation name to be authorized") {
		log.Printf("[WARN] Skipping Route53 query logging configurations sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Route53 query logging configurations: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepZones(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).Route53Conn
	sweepResources := make([]*sweep.SweepResource, 0)
	var errs *multierror.Error

	input := &route53.ListHostedZonesInput{}

	err = conn.ListHostedZonesPages(input, func(page *route53.ListHostedZonesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

	MAIN:
		for _, detail := range page.HostedZones {
			if detail == nil {
				continue
			}

			id := aws.StringValue(detail.Id)

			for _, domain := range hostedZonesToPreserve() {
				if strings.Contains(aws.StringValue(detail.Name), domain) {
					log.Printf("[DEBUG] Skipping Route53 Hosted Zone (%s): %s", domain, id)
					continue MAIN
				}
			}

			r := ResourceZone()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("force_destroy", true)
			d.Set("name", detail.Name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing Route53 Hosted Zones for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestratorContext(context.Background(), sweepResources, 0*time.Minute, 1*time.Minute, 10*time.Second, 18*time.Second, 10*time.Minute); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping Route53 Hosted Zones for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping Route53 Hosted Zones sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func hostedZonesToPreserve() []string {
	return []string{
		"acmetest.hashicorp.engineering",
		"tfacc.hashicorptest.com",
		"aws.tfacc.hashicorptest.com",
		"hashicorp.com",
		"terraform-provider-aws-acctest-acm.com",
	}
}
