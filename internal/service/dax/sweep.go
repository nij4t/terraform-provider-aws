//go:build sweep
// +build sweep

package dax

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/dax"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_dax_cluster", &resource.Sweeper{
		Name: "aws_dax_cluster",
		F:    sweepClusters,
	})
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).DAXConn

	resp, err := conn.DescribeClusters(&dax.DescribeClustersInput{})
	if err != nil {
		// GovCloud (with no DAX support) has an endpoint that responds with:
		// InvalidParameterValueException: Access Denied to API Version: DAX_V3
		if sweep.SkipSweepError(err) || tfawserr.ErrMessageContains(err, "InvalidParameterValueException", "Access Denied to API Version: DAX_V3") {
			log.Printf("[WARN] Skipping DAX Cluster sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DAX clusters: %s", err)
	}

	if len(resp.Clusters) == 0 {
		log.Print("[DEBUG] No DAX clusters to sweep")
		return nil
	}

	log.Printf("[INFO] Found %d DAX clusters", len(resp.Clusters))

	for _, cluster := range resp.Clusters {
		log.Printf("[INFO] Deleting DAX cluster %s", *cluster.ClusterName)
		_, err := conn.DeleteCluster(&dax.DeleteClusterInput{
			ClusterName: cluster.ClusterName,
		})
		if err != nil {
			return fmt.Errorf("Error deleting DAX cluster %s: %s", *cluster.ClusterName, err)
		}
	}

	return nil
}
