package mediastore_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mediastore"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccMediaStoreContainerPolicy_basic(t *testing.T) {
	rname := sdkacctest.RandString(5)
	resourceName := "aws_media_store_container_policy.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, mediastore.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckContainerPolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccMediaStoreContainerPolicyConfig(rname, sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccMediaStoreContainerPolicyConfig(rname, sdkacctest.RandString(5)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckContainerPolicyExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "container_name"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func testAccCheckContainerPolicyDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_media_store_container_policy" {
			continue
		}

		input := &mediastore.GetContainerPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerPolicy(input)
		if err != nil {
			if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerNotFoundException, "") {
				return nil
			}
			if tfawserr.ErrMessageContains(err, mediastore.ErrCodePolicyNotFoundException, "") {
				return nil
			}
			if tfawserr.ErrMessageContains(err, mediastore.ErrCodeContainerInUseException, "Container must be ACTIVE in order to perform this operation") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected MediaStore Container Policy to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}

func testAccCheckContainerPolicyExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).MediaStoreConn

		input := &mediastore.GetContainerPolicyInput{
			ContainerName: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetContainerPolicy(input)

		return err
	}
}

func testAccMediaStoreContainerPolicyConfig(rName, sid string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_media_store_container" "test" {
  name = "tf_mediastore_%s"
}

resource "aws_media_store_container_policy" "test" {
  container_name = aws_media_store_container.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "%s",
      "Action": [
        "mediastore:*"
      ],
      "Principal": {
        "AWS": "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
      },
      "Effect": "Allow",
      "Resource": "arn:${data.aws_partition.current.partition}:mediastore:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:container/${aws_media_store_container.test.name}/*",
      "Condition": {
        "Bool": {
          "aws:SecureTransport": "true"
        }
      }
    }
  ]
}
EOF
}
`, rName, sid)
}
