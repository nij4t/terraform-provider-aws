package ec2_test

import (
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfec2 "github.com/nij4t/terraform-provider-aws/internal/service/ec2"
	tfs3 "github.com/nij4t/terraform-provider-aws/internal/service/s3"
)

func TestAccEC2EBSSnapshotImport_basic(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEBSSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportBasicConfig(rName, t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_tags(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEBSSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportTags1Config(rName, t, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				Config: testAccEBSSnapshotImportTags2Config(rName, t, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				Config: testAccEBSSnapshotImportTags1Config(rName, t, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_disappears(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEBSSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportBasicConfig(rName, t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceEBSSnapshotImport(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2EBSSnapshotImport_Disappears_s3BucketObject(t *testing.T) {
	var v ec2.Snapshot
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parentResourceName := "aws_s3_bucket_object.image"
	resourceName := "aws_ebs_snapshot_import.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEBSSnapshotImportDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEBSSnapshotImportBasicConfig(rName, t),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSnapshotImportExists(resourceName, &v),
					acctest.MatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "ec2", regexp.MustCompile(`snapshot/snap-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					acctest.CheckResourceAttrAccountID(resourceName, "owner_id"),
					acctest.CheckResourceDisappears(acctest.Provider, tfs3.ResourceBucketObject(), parentResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSnapshotImportExists(n string, v *ec2.Snapshot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		request := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		response, err := conn.DescribeSnapshots(request)
		if err == nil {
			if response.Snapshots != nil && len(response.Snapshots) > 0 {
				*v = *response.Snapshots[0]
				return nil
			}
		}
		return fmt.Errorf("Error finding EC2 Snapshot %s", rs.Primary.ID)
	}
}

func testAccCheckEBSSnapshotImportDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ebs_snapshot_import" {
			continue
		}
		input := &ec2.DescribeSnapshotsInput{
			SnapshotIds: []*string{aws.String(rs.Primary.ID)},
		}

		output, err := conn.DescribeSnapshots(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, "InvalidSnapshot.NotFound", "") {
				continue
			}
			return err
		}
		if output != nil && len(output.Snapshots) > 0 && aws.StringValue(output.Snapshots[0].SnapshotId) == rs.Primary.ID {
			return fmt.Errorf("EBS Snapshot %q still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccEBSSnapshotImportConfig_Base(t *testing.T) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "images" {
  bucket_prefix = "images-"
  force_destroy = true
}

resource "aws_s3_bucket_object" "image" {
  bucket         = aws_s3_bucket.images.id
  key            = "diskimage.vhd"
  content_base64 = %[1]q
}

# The following resources are for the *vmimport service user*
# See: https://docs.aws.amazon.com/vm-import/latest/userguide/vmie_prereqs.html#vmimport-role
resource "aws_iam_role" "vmimport" {
  assume_role_policy = data.aws_iam_policy_document.vmimport-trust.json
}

resource "aws_iam_role_policy" "vmimport-access" {
  role   = aws_iam_role.vmimport.id
  policy = data.aws_iam_policy_document.vmimport-access.json
}

data "aws_iam_policy_document" "vmimport-access" {
  statement {
    effect = "Allow"
    actions = [
      "s3:GetBucketLocation",
      "s3:GetObject",
      "s3:ListBucket",
    ]
    resources = [
      aws_s3_bucket.images.arn,
      "${aws_s3_bucket.images.arn}/*"
    ]
  }
  statement {
    effect = "Allow"
    actions = [
      "ec2:ModifySnapshotAttribute",
      "ec2:CopySnapshot",
      "ec2:RegisterImage",
      "ec2:Describe*"
    ]
    resources = [
      "*"
    ]
  }
}


data "aws_iam_policy_document" "vmimport-trust" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["vmie.amazonaws.com"]
    }

    actions = [
      "sts:AssumeRole"
    ]

    condition {
      test     = "StringEquals"
      variable = "sts:ExternalId"
      values   = ["vmimport"]
    }
  }
}
`, testAccEBSSnapshotDisk(t))
}

func testAccEBSSnapshotImportBasicConfig(rName string, t *testing.T) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportConfig_Base(t), fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = %[1]q
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.images.id
      s3_key    = aws_s3_bucket_object.image.key
    }
  }

  role_name = aws_iam_role.vmimport.name

  timeouts {
    create = "10m"
    delete = "10m"
  }
}
`, rName))
}

func testAccEBSSnapshotImportTags1Config(rName string, t *testing.T, tagKey1 string, tagValue1 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportConfig_Base(t), fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = %[1]q
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.images.id
      s3_key    = aws_s3_bucket_object.image.key
    }
  }

  role_name = aws_iam_role.vmimport.name

  timeouts {
    create = "10m"
    delete = "10m"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccEBSSnapshotImportTags2Config(rName string, t *testing.T, tagKey1 string, tagValue1 string, tagKey2 string, tagValue2 string) string {
	return acctest.ConfigCompose(testAccEBSSnapshotImportConfig_Base(t), fmt.Sprintf(`
resource "aws_ebs_snapshot_import" "test" {
  disk_container {
    description = %[1]q
    format      = "VHD"
    user_bucket {
      s3_bucket = aws_s3_bucket.images.id
      s3_key    = aws_s3_bucket_object.image.key
    }
  }

  role_name = aws_iam_role.vmimport.name

  timeouts {
    create = "10m"
    delete = "10m"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccEBSSnapshotDisk(t *testing.T) string {
	// Take a compressed then base64'd disk image,
	// base64 decode, then decompress, then re-base64
	// the image, so it can be uploaded to s3.

	// little vmdk built by:
	// $ VBoxManage createmedium disk --filename ./image.vhd --sizebytes 512 --format vhd
	// $ cat image.vhd | gzip --best | base64
	b64_compressed := "H4sIAAAAAAACA0vOz0tNLsmsYGBgYGJgZIACJgZ1789hZUn5FQxsDIzhmUbZMHEEzSIIJJj///+QlV1rMXFVnLzHwteXYmWDDfYxjIIhA5IrigsSi4pT/0MBRJSNAZoWGBkUGBj+//9SNhpSo2AUDD+AyPOjYESW/6P1/4gGAAvDpVcACgAA"

	decoder := base64.NewDecoder(base64.StdEncoding, strings.NewReader(b64_compressed))

	zr, err := gzip.NewReader(decoder)
	if err != nil {
		t.Fatal(err)
	}

	var out strings.Builder
	encoder := base64.NewEncoder(base64.StdEncoding, &out)

	_, err = io.Copy(encoder, zr)
	if err != nil {
		t.Fatal(err)
	}
	encoder.Close()

	return out.String()
}
