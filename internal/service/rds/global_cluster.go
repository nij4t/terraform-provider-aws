package rds

import (
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

const (
	rdsGlobalClusterRemovalTimeout = 2 * time.Minute
)

func ResourceGlobalCluster() *schema.Resource {
	return &schema.Resource{
		Create: resourceGlobalClusterCreate,
		Read:   resourceGlobalClusterRead,
		Update: resourceGlobalClusterUpdate,
		Delete: resourceGlobalClusterDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"database_name": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"deletion_protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"engine": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"source_db_cluster_identifier"},
				ValidateFunc: validation.StringInSlice([]string{
					"aurora",
					"aurora-mysql",
					"aurora-postgresql",
				}, false),
			},
			"engine_version": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"force_destroy": {
				Type:     schema.TypeBool,
				Optional: true,
			},
			"global_cluster_identifier": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"global_cluster_members": {
				Type:     schema.TypeSet,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"db_cluster_arn": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"is_writer": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
			"global_cluster_resource_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"source_db_cluster_identifier": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"engine"},
				RequiredWith:  []string{"force_destroy"},
			},
			"storage_encrypted": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
		},
	}
}

func resourceGlobalClusterCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.CreateGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(d.Get("global_cluster_identifier").(string)),
	}

	if v, ok := d.GetOk("database_name"); ok {
		input.DatabaseName = aws.String(v.(string))
	}

	if v, ok := d.GetOk("deletion_protection"); ok {
		input.DeletionProtection = aws.Bool(v.(bool))
	}

	if v, ok := d.GetOk("engine"); ok {
		input.Engine = aws.String(v.(string))
	}

	if v, ok := d.GetOk("engine_version"); ok {
		input.EngineVersion = aws.String(v.(string))
	}

	if v, ok := d.GetOk("source_db_cluster_identifier"); ok {
		input.SourceDBClusterIdentifier = aws.String(v.(string))
	}

	if v, ok := d.GetOk("storage_encrypted"); ok {
		input.StorageEncrypted = aws.Bool(v.(bool))
	}

	// Prevent the following error and keep the previous default,
	// since we cannot have Engine default after adding SourceDBClusterIdentifier:
	// InvalidParameterValue: When creating standalone global cluster, value for engineName should be specified
	if input.Engine == nil && input.SourceDBClusterIdentifier == nil {
		input.Engine = aws.String("aurora")
	}

	output, err := conn.CreateGlobalCluster(input)
	if err != nil {
		return fmt.Errorf("error creating RDS Global Cluster: %s", err)
	}

	d.SetId(aws.StringValue(output.GlobalCluster.GlobalClusterIdentifier))

	if err := waitForGlobalClusterCreation(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for RDS Global Cluster (%s) availability: %s", d.Id(), err)
	}

	return resourceGlobalClusterRead(d, meta)
}

func resourceGlobalClusterRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	globalCluster, err := DescribeGlobalCluster(conn, d.Id())

	if tfawserr.ErrMessageContains(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
		log.Printf("[WARN] RDS Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading RDS Global Cluster: %s", err)
	}

	if globalCluster == nil {
		log.Printf("[WARN] RDS Global Cluster (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if aws.StringValue(globalCluster.Status) == "deleting" || aws.StringValue(globalCluster.Status) == "deleted" {
		log.Printf("[WARN] RDS Global Cluster (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(globalCluster.Status))
		d.SetId("")
		return nil
	}

	d.Set("arn", globalCluster.GlobalClusterArn)
	d.Set("database_name", globalCluster.DatabaseName)
	d.Set("deletion_protection", globalCluster.DeletionProtection)
	d.Set("engine", globalCluster.Engine)
	d.Set("engine_version", globalCluster.EngineVersion)
	d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)

	if err := d.Set("global_cluster_members", flattenGlobalClusterMembers(globalCluster.GlobalClusterMembers)); err != nil {
		return fmt.Errorf("error setting global_cluster_members: %w", err)
	}

	d.Set("global_cluster_resource_id", globalCluster.GlobalClusterResourceId)
	d.Set("storage_encrypted", globalCluster.StorageEncrypted)

	return nil
}

func resourceGlobalClusterUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	input := &rds.ModifyGlobalClusterInput{
		DeletionProtection:      aws.Bool(d.Get("deletion_protection").(bool)),
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	if d.HasChange("engine_version") {
		if err := resourceGlobalClusterUpgradeEngineVersion(d, conn); err != nil {
			return err
		}
	}

	log.Printf("[DEBUG] Updating RDS Global Cluster (%s): %s", d.Id(), input)
	_, err := conn.ModifyGlobalCluster(input)

	if tfawserr.ErrMessageContains(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting RDS Global Cluster: %s", err)
	}

	if err := waitForGlobalClusterUpdate(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for RDS Global Cluster (%s) update: %s", d.Id(), err)
	}

	return resourceGlobalClusterRead(d, meta)
}

func resourceGlobalClusterGetIdByARN(conn *rds.RDS, arn string) string {
	result, err := conn.DescribeDBClusters(&rds.DescribeDBClustersInput{})
	if err != nil {
		return ""
	}
	for _, cluster := range result.DBClusters {
		if aws.StringValue(cluster.DBClusterArn) == arn {
			return aws.StringValue(cluster.DBClusterIdentifier)
		}
	}
	return ""
}

func resourceGlobalClusterDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).RDSConn

	if d.Get("force_destroy").(bool) {
		for _, globalClusterMemberRaw := range d.Get("global_cluster_members").(*schema.Set).List() {
			globalClusterMember, ok := globalClusterMemberRaw.(map[string]interface{})

			if !ok {
				continue
			}

			dbClusterArn, ok := globalClusterMember["db_cluster_arn"].(string)

			if !ok {
				continue
			}

			input := &rds.RemoveFromGlobalClusterInput{
				DbClusterIdentifier:     aws.String(dbClusterArn),
				GlobalClusterIdentifier: aws.String(d.Id()),
			}

			_, err := conn.RemoveFromGlobalCluster(input)

			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "is not found in global cluster") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error removing RDS DB Cluster (%s) from Global Cluster (%s): %w", dbClusterArn, d.Id(), err)
			}

			if err := waitForGlobalClusterRemoval(conn, dbClusterArn); err != nil {
				return fmt.Errorf("error waiting for RDS DB Cluster (%s) removal from RDS Global Cluster (%s): %w", dbClusterArn, d.Id(), err)
			}
		}
	}

	input := &rds.DeleteGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Deleting RDS Global Cluster (%s): %s", d.Id(), input)

	// Allow for eventual consistency
	// InvalidGlobalClusterStateFault: Global Cluster arn:aws:rds::123456789012:global-cluster:tf-acc-test-5618525093076697001-0 is not empty
	err := resource.Retry(1*time.Minute, func() *resource.RetryError {
		_, err := conn.DeleteGlobalCluster(input)

		if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidGlobalClusterStateFault, "is not empty") {
			return resource.RetryableError(err)
		}

		if err != nil {
			return resource.NonRetryableError(err)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = conn.DeleteGlobalCluster(input)
	}

	if tfawserr.ErrMessageContains(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting RDS Global Cluster: %s", err)
	}

	if err := WaitForGlobalClusterDeletion(conn, d.Id()); err != nil {
		return fmt.Errorf("error waiting for RDS Global Cluster (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func flattenGlobalClusterMembers(apiObjects []*rds.GlobalClusterMember) []interface{} {
	if len(apiObjects) == 0 {
		return nil
	}

	var tfList []interface{}

	for _, apiObject := range apiObjects {
		tfMap := map[string]interface{}{
			"db_cluster_arn": aws.StringValue(apiObject.DBClusterArn),
			"is_writer":      aws.BoolValue(apiObject.IsWriter),
		}

		tfList = append(tfList, tfMap)
	}

	return tfList
}

func DescribeGlobalCluster(conn *rds.RDS, globalClusterID string) (*rds.GlobalCluster, error) {
	var globalCluster *rds.GlobalCluster

	input := &rds.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	log.Printf("[DEBUG] Reading RDS Global Cluster (%s): %s", globalClusterID, input)
	err := conn.DescribeGlobalClustersPages(input, func(page *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalClusters {
			if gc == nil {
				continue
			}

			if aws.StringValue(gc.GlobalClusterIdentifier) == globalClusterID {
				globalCluster = gc
				return false
			}
		}

		return !lastPage
	})

	return globalCluster, err
}

func DescribeGlobalClusterFromClusterARN(conn *rds.RDS, dbClusterARN string) (*rds.GlobalCluster, error) {
	var globalCluster *rds.GlobalCluster

	input := &rds.DescribeGlobalClustersInput{
		Filters: []*rds.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: []*string{aws.String(dbClusterARN)},
			},
		},
	}

	log.Printf("[DEBUG] Reading RDS Global Clusters: %s", input)
	err := conn.DescribeGlobalClustersPages(input, func(page *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalClusters {
			if gc == nil {
				continue
			}

			for _, globalClusterMember := range gc.GlobalClusterMembers {
				if aws.StringValue(globalClusterMember.DBClusterArn) == dbClusterARN {
					globalCluster = gc
					return false
				}
			}
		}

		return !lastPage
	})

	return globalCluster, err
}

