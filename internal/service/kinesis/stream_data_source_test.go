package kinesis_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/nij4t/terraform-provider-aws/internal/service/kinesis"
)

func TestAccKinesisStreamDataSource_basic(t *testing.T) {
	var stream kinesis.StreamDescription

	sn := fmt.Sprintf("terraform-kinesis-test-%d", sdkacctest.RandInt())
	config := fmt.Sprintf(testAccCheckStreamDataSourceConfig, sn)

	updateShardCount := func() {
		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisConn
		_, err := conn.UpdateShardCount(&kinesis.UpdateShardCountInput{
			ScalingType:      aws.String(kinesis.ScalingTypeUniformScaling),
			StreamName:       aws.String(sn),
			TargetShardCount: aws.Int64(3),
		})
		if err != nil {
			t.Fatalf("Error calling UpdateShardCount: %s", err)
		}
		if err := tfkinesis.WaitForToBeActive(conn, 5*time.Minute, sn); err != nil {
			t.Fatal(err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kinesis.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckKinesisStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream.test_stream", "arn"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "name", sn),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "open_shards.#", "2"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "closed_shards.#", "0"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "shard_level_metrics.#", "2"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "retention_period", "72"),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream.test_stream", "creation_timestamp"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "tags.Name", "tf-test"),
				),
			},
			{
				Config:    config,
				PreConfig: updateShardCount,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisStreamExists("aws_kinesis_stream.test_stream", &stream),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream.test_stream", "arn"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "name", sn),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "status", "ACTIVE"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "open_shards.#", "3"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "closed_shards.#", "4"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "shard_level_metrics.#", "2"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "retention_period", "72"),
					resource.TestCheckResourceAttrSet("data.aws_kinesis_stream.test_stream", "creation_timestamp"),
					resource.TestCheckResourceAttr("data.aws_kinesis_stream.test_stream", "tags.Name", "tf-test"),
				),
			},
		},
	})
}

var testAccCheckStreamDataSourceConfig = `
resource "aws_kinesis_stream" "test_stream" {
  name             = "%s"
  shard_count      = 2
  retention_period = 72

  tags = {
    Name = "tf-test"
  }

  shard_level_metrics = [
    "IncomingBytes",
    "OutgoingBytes"
  ]

  lifecycle {
    ignore_changes = ["shard_count"]
  }
}

data "aws_kinesis_stream" "test_stream" {
  name = aws_kinesis_stream.test_stream.name
}
`
