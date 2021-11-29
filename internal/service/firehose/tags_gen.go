// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package firehose

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
)

// ListTags lists firehose service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn *firehose.Firehose, identifier string) (tftags.KeyValueTags, error) {
	input := &firehose.ListTagsForDeliveryStreamInput{
		DeliveryStreamName: aws.String(identifier),
	}

	output, err := conn.ListTagsForDeliveryStream(input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.Tags), nil
}

// []*SERVICE.Tag handling

// Tags returns firehose service tags.
func Tags(tags tftags.KeyValueTags) []*firehose.Tag {
	result := make([]*firehose.Tag, 0, len(tags))

	for k, v := range tags.Map() {
		tag := &firehose.Tag{
			Key:   aws.String(k),
			Value: aws.String(v),
		}

		result = append(result, tag)
	}

	return result
}

// KeyValueTags creates tftags.KeyValueTags from firehose service tags.
func KeyValueTags(tags []*firehose.Tag) tftags.KeyValueTags {
	m := make(map[string]*string, len(tags))

	for _, tag := range tags {
		m[aws.StringValue(tag.Key)] = tag.Value
	}

	return tftags.New(m)
}

// UpdateTags updates firehose service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *firehose.Firehose, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &firehose.UntagDeliveryStreamInput{
			DeliveryStreamName: aws.String(identifier),
			TagKeys:            aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.UntagDeliveryStream(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &firehose.TagDeliveryStreamInput{
			DeliveryStreamName: aws.String(identifier),
			Tags:               Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.TagDeliveryStream(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