func rdsGlobalClusterRefreshFunc(conn *rds.RDS, globalClusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		globalCluster, err := DescribeGlobalCluster(conn, globalClusterID)

		if tfawserr.ErrMessageContains(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
			return nil, "deleted", nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading RDS Global Cluster (%s): %s", globalClusterID, err)
		}

		if globalCluster == nil {
			return nil, "deleted", nil
		}

		return globalCluster, aws.StringValue(globalCluster.Status), nil
	}
}

func waitForGlobalClusterCreation(conn *rds.RDS, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating"},
		Target:  []string{"available"},
		Refresh: rdsGlobalClusterRefreshFunc(conn, globalClusterID),
		Timeout: 10 * time.Minute,
	}

	log.Printf("[DEBUG] Waiting for RDS Global Cluster (%s) availability", globalClusterID)
	_, err := stateConf.WaitForState()

	return err
}

func waitForGlobalClusterUpdate(conn *rds.RDS, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"modifying", "upgrading"},
		Target:  []string{"available"},
		Refresh: rdsGlobalClusterRefreshFunc(conn, globalClusterID),
		Timeout: 10 * time.Minute,
		Delay:   30 * time.Second,
	}

	log.Printf("[DEBUG] Waiting for RDS Global Cluster (%s) availability", globalClusterID)
	_, err := stateConf.WaitForState()

	return err
}

func WaitForGlobalClusterDeletion(conn *rds.RDS, globalClusterID string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"available",
			"deleting",
		},
		Target:         []string{"deleted"},
		Refresh:        rdsGlobalClusterRefreshFunc(conn, globalClusterID),
		Timeout:        10 * time.Minute,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for RDS Global Cluster (%s) deletion", globalClusterID)
	_, err := stateConf.WaitForState()

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitForGlobalClusterRemoval(conn *rds.RDS, dbClusterIdentifier string) error {
	var globalCluster *rds.GlobalCluster
	stillExistsErr := fmt.Errorf("RDS DB Cluster still exists in RDS Global Cluster")

	err := resource.Retry(rdsGlobalClusterRemovalTimeout, func() *resource.RetryError {
		var err error

		globalCluster, err = DescribeGlobalClusterFromClusterARN(conn, dbClusterIdentifier)

		if err != nil {
			return resource.NonRetryableError(err)
		}

		if globalCluster != nil {
			return resource.RetryableError(stillExistsErr)
		}

		return nil
	})

	if tfresource.TimedOut(err) {
		_, err = DescribeGlobalClusterFromClusterARN(conn, dbClusterIdentifier)
	}

	if err != nil {
		return err
	}

	if globalCluster != nil {
		return stillExistsErr
	}

	return nil
}

