package dynamodb

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	"github.com/nij4t/terraform-provider-aws/internal/verify"
)

func ResourceTableItem() *schema.Resource {
	return &schema.Resource{
		Create: resourceTableItemCreate,
		Read:   resourceTableItemRead,
		Update: resourceTableItemUpdate,
		Delete: resourceTableItemDelete,

		Schema: map[string]*schema.Schema{
			"table_name": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"hash_key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"range_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"item": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateDynamoDbTableItem,
			},
		},
	}
}

func validateDynamoDbTableItem(v interface{}, k string) (ws []string, errors []error) {
	_, err := ExpandTableItemAttributes(v.(string))
	if err != nil {
		errors = append(errors, fmt.Errorf("Invalid format of %q: %s", k, err))
	}
	return
}

func resourceTableItemCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	item := d.Get("item").(string)
	attributes, err := ExpandTableItemAttributes(item)
	if err != nil {
		return err
	}

	log.Printf("[DEBUG] DynamoDB item create: %s", tableName)

	_, err = conn.PutItem(&dynamodb.PutItemInput{
		Item: attributes,
		// Explode if item exists. We didn't create it.
		Expected: map[string]*dynamodb.ExpectedAttributeValue{
			hashKey: {
				Exists: aws.Bool(false),
			},
		},
		TableName: aws.String(tableName),
	})
	if err != nil {
		return err
	}

	rangeKey := d.Get("range_key").(string)
	id := buildDynamoDbTableItemId(tableName, hashKey, rangeKey, attributes)

	d.SetId(id)

	return resourceTableItemRead(d, meta)
}

func resourceTableItemUpdate(d *schema.ResourceData, meta interface{}) error {
	log.Printf("[DEBUG] Updating DynamoDB table %s", d.Id())
	conn := meta.(*conns.AWSClient).DynamoDBConn

	if d.HasChange("item") {
		tableName := d.Get("table_name").(string)
		hashKey := d.Get("hash_key").(string)
		rangeKey := d.Get("range_key").(string)

		oldItem, newItem := d.GetChange("item")

		attributes, err := ExpandTableItemAttributes(newItem.(string))
		if err != nil {
			return err
		}
		newQueryKey := BuildTableItemqueryKey(attributes, hashKey, rangeKey)

		updates := map[string]*dynamodb.AttributeValueUpdate{}
		for key, value := range attributes {
			// Hash keys and range keys are not updatable, so we'll basically create
			// a new record and delete the old one below
			if key == hashKey || key == rangeKey {
				continue
			}
			updates[key] = &dynamodb.AttributeValueUpdate{
				Action: aws.String(dynamodb.AttributeActionPut),
				Value:  value,
			}
		}

		_, err = conn.UpdateItem(&dynamodb.UpdateItemInput{
			AttributeUpdates: updates,
			TableName:        aws.String(tableName),
			Key:              newQueryKey,
		})
		if err != nil {
			return err
		}

		oItem := oldItem.(string)
		oldAttributes, err := ExpandTableItemAttributes(oItem)
		if err != nil {
			return err
		}

		// New record is created via UpdateItem in case we're changing hash key
		// so we need to get rid of the old one
		oldQueryKey := BuildTableItemqueryKey(oldAttributes, hashKey, rangeKey)
		if !reflect.DeepEqual(oldQueryKey, newQueryKey) {
			log.Printf("[DEBUG] Deleting old record: %#v", oldQueryKey)
			_, err := conn.DeleteItem(&dynamodb.DeleteItemInput{
				Key:       oldQueryKey,
				TableName: aws.String(tableName),
			})
			if err != nil {
				return err
			}
		}

		id := buildDynamoDbTableItemId(tableName, hashKey, rangeKey, attributes)
		d.SetId(id)
	}

	return resourceTableItemRead(d, meta)
}

func resourceTableItemRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	log.Printf("[DEBUG] Loading data for DynamoDB table item '%s'", d.Id())

	tableName := d.Get("table_name").(string)
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	attributes, err := ExpandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return err
	}

	result, err := conn.GetItem(&dynamodb.GetItemInput{
		TableName:                aws.String(tableName),
		ConsistentRead:           aws.Bool(true),
		Key:                      BuildTableItemqueryKey(attributes, hashKey, rangeKey),
		ProjectionExpression:     BuildProjectionExpression(attributes),
		ExpressionAttributeNames: BuildExpressionAttributeNames(attributes),
	})
	if err != nil {
		if tfawserr.ErrMessageContains(err, dynamodb.ErrCodeResourceNotFoundException, "") {
			log.Printf("[WARN] Dynamodb Table Item (%s) not found, error code (404)", d.Id())
			d.SetId("")
			return nil
		}

		return fmt.Errorf("Error retrieving DynamoDB table item: %s", err)
	}

	if result.Item == nil {
		log.Printf("[WARN] Dynamodb Table Item (%s) not found", d.Id())
		d.SetId("")
		return nil
	}

	// The record exists, now test if it differs from what is desired
	if !reflect.DeepEqual(result.Item, attributes) {
		itemAttrs, err := flattenDynamoDBTableItemAttributes(result.Item)
		if err != nil {
			return err
		}
		d.Set("item", itemAttrs)
		id := buildDynamoDbTableItemId(tableName, hashKey, rangeKey, result.Item)
		d.SetId(id)
	}

	return nil
}

func resourceTableItemDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).DynamoDBConn

	attributes, err := ExpandTableItemAttributes(d.Get("item").(string))
	if err != nil {
		return err
	}
	hashKey := d.Get("hash_key").(string)
	rangeKey := d.Get("range_key").(string)
	queryKey := BuildTableItemqueryKey(attributes, hashKey, rangeKey)

	_, err = conn.DeleteItem(&dynamodb.DeleteItemInput{
		Key:       queryKey,
		TableName: aws.String(d.Get("table_name").(string)),
	})
	return err
}

// Helpers

func BuildExpressionAttributeNames(attrs map[string]*dynamodb.AttributeValue) map[string]*string {
	names := map[string]*string{}
	for key := range attrs {
		names["#a_"+key] = aws.String(key)
	}

	return names
}

func BuildProjectionExpression(attrs map[string]*dynamodb.AttributeValue) *string {
	keys := []string{}
	for key := range attrs {
		keys = append(keys, key)
	}
	return aws.String("#a_" + strings.Join(keys, ", #a_"))
}

func buildDynamoDbTableItemId(tableName string, hashKey string, rangeKey string, attrs map[string]*dynamodb.AttributeValue) string {
	id := []string{tableName, hashKey}

	if hashVal, ok := attrs[hashKey]; ok {
		id = append(id, verify.Base64Encode(hashVal.B))
		id = append(id, aws.StringValue(hashVal.S))
		id = append(id, aws.StringValue(hashVal.N))
	}
	if rangeVal, ok := attrs[rangeKey]; ok && rangeKey != "" {
		id = append(id, rangeKey, verify.Base64Encode(rangeVal.B))
		id = append(id, aws.StringValue(rangeVal.S))
		id = append(id, aws.StringValue(rangeVal.N))
	}
	return strings.Join(id, "|")
}

func BuildTableItemqueryKey(attrs map[string]*dynamodb.AttributeValue, hashKey string, rangeKey string) map[string]*dynamodb.AttributeValue {
	queryKey := map[string]*dynamodb.AttributeValue{
		hashKey: attrs[hashKey],
	}
	if rangeKey != "" {
		queryKey[rangeKey] = attrs[rangeKey]
	}

	return queryKey
}
