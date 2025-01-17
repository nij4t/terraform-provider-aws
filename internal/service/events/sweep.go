//go:build sweep
// +build sweep

package events

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_event_api_destination", &resource.Sweeper{
		Name: "aws_cloudwatch_event_api_destination",
		F:    sweepAPIDestination,
		Dependencies: []string{
			"aws_cloudwatch_event_connection",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_archive", &resource.Sweeper{
		Name: "aws_cloudwatch_event_archive",
		F:    sweepArchives,
		Dependencies: []string{
			"aws_cloudwatch_event_bus",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_bus", &resource.Sweeper{
		Name: "aws_cloudwatch_event_bus",
		F:    sweepBuses,
		Dependencies: []string{
			"aws_cloudwatch_event_rule",
			"aws_cloudwatch_event_target",
			"aws_schemas_discoverer",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_connection", &resource.Sweeper{
		Name: "aws_cloudwatch_event_connection",
		F:    sweepConnection,
	})

	resource.AddTestSweepers("aws_cloudwatch_event_permission", &resource.Sweeper{
		Name: "aws_cloudwatch_event_permission",
		F:    sweepPermissions,
	})

	resource.AddTestSweepers("aws_cloudwatch_event_rule", &resource.Sweeper{
		Name: "aws_cloudwatch_event_rule",
		F:    sweepRules,
		Dependencies: []string{
			"aws_cloudwatch_event_target",
		},
	})

	resource.AddTestSweepers("aws_cloudwatch_event_target", &resource.Sweeper{
		Name: "aws_cloudwatch_event_target",
		F:    sweepTargets,
	})
}

