// Code generated by internal/generate/tags/main.go; DO NOT EDIT.
package mediaconvert

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediaconvert"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
)

// ListTags lists mediaconvert service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func ListTags(conn *mediaconvert.MediaConvert, identifier string) (tftags.KeyValueTags, error) {
	input := &mediaconvert.ListTagsForResourceInput{
		Arn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(input)

	if err != nil {
		return tftags.New(nil), err
	}

	return KeyValueTags(output.ResourceTags.Tags), nil
}

// map[string]*string handling

// Tags returns mediaconvert service tags.
func Tags(tags tftags.KeyValueTags) map[string]*string {
	return aws.StringMap(tags.Map())
}

// KeyValueTags creates KeyValueTags from mediaconvert service tags.
func KeyValueTags(tags map[string]*string) tftags.KeyValueTags {
	return tftags.New(tags)
}

// UpdateTags updates mediaconvert service tags.
// The identifier is typically the Amazon Resource Name (ARN), although
// it may also be a different identifier depending on the service.
func UpdateTags(conn *mediaconvert.MediaConvert, identifier string, oldTagsMap interface{}, newTagsMap interface{}) error {
	oldTags := tftags.New(oldTagsMap)
	newTags := tftags.New(newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := &mediaconvert.UntagResourceInput{
			Arn:     aws.String(identifier),
			TagKeys: aws.StringSlice(removedTags.IgnoreAWS().Keys()),
		}

		_, err := conn.UntagResource(input)

		if err != nil {
			return fmt.Errorf("error untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := &mediaconvert.TagResourceInput{
			Arn:  aws.String(identifier),
			Tags: Tags(updatedTags.IgnoreAWS()),
		}

		_, err := conn.TagResource(input)

		if err != nil {
			return fmt.Errorf("error tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}
