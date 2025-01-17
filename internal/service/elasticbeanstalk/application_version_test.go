package elasticbeanstalk_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/elasticbeanstalk"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
)

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_basic(t *testing.T) {
	var appVersion elasticbeanstalk.ApplicationVersionDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckApplicationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkApplicationVersionConfig(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists("aws_elastic_beanstalk_application_version.default", &appVersion),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_duplicateLabels(t *testing.T) {
	var firstAppVersion elasticbeanstalk.ApplicationVersionDescription
	var secondAppVersion elasticbeanstalk.ApplicationVersionDescription

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckApplicationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkApplicationVersionConfig_duplicateLabel(sdkacctest.RandInt()),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists("aws_elastic_beanstalk_application_version.first", &firstAppVersion),
					testAccCheckApplicationVersionExists("aws_elastic_beanstalk_application_version.second", &secondAppVersion),
				),
			},
		},
	})
}

func TestAccElasticBeanstalkApplicationVersion_BeanstalkApp_tags(t *testing.T) {
	var appVersion elasticbeanstalk.ApplicationVersionDescription
	resourceName := "aws_elastic_beanstalk_application_version.default"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, elasticbeanstalk.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckApplicationVersionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBeanstalkApplicationVersionConfigWithTags(sdkacctest.RandInt(), "test1", "test2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "test1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "test2"),
				),
			},
			{
				Config: testAccBeanstalkApplicationVersionConfigWithTags(sdkacctest.RandInt(), "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
			{
				Config: testAccBeanstalkApplicationVersionConfigWithAddTags(sdkacctest.RandInt(), "updateTest1", "updateTest2", "addTest3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
					resource.TestCheckResourceAttr(resourceName, "tags.thirdTag", "addTest3"),
				),
			},
			{
				Config: testAccBeanstalkApplicationVersionConfigWithTags(sdkacctest.RandInt(), "updateTest1", "updateTest2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckApplicationVersionExists(resourceName, &appVersion),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.firstTag", "updateTest1"),
					resource.TestCheckResourceAttr(resourceName, "tags.secondTag", "updateTest2"),
				),
			},
		},
	})
}

func testAccCheckApplicationVersionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_elastic_beanstalk_application_version" {
			continue
		}

		describeApplicationVersionOpts := &elasticbeanstalk.DescribeApplicationVersionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			VersionLabels:   []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeApplicationVersions(describeApplicationVersionOpts)
		if err == nil {
			if len(resp.ApplicationVersions) > 0 {
				return fmt.Errorf("Elastic Beanstalk Application Verson still exists.")
			}

			return nil
		}
		ec2err, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ec2err.Code() != "InvalidParameterValue" {
			return err
		}
	}

	return nil
}

func testAccCheckApplicationVersionExists(n string, app *elasticbeanstalk.ApplicationVersionDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Elastic Beanstalk Application Version is not set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElasticBeanstalkConn
		describeApplicationVersionOpts := &elasticbeanstalk.DescribeApplicationVersionsInput{
			ApplicationName: aws.String(rs.Primary.Attributes["application"]),
			VersionLabels:   []*string{aws.String(rs.Primary.ID)},
		}

		log.Printf("[DEBUG] Elastic Beanstalk Application Version TEST describe opts: %s", describeApplicationVersionOpts)

		resp, err := conn.DescribeApplicationVersions(describeApplicationVersionOpts)
		if err != nil {
			return err
		}
		if len(resp.ApplicationVersions) == 0 {
			return fmt.Errorf("Elastic Beanstalk Application Version not found.")
		}

		*app = *resp.ApplicationVersions[0]

		return nil
	}
}

func testAccBeanstalkApplicationVersionConfig(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%d"
}

resource "aws_s3_bucket_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name-%d"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = aws_elastic_beanstalk_application.default.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_bucket_object.default.id
}
`, randInt, randInt, randInt)
}

func testAccBeanstalkApplicationVersionConfig_duplicateLabel(randInt int) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%d"
}

resource "aws_s3_bucket_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "first" {
  name        = "tf-test-name-%d-first"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "first" {
  application = aws_elastic_beanstalk_application.first.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_bucket_object.default.id
}

resource "aws_elastic_beanstalk_application" "second" {
  name        = "tf-test-name-%d-second"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "second" {
  application = aws_elastic_beanstalk_application.second.name
  name        = "tf-test-version-label-%d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_bucket_object.default.id
}
`, randInt, randInt, randInt, randInt, randInt)
}

func testAccBeanstalkApplicationVersionConfigWithTags(randInt int, tag1, tag2 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%[1]d"
}

resource "aws_s3_bucket_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name-%[1]d"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = aws_elastic_beanstalk_application.default.name
  name        = "tf-test-version-label-%[1]d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_bucket_object.default.id

  tags = {
    firstTag  = "%[2]s"
    secondTag = "%[3]s"
  }
}
`, randInt, tag1, tag2)
}

func testAccBeanstalkApplicationVersionConfigWithAddTags(randInt int, tag1, tag2, tag3 string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "default" {
  bucket = "tftest.applicationversion.bucket-%[1]d"
}

resource "aws_s3_bucket_object" "default" {
  bucket = aws_s3_bucket.default.id
  key    = "beanstalk/python-v1.zip"
  source = "test-fixtures/python-v1.zip"
}

resource "aws_elastic_beanstalk_application" "default" {
  name        = "tf-test-name-%[1]d"
  description = "tf-test-desc"
}

resource "aws_elastic_beanstalk_application_version" "default" {
  application = aws_elastic_beanstalk_application.default.name
  name        = "tf-test-version-label-%[1]d"
  bucket      = aws_s3_bucket.default.id
  key         = aws_s3_bucket_object.default.id

  tags = {
    firstTag  = "%[2]s"
    secondTag = "%[3]s"
    thirdTag  = "%[4]s"
  }
}
`, randInt, tag1, tag2, tag3)
}