func sweepAPIDestination(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn

	var sweeperErrs *multierror.Error

	input := &eventbridge.ListApiDestinationsInput{
		Limit: aws.Int64(100),
	}
	var apiDestinations []*eventbridge.ApiDestination
	for {
		output, err := conn.ListApiDestinations(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge API Destination sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EventBridge API Destinations: %w", err)
		}

		apiDestinations = append(apiDestinations, output.ApiDestinations...)

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, apiDestination := range apiDestinations {

		input := &eventbridge.DeleteApiDestinationInput{
			Name: apiDestination.Name,
		}
		_, err := conn.DeleteApiDestination(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting EventBridge Api Destination (%s): %w", *apiDestination.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d EventBridge Api Destinations", len(apiDestinations))

	return sweeperErrs.ErrorOrNil()
}

func sweepArchives(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn

	input := &eventbridge.ListArchivesInput{}

	for {
		output, err := conn.ListArchives(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge archive sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EventBridge archive: %w", err)
		}

		if len(output.Archives) == 0 {
			log.Print("[DEBUG] No EventBridge archives to sweep")
			return nil
		}

		for _, archive := range output.Archives {
			name := aws.StringValue(archive.ArchiveName)
			if name == "default" {
				continue
			}

			log.Printf("[INFO] Deleting EventBridge archive (%s)", name)
			_, err := conn.DeleteArchive(&eventbridge.DeleteArchiveInput{
				ArchiveName: aws.String(name),
			})
			if err != nil {
				return fmt.Errorf("Error deleting EventBridge archive (%s): %w", name, err)
			}
		}

		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return nil
}

func sweepBuses(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn
	input := &eventbridge.ListEventBusesInput{}
	var sweeperErrs *multierror.Error

	err = listEventBusesPages(conn, input, func(page *eventbridge.ListEventBusesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventBus := range page.EventBuses {
			name := aws.StringValue(eventBus.Name)
			if name == DefaultEventBusName {
				continue
			}

			r := ResourceBus()
			d := r.Data(nil)
			d.SetId(name)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge event bus sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge event buses: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepConnection(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn

	var sweeperErrs *multierror.Error

	input := &eventbridge.ListConnectionsInput{
		Limit: aws.Int64(100),
	}
	var connections []*eventbridge.Connection
	for {
		output, err := conn.ListConnections(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Connection sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("Error retrieving EventBridge Connections: %w", err)
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	for _, connection := range connections {
		input := &eventbridge.DeleteConnectionInput{
			Name: connection.Name,
		}
		_, err := conn.DeleteConnection(input)
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting EventBridge Connection (%s): %w", *connection.Name, err))
			continue
		}
	}

	log.Printf("[INFO] Deleted %d EventBridge Connections", len(connections))

	return sweeperErrs.ErrorOrNil()
}

func sweepPermissions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn

	output, err := conn.DescribeEventBus(&eventbridge.DescribeEventBusInput{})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping EventBridge Permission sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving EventBridge Permissions: %w", err)
	}

	policy := aws.StringValue(output.Policy)

	if policy == "" {
		log.Print("[DEBUG] No EventBridge Permissions to sweep")
		return nil
	}

	var policyDoc PermissionPolicyDoc
	err = json.Unmarshal([]byte(policy), &policyDoc)
	if err != nil {
		return fmt.Errorf("Parsing EventBridge Permissions policy %q failed: %w", policy, err)
	}

	for _, statement := range policyDoc.Statements {
		sid := statement.Sid

		log.Printf("[INFO] Deleting EventBridge Permission %s", sid)
		_, err := conn.RemovePermission(&eventbridge.RemovePermissionInput{
			StatementId: aws.String(sid),
		})
		if err != nil {
			return fmt.Errorf("Error deleting EventBridge Permission %s: %w", sid, err)
		}
	}

	return nil
}

func sweepRules(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn

	var sweeperErrs *multierror.Error
	var count int

	rulesInput := &eventbridge.ListRulesInput{}

	err = listRulesPages(conn, rulesInput, func(rulesPage *eventbridge.ListRulesOutput, lastPage bool) bool {
		if rulesPage == nil {
			return !lastPage
		}

		for _, rule := range rulesPage.Rules {
			count++
			name := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting EventBridge rule (%s)", name)
			_, err := conn.DeleteRule(&eventbridge.DeleteRuleInput{
				Name:  aws.String(name),
				Force: aws.Bool(true), // Required for AWS-managed rules, ignored otherwise
			})
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error deleting EventBridge rule (%s): %w", name, err))
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge rule sweeper for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge rules: %w", err))
	}

	log.Printf("[INFO] Deleted %d EventBridge rules", count)

	return sweeperErrs.ErrorOrNil()
}

func sweepTargets(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).EventsConn

	var sweeperErrs *multierror.Error
	var rulesCount, targetsCount int

	rulesInput := &eventbridge.ListRulesInput{}

	err = listRulesPages(conn, rulesInput, func(rulesPage *eventbridge.ListRulesOutput, lastPage bool) bool {
		if rulesPage == nil {
			return !lastPage
		}

		for _, rule := range rulesPage.Rules {
			rulesCount++
			ruleName := aws.StringValue(rule.Name)

			log.Printf("[INFO] Deleting EventBridge targets for rule (%s)", ruleName)
			targetsInput := &eventbridge.ListTargetsByRuleInput{
				Rule:  rule.Name,
				Limit: aws.Int64(100), // Set limit to allowed maximum to prevent API throttling
			}

			err := listTargetsByRulePages(conn, targetsInput, func(targetsPage *eventbridge.ListTargetsByRuleOutput, lastPage bool) bool {
				if targetsPage == nil {
					return !lastPage
				}

				for _, target := range targetsPage.Targets {
					targetsCount++
					removeTargetsInput := &eventbridge.RemoveTargetsInput{
						Ids:   []*string{target.Id},
						Rule:  rule.Name,
						Force: aws.Bool(true), // Required for AWS-managed rules, ignored otherwise
					}
					targetID := aws.StringValue(target.Id)

					log.Printf("[INFO] Deleting EventBridge target (%s/%s)", ruleName, targetID)
					_, err := conn.RemoveTargets(removeTargetsInput)

					if err != nil {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("Error deleting EventBridge target (%s/%s): %w", ruleName, targetID, err))
						continue
					}
				}

				return !lastPage
			})

			if sweep.SkipSweepError(err) {
				log.Printf("[WARN] Skipping EventBridge target sweeper for %q: %s", region, err)
				return false
			}
			if err != nil {
				sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge targets for rule (%s): %w", ruleName, err))
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping EventBridge rule target sweeper for %q: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EventBridge rules: %w", err))
	}

	log.Printf("[INFO] Deleted %d EventBridge targets across %d EventBridge rules", targetsCount, rulesCount)

	return sweeperErrs.ErrorOrNil()
}
