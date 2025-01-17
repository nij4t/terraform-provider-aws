package kinesis_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfkinesis "github.com/nij4t/terraform-provider-aws/internal/service/kinesis"
)

func TestAccKinesisStreamConsumer_basic(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	streamName := "aws_kinesis_stream.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kinesis.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "kinesis", regexp.MustCompile(fmt.Sprintf("stream/%[1]s/consumer/%[1]s", rName))),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrPair(resourceName, "stream_arn", streamName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_timestamp"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccKinesisStreamConsumer_disappears(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kinesis.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccStreamConsumerConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfkinesis.ResourceStreamConsumer(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccKinesisStreamConsumer_maxConcurrentConsumers(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kinesis.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				// Test creation of max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccStreamConsumerConfig_multiple(rName, 5),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(fmt.Sprintf("%s.0", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.1", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.2", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.3", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.4", resourceName)),
				),
			},
		},
	})
}

func TestAccKinesisStreamConsumer_exceedMaxConcurrentConsumers(t *testing.T) {
	resourceName := "aws_kinesis_stream_consumer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, kinesis.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckStreamConsumerDestroy,
		Steps: []resource.TestStep{
			{
				// Test creation of more than the max number (5 according to AWS API docs) of concurrent consumers for a single stream
				Config: testAccStreamConsumerConfig_multiple(rName, 10),
				Check: resource.ComposeTestCheckFunc(
					testAccStreamConsumerExists(fmt.Sprintf("%s.0", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.1", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.2", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.3", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.4", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.5", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.6", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.7", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.8", resourceName)),
					testAccStreamConsumerExists(fmt.Sprintf("%s.9", resourceName)),
				),
			},
		},
	})
}

func testAccCheckStreamConsumerDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_stream_consumer" {
			continue
		}

		consumer, err := tfkinesis.FindStreamConsumerByARN(conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, kinesis.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Kinesis Stream Consumer (%s): %w", rs.Primary.ID, err)
		}

		if consumer != nil {
			return fmt.Errorf("Kinesis Stream Consumer (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccStreamConsumerExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource %s not found", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).KinesisConn

		consumer, err := tfkinesis.FindStreamConsumerByARN(conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading Kinesis Stream Consumer (%s): %w", rs.Primary.ID, err)
		}

		if consumer == nil {
			return fmt.Errorf("Kinesis Stream Consumer (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccStreamConsumerBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %q
  shard_count = 2
}
`, rName)
}

func testAccStreamConsumerConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccStreamConsumerBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  name       = %q
  stream_arn = aws_kinesis_stream.test.arn
}
`, rName))
}

func testAccStreamConsumerConfig_multiple(rName string, count int) string {
	return acctest.ConfigCompose(
		testAccStreamConsumerBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_kinesis_stream_consumer" "test" {
  count      = %d
  name       = "%s-${count.index}"
  stream_arn = aws_kinesis_stream.test.arn
}
`, count, rName))
}
