// Code generated by internal/generate/tagresource/main.go; DO NOT EDIT.

package ecs

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tftags "github.com/nij4t/terraform-provider-aws/internal/tags"
	"github.com/nij4t/terraform-provider-aws/internal/tfresource"
)

func ResourceTag() *schema.Resource {
	return &schema.Resource{
		Create: resourceTagCreate,
		Read:   resourceTagRead,
		Update: resourceTagUpdate,
		Delete: resourceTagDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"resource_arn": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"value": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func resourceTagCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn

	identifier := d.Get("resource_arn").(string)
	key := d.Get("key").(string)
	value := d.Get("value").(string)

	if err := UpdateTags(conn, identifier, nil, map[string]string{key: value}); err != nil {
		return fmt.Errorf("error creating %s resource (%s) tag (%s): %w", ecs.ServiceID, identifier, key, err)
	}

	d.SetId(tftags.SetResourceID(identifier, key))

	return resourceTagRead(d, meta)
}

func resourceTagRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	value, err := GetTag(conn, identifier, key)

	if !d.IsNewResource() && tfresource.NotFound(err) {
		log.Printf("[WARN] %s resource (%s) tag (%s) not found, removing from state", ecs.ServiceID, identifier, key)
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading %s resource (%s) tag (%s): %w", ecs.ServiceID, identifier, key, err)
	}

	d.Set("resource_arn", identifier)
	d.Set("key", key)
	d.Set("value", value)

	return nil
}

func resourceTagUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := UpdateTags(conn, identifier, nil, map[string]string{key: d.Get("value").(string)}); err != nil {
		return fmt.Errorf("error updating %s resource (%s) tag (%s): %w", ecs.ServiceID, identifier, key, err)
	}

	return resourceTagRead(d, meta)
}

func resourceTagDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).ECSConn
	identifier, key, err := tftags.GetResourceID(d.Id())

	if err != nil {
		return err
	}

	if err := UpdateTags(conn, identifier, map[string]string{key: d.Get("value").(string)}, nil); err != nil {
		return fmt.Errorf("error deleting %s resource (%s) tag (%s): %w", ecs.ServiceID, identifier, key, err)
	}

	return nil
}