func resourceGlobalClusterUpgradeMajorEngineVersion(clusterId string, engineVersion string, conn *rds.RDS) error {
	input := &rds.ModifyGlobalClusterInput{
		GlobalClusterIdentifier: aws.String(clusterId),
	}
	input.AllowMajorVersionUpgrade = aws.Bool(true)
	input.EngineVersion = aws.String(engineVersion)
	err := resource.Retry(rdsClusterInitiateUpgradeTimeout, func() *resource.RetryError {
		_, err := conn.ModifyGlobalCluster(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, rds.ErrCodeGlobalClusterNotFoundFault, "") {
				return resource.NonRetryableError(err)
			}
			if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "ModifyGlobalCluster only supports Major Version Upgrades. To patch the members of your global cluster to a newer minor version you need to call ModifyDbCluster in each one of them.") {
				return resource.NonRetryableError(err)
			}

			return resource.RetryableError(err)
		}

		return nil
	})
	if tfresource.TimedOut(err) {
		_, err = conn.ModifyGlobalCluster(input)
	}
	return err
}

func resourceGlobalClusterUpgradeMinorEngineVersion(clusterMembers *schema.Set, engineVersion string, conn *rds.RDS) error {
	for _, clusterMemberRaw := range clusterMembers.List() {
		clusterMember := clusterMemberRaw.(map[string]interface{})
		if clusterMemberArn, ok := clusterMember["db_cluster_arn"]; ok && clusterMemberArn.(string) != "" {
			modInput := &rds.ModifyDBClusterInput{
				ApplyImmediately:    aws.Bool(true),
				DBClusterIdentifier: aws.String(clusterMemberArn.(string)),
				EngineVersion:       aws.String(engineVersion),
			}
			err := resource.Retry(rdsClusterInitiateUpgradeTimeout, func() *resource.RetryError {
				_, err := conn.ModifyDBCluster(modInput)
				if err != nil {
					if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "IAM role ARN value is invalid or does not include the required permissions") {
						return resource.RetryableError(err)
					}

					if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBClusterStateFault, "Cannot modify engine version without a primary instance in DB cluster") {
						return resource.NonRetryableError(err)
					}

					if tfawserr.ErrMessageContains(err, rds.ErrCodeInvalidDBClusterStateFault, "") {
						return resource.RetryableError(err)
					}
					return resource.NonRetryableError(err)
				}
				return nil
			})
			if tfresource.TimedOut(err) {
				_, err := conn.ModifyDBCluster(modInput)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return fmt.Errorf("Failed to update engine_version on global cluster member (%s): %s", clusterMemberArn, err)
			}
		}
	}
	return nil
}

func resourceGlobalClusterUpgradeEngineVersion(d *schema.ResourceData, conn *rds.RDS) error {
	log.Printf("[DEBUG] Upgrading RDS Global Cluster (%s) engine version: %s", d.Id(), d.Get("engine_version"))
	err := resourceGlobalClusterUpgradeMajorEngineVersion(d.Id(), d.Get("engine_version").(string), conn)
	if tfawserr.ErrMessageContains(err, "InvalidParameterValue", "ModifyGlobalCluster only supports Major Version Upgrades. To patch the members of your global cluster to a newer minor version you need to call ModifyDbCluster in each one of them.") {
		err = resourceGlobalClusterUpgradeMinorEngineVersion(d.Get("global_cluster_members").(*schema.Set), d.Get("engine_version").(string), conn)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}
	globalCluster, err := DescribeGlobalCluster(conn, d.Id())
	if err != nil {
		return err
	}
	for _, clusterMember := range globalCluster.GlobalClusterMembers {
		err := waitForRDSClusterUpdate(conn, resourceGlobalClusterGetIdByARN(conn, aws.StringValue(clusterMember.DBClusterArn)), d.Timeout(schema.TimeoutUpdate))
		if err != nil {
			return err
		}
	}
	return nil
}
