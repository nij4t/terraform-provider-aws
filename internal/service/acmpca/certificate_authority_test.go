package acmpca_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/nij4t/terraform-provider-aws/internal/acctest"
	"github.com/nij4t/terraform-provider-aws/internal/conns"
	tfacmpca "github.com/nij4t/terraform-provider-aws/internal/service/acmpca"
)

func TestAccACMPCACertificateAuthority_basic(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_Required(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.key_algorithm", "RSA_4096"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.signing_algorithm", "SHA512WITHRSA"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.0.common_name", commonName),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "certificate_chain", ""),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_after"),
					acctest.CheckResourceAttrRFC3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "permanent_deletion_time_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "serial", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "PENDING_CERTIFICATE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "SUBORDINATE"),
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

func TestAccACMPCACertificateAuthority_disappears(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_Required(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					acctest.CheckResourceDisappears(acctest.Provider, tfacmpca.ResourceCertificateAuthority(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_enabled(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_Enabled(commonName, acmpca.CertificateAuthorityTypeRoot, true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "type", acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusPendingCertificate),
					acctest.CheckACMPCACertificateAuthorityActivateCA(&certificateAuthority),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_Enabled(commonName, acmpca.CertificateAuthorityTypeRoot, true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "type", acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusActive),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_Enabled(commonName, acmpca.CertificateAuthorityTypeRoot, false),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_deleteFromActiveState(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_WithRootCertificate(commonName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "type", acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					// Since the status of the CA is changed by importing the certificate in
					// aws_acmpca_certificate_authority_certificate, the value of `status` is no longer accurate
					// resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusActive),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_customCNAME(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"

	domain := acctest.RandomDomain()
	commonName := domain.String()
	customCName := domain.Subdomain("crl").String()
	customCName2 := domain.Subdomain("crl2").String()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCNAME(rName, commonName, customCName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", customCName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCNAME(rName, commonName, customCName2),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", customCName2),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing custom cname on resource update
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, commonName, true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test adding custom cname on resource update
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCNAME(rName, commonName, customCName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", customCName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_Required(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_enabled(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, commonName, true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test disabling revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, commonName, false),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
			// Test enabling revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, commonName, true),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_Required(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_expirationInDays(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_ExpirationInDays(rName, commonName, 1),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", "PUBLIC_READ"),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_ExpirationInDays(rName, commonName, 2),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "2"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccCertificateAuthorityConfig_Required(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_RevocationCrl_s3ObjectACL(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_s3ObjectACL(rName, commonName, "BUCKET_OWNER_FULL_CONTROL"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", "BUCKET_OWNER_FULL_CONTROL"),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_s3ObjectACL(rName, commonName, "PUBLIC_READ"),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_object_acl", "PUBLIC_READ"),
				),
			},
		},
	})
}

func TestAccACMPCACertificateAuthority_tags(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	commonName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, acmpca.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateAuthorityConfig_Tags_Single(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_Tags_SingleUpdated(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value-updated"),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_Tags_Multiple(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccCertificateAuthorityConfig_Tags_Single(commonName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckACMPCACertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func testAccCheckCertificateAuthorityDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ACMPCAConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_certificate_authority" {
			continue
		}

		input := &acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCertificateAuthority(input)

		if err != nil {
			if tfawserr.ErrMessageContains(err, acmpca.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if output != nil && output.CertificateAuthority != nil && aws.StringValue(output.CertificateAuthority.Arn) == rs.Primary.ID && aws.StringValue(output.CertificateAuthority.Status) != acmpca.CertificateAuthorityStatusDeleted {
			return fmt.Errorf("ACM PCA Certificate Authority %q still exists in non-DELETED state: %s", rs.Primary.ID, aws.StringValue(output.CertificateAuthority.Status))
		}
	}

	return nil
}

func testAccCertificateAuthorityConfig_Enabled(commonName, certificateAuthorityType string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  enabled                         = %[1]t
  permanent_deletion_time_in_days = 7
  type                            = %[2]q

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[3]q
    }
  }
}
`, enabled, certificateAuthorityType, commonName)
}

func testAccCertificateAuthorityConfig_WithRootCertificate(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}

resource "aws_acmpca_certificate_authority_certificate" "test" {
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn

  certificate       = aws_acmpca_certificate.test.certificate
  certificate_chain = aws_acmpca_certificate.test.certificate_chain
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA512WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

data "aws_partition" "current" {}
`, commonName)
}

func testAccCertificateAuthorityConfig_Required(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }
}
`, commonName)
}

func testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCNAME(rName, commonName, customCname string) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      custom_cname       = %[2]q
      enabled            = true
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, commonName, customCname))
}

func testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, commonName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = %[2]t
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }
}
`, commonName, enabled))
}

func testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_ExpirationInDays(rName, commonName string, expirationInDays int) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = true
      expiration_in_days = %[2]d
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }
}
`, commonName, expirationInDays))
}

func testAccCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_s3ObjectACL(rName, commonName, s3ObjectAcl string) string {
	return acctest.ConfigCompose(
		testAccCertificateAuthorityConfig_S3Bucket(rName),
		fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = true
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
      s3_object_acl      = %[2]q
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, commonName, s3ObjectAcl))
}

func testAccCertificateAuthorityConfig_S3Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

data "aws_iam_policy_document" "acmpca_bucket_access" {
  statement {
    actions = [
      "s3:GetBucketAcl",
      "s3:GetBucketLocation",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      identifiers = ["acm-pca.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.acmpca_bucket_access.json
}
`, rName)
}

func testAccCertificateAuthorityConfig_Tags_Single(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  tags = {
    tag1 = "tag1value"
  }
}
`, commonName)
}

func testAccCertificateAuthorityConfig_Tags_SingleUpdated(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  tags = {
    tag1 = "tag1value-updated"
  }
}
`, commonName)
}

func testAccCertificateAuthorityConfig_Tags_Multiple(commonName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = %[1]q
    }
  }

  tags = {
    tag1 = "tag1value"
    tag2 = "tag2value"
  }
}
`, commonName)
}
