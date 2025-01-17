package athena_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/athena"
	"github.com/aws/aws-sdk-go/service/s3"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfathena "github.com/nij4t/terraform-provider-aws/internal/service/athena"
)

func TestAccAthenaDatabase_basic(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dbName := sdkacctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaDatabaseConfig(rInt, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.hoge"),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_encryption(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dbName := sdkacctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaDatabaseWithKMSConfig(rInt, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.hoge"),
					resource.TestCheckResourceAttr("aws_athena_database.hoge", "encryption_configuration.0.encryption_option", "SSE_KMS"),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_nameStartsWithUnderscore(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dbName := "_" + sdkacctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaDatabaseConfig(rInt, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.hoge"),
					resource.TestCheckResourceAttr("aws_athena_database.hoge", "name", dbName),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_nameCantHaveUppercase(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dbName := "A" + sdkacctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAthenaDatabaseConfig(rInt, dbName, false),
				ExpectError: regexp.MustCompile(`must be lowercase letters, numbers, or underscore \('_'\)`),
			},
		},
	})
}

func TestAccAthenaDatabase_destroyFailsIfTablesExist(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dbName := sdkacctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaDatabaseConfig(rInt, dbName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.hoge"),
					testAccDatabaseCreateTables(dbName),
					testAccCheckDatabaseDropFails(dbName),
					testAccDatabaseDestroyTables(dbName),
				),
			},
		},
	})
}

func TestAccAthenaDatabase_forceDestroyAlwaysSucceeds(t *testing.T) {
	rInt := sdkacctest.RandInt()
	dbName := sdkacctest.RandString(8)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, athena.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckDatabaseDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAthenaDatabaseConfig(rInt, dbName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists("aws_athena_database.hoge"),
					testAccDatabaseCreateTables(dbName),
				),
			},
		},
	})
}

// StartQueryExecution requires OutputLocation but terraform destroy deleted S3 bucket as well.
// So temporary S3 bucket as OutputLocation is created to confirm whether the database is actually deleted.
func testAccCheckDatabaseDestroy(s *terraform.State) error {
	athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn
	s3conn := acctest.Provider.Meta().(*conns.AWSClient).S3Conn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_athena_database" {
			continue
		}

		rInt := sdkacctest.RandInt()
		bucketName := fmt.Sprintf("tf-test-athena-db-%d", rInt)
		_, err := s3conn.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return err
		}

		input := &athena.StartQueryExecutionInput{
			QueryString: aws.String("show databases;"),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		ers, err := tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		if err != nil {
			return err
		}
		found := false
		dbName := rs.Primary.Attributes["name"]
		for _, row := range ers.Rows {
			for _, datum := range row.Data {
				if *datum.VarCharValue == dbName {
					found = true
				}
			}
		}
		if found {
			return fmt.Errorf("[DELETE ERROR] Athena failed to drop database: %s", dbName)
		}

		loresp, err := s3conn.ListObjectsV2(
			&s3.ListObjectsV2Input{
				Bucket: aws.String(bucketName),
			},
		)
		if err != nil {
			return fmt.Errorf("[DELETE ERROR] S3 Bucket list Objects err: %s", err)
		}

		objectsToDelete := make([]*s3.ObjectIdentifier, 0)

		if len(loresp.Contents) != 0 {
			for _, v := range loresp.Contents {
				objectsToDelete = append(objectsToDelete, &s3.ObjectIdentifier{
					Key: v.Key,
				})
			}
		}

		_, err = s3conn.DeleteObjects(&s3.DeleteObjectsInput{
			Bucket: aws.String(bucketName),
			Delete: &s3.Delete{
				Objects: objectsToDelete,
			},
		})
		if err != nil {
			return fmt.Errorf("[DELETE ERROR] S3 Bucket delete Objects err: %s", err)
		}

		_, err = s3conn.DeleteBucket(&s3.DeleteBucketInput{
			Bucket: aws.String(bucketName),
		})
		if err != nil {
			return fmt.Errorf("[DELETE ERROR] S3 Bucket delete Bucket err: %s", err)
		}

	}
	return nil
}

func testAccCheckDatabaseExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s, %v", name, s.RootModule().Resources)
		}
		return nil
	}
}

func testAccDatabaseCreateTables(dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccAthenaDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &athena.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String(fmt.Sprintf(
				"create external table foo (bar int) location 's3://%s/';", bucketName)),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		return err
	}
}

func testAccDatabaseDestroyTables(dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccAthenaDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &athena.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String("drop table foo;"),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		return err
	}
}

func testAccCheckDatabaseDropFails(dbName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bucketName, err := testAccAthenaDatabaseFindBucketName(s, dbName)
		if err != nil {
			return err
		}

		athenaconn := acctest.Provider.Meta().(*conns.AWSClient).AthenaConn

		input := &athena.StartQueryExecutionInput{
			QueryExecutionContext: &athena.QueryExecutionContext{
				Database: aws.String(dbName),
			},
			QueryString: aws.String(fmt.Sprintf("drop database `%s`;", dbName)),
			ResultConfiguration: &athena.ResultConfiguration{
				OutputLocation: aws.String("s3://" + bucketName),
			},
		}

		resp, err := athenaconn.StartQueryExecution(input)
		if err != nil {
			return err
		}

		_, err = tfathena.QueryExecutionResult(*resp.QueryExecutionId, athenaconn)
		if err == nil {
			return fmt.Errorf("drop database unexpectedly succeeded for a database with tables")
		}

		return nil
	}
}

func testAccAthenaDatabaseFindBucketName(s *terraform.State, dbName string) (bucket string, err error) {
	for _, rs := range s.RootModule().Resources {
		if rs.Type == "aws_athena_database" && rs.Primary.Attributes["name"] == dbName {
			bucket = rs.Primary.Attributes["bucket"]
			break
		}
	}

	if bucket == "" {
		err = fmt.Errorf("cannot find database %s", dbName)
	}

	return bucket, err
}

func testAccAthenaDatabaseConfig(randInt int, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "hoge" {
  bucket        = "tf-test-athena-db-%[1]d"
  force_destroy = true
}

resource "aws_athena_database" "hoge" {
  name          = "%[2]s"
  bucket        = aws_s3_bucket.hoge.bucket
  force_destroy = %[3]t
}
`, randInt, dbName, forceDestroy)
}

func testAccAthenaDatabaseWithKMSConfig(randInt int, dbName string, forceDestroy bool) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "hoge" {
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "hoge" {
  bucket        = "tf-test-athena-db-%[1]d"
  force_destroy = true

  server_side_encryption_configuration {
    rule {
      apply_server_side_encryption_by_default {
        kms_master_key_id = aws_kms_key.hoge.arn
        sse_algorithm     = "aws:kms"
      }
    }
  }
}

resource "aws_athena_database" "hoge" {
  name          = "%[2]s"
  bucket        = aws_s3_bucket.hoge.bucket
  force_destroy = %[3]t

  encryption_configuration {
    encryption_option = "SSE_KMS"
    kms_key           = aws_kms_key.hoge.arn
  }
}
`, randInt, dbName, forceDestroy)
}
