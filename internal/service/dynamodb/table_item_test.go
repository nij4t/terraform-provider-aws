package dynamodb_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfdynamodb "github.com/nij4t/terraform-provider-aws/internal/service/dynamodb"
)

func TestAccDynamoDBTableItem_basic(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	hashKey := "hashKey"
	itemContent := `{
	"hashKey": {"S": "something"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dynamodb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccItemBasicConfig(tableName, hashKey, itemContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemContent+"\n"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableItem_rangeKey(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	hashKey := "hashKey"
	rangeKey := "rangeKey"
	itemContent := `{
	"hashKey": {"S": "something"},
	"rangeKey": {"S": "something-else"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dynamodb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccItemWithRangeKeyConfig(tableName, hashKey, rangeKey, itemContent),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemContent+"\n"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableItem_withMultipleItems(t *testing.T) {
	var conf1 dynamodb.GetItemOutput
	var conf2 dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	hashKey := "hashKey"
	rangeKey := "rangeKey"
	firstItem := `{
	"hashKey": {"S": "something"},
	"rangeKey": {"S": "first"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"}
}`
	secondItem := `{
	"hashKey": {"S": "something"},
	"rangeKey": {"S": "second"},
	"one": {"S": "one"},
	"two": {"S": "two"},
	"three": {"S": "three"},
	"four": {"S": "four"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dynamodb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccItemWithMultipleItemsConfig(tableName, hashKey, rangeKey, firstItem, secondItem),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test1", &conf1),
					testAccCheckTableItemExists("aws_dynamodb_table_item.test2", &conf2),
					testAccCheckTableItemCount(tableName, 2),

					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test1", "item", firstItem+"\n"),

					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test2", "item", secondItem+"\n"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableItem_update(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	hashKey := "hashKey"

	itemBefore := `{
	"hashKey": {"S": "before"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"three": {"N": "33333"},
	"four": {"N": "44444"}
}`
	itemAfter := `{
	"hashKey": {"S": "before"},
	"one": {"N": "11111"},
	"two": {"N": "22222"},
	"new": {"S": "shiny new one"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dynamodb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccItemBasicConfig(tableName, hashKey, itemBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemBefore+"\n"),
				),
			},
			{
				Config: testAccItemBasicConfig(tableName, hashKey, itemAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemAfter+"\n"),
				),
			},
		},
	})
}

func TestAccDynamoDBTableItem_updateWithRangeKey(t *testing.T) {
	var conf dynamodb.GetItemOutput

	tableName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	hashKey := "hashKey"
	rangeKey := "rangeKey"

	itemBefore := `{
	"hashKey": {"S": "before"},
	"rangeKey": {"S": "rangeBefore"},
	"value": {"S": "valueBefore"}
}`
	itemAfter := `{
	"hashKey": {"S": "before"},
	"rangeKey": {"S": "rangeAfter"},
	"value": {"S": "valueAfter"}
}`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dynamodb.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckItemDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccItemWithRangeKeyConfig(tableName, hashKey, rangeKey, itemBefore),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemBefore+"\n"),
				),
			},
			{
				Config: testAccItemWithRangeKeyConfig(tableName, hashKey, rangeKey, itemAfter),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableItemExists("aws_dynamodb_table_item.test", &conf),
					testAccCheckTableItemCount(tableName, 1),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "hash_key", hashKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "range_key", rangeKey),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "table_name", tableName),
					resource.TestCheckResourceAttr("aws_dynamodb_table_item.test", "item", itemAfter+"\n"),
				),
			},
		},
	})
}

func testAccCheckItemDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dynamodb_table_item" {
			continue
		}

		attrs := rs.Primary.Attributes
		attributes, err := tfdynamodb.ExpandTableItemAttributes(attrs["item"])
		if err != nil {
			return err
		}

		result, err := conn.GetItem(&dynamodb.GetItemInput{
			TableName:                aws.String(attrs["table_name"]),
			ConsistentRead:           aws.Bool(true),
			Key:                      tfdynamodb.BuildTableItemqueryKey(attributes, attrs["hash_key"], attrs["range_key"]),
			ProjectionExpression:     tfdynamodb.BuildProjectionExpression(attributes),
			ExpressionAttributeNames: tfdynamodb.BuildExpressionAttributeNames(attributes),
		})
		if err != nil {
			if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return fmt.Errorf("Error retrieving DynamoDB table item: %s", err)
		}
		if result.Item == nil {
			return nil
		}

		return fmt.Errorf("DynamoDB table item %s still exists.", rs.Primary.ID)
	}

	return nil
}

func testAccCheckTableItemExists(n string, item *dynamodb.GetItemOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No DynamoDB table item ID specified!")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn

		attrs := rs.Primary.Attributes
		attributes, err := tfdynamodb.ExpandTableItemAttributes(attrs["item"])
		if err != nil {
			return err
		}

		result, err := conn.GetItem(&dynamodb.GetItemInput{
			TableName:                aws.String(attrs["table_name"]),
			ConsistentRead:           aws.Bool(true),
			Key:                      tfdynamodb.BuildTableItemqueryKey(attributes, attrs["hash_key"], attrs["range_key"]),
			ProjectionExpression:     tfdynamodb.BuildProjectionExpression(attributes),
			ExpressionAttributeNames: tfdynamodb.BuildExpressionAttributeNames(attributes),
		})
		if err != nil {
			return fmt.Errorf("Problem getting table item '%s': %s", rs.Primary.ID, err)
		}

		*item = *result

		return nil
	}
}

func testAccCheckTableItemCount(tableName string, count int64) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).DynamoDBConn
		out, err := conn.Scan(&dynamodb.ScanInput{
			ConsistentRead: aws.Bool(true),
			TableName:      aws.String(tableName),
			Select:         aws.String(dynamodb.SelectCount),
		})
		if err != nil {
			return err
		}
		expectedCount := count
		if *out.Count != expectedCount {
			return fmt.Errorf("Expected %d items, got %d", expectedCount, *out.Count)
		}
		return nil
	}
}

func testAccItemBasicConfig(tableName, hashKey, item string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "%s"

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, hashKey, item)
}

func testAccItemWithRangeKeyConfig(tableName, hashKey, rangeKey, item string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "%s"
  range_key      = "%s"

  attribute {
    name = "%s"
    type = "S"
  }

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  range_key  = aws_dynamodb_table.test.range_key

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, rangeKey, hashKey, rangeKey, item)
}

func testAccItemWithMultipleItemsConfig(tableName, hashKey, rangeKey, firstItem, secondItem string) string {
	return fmt.Sprintf(`
resource "aws_dynamodb_table" "test" {
  name           = "%s"
  read_capacity  = 10
  write_capacity = 10
  hash_key       = "%s"
  range_key      = "%s"

  attribute {
    name = "%s"
    type = "S"
  }

  attribute {
    name = "%s"
    type = "S"
  }
}

resource "aws_dynamodb_table_item" "test1" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  range_key  = aws_dynamodb_table.test.range_key

  item = <<ITEM
%s
ITEM
}

resource "aws_dynamodb_table_item" "test2" {
  table_name = aws_dynamodb_table.test.name
  hash_key   = aws_dynamodb_table.test.hash_key
  range_key  = aws_dynamodb_table.test.range_key

  item = <<ITEM
%s
ITEM
}
`, tableName, hashKey, rangeKey, hashKey, rangeKey, firstItem, secondItem)
}
