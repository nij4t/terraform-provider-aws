// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package dynamodb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

// GetTag fetches an individual dynamodb service tag for a resource.
// Returns whether the key value and any errors. A NotFoundError is used to signal that no value was found.
// This function will optimise the handling over ListTags, if possible.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func GetTag(conn *dynamodb.DynamoDB, identifier string, key string) (*string, error) {
	listTags, err := ListTags(conn, identifier)

	if err != nil {
		return nil, err
	}

	if !listTags.KeyExists(key) {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return listTags.KeyValue(key), nil
}

// ListTags lists dynamodb service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn *dynamodb.DynamoDB, identifier string) (tftags.KeyValueTags, error) {
	input := &dynamodb.ListTagsOfResourceInput{
		ResourceArn: aws.String(identifier),
	}

	output, err := conn.ListTagsOfResource(input)

	if tfawserr.ErrCodeEquals(err, "ResourceNotFoundException") {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.Tags), nil
}

// []*SERVICE.Tag handling

// Tags returns dynamodb service tags.
func Tags(tags tftags.KeyValueTags) []*dynamodb.Tag {
	result := make([]*dynamodb.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &dynamodb.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from dynamodb service tags.
func KeyValueTags(tags []*dynamodb.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates dynamodb service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *dynamodb.DynamoDB, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &dynamodb.UntagResourceInput{
			ResourceArn: aws.String(identifier),
			TagKeys:     aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.UntagResource(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &dynamodb.TagResourceInput{
			ResourceArn: aws.String(identifier),
			Tags:        Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.TagResource(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
