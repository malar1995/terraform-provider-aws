package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsEc2ClientVpnEndpoint_basic(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttr(resourceName, "client_cidr_block", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "transport_protocol", "udp"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "certificate-authentication"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "dns_name"),
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

func TestAccAwsEc2ClientVpnEndpoint_msAD(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "authentication_options.0.type", "directory-service-authentication"),
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

func TestAccAwsEc2ClientVpnEndpoint_withLogGroup(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
				),
			},

			{
				Config: testAccEc2ClientVpnEndpointConfigWithLogGroup(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "connection_log_options.0.enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "connection_log_options.0.cloudwatch_log_group"),
					resource.TestCheckResourceAttrSet(resourceName, "connection_log_options.0.cloudwatch_log_stream"),
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

func TestAccAwsEc2ClientVpnEndpoint_withDNSServers(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
				),
			},

			{
				Config: testAccEc2ClientVpnEndpointConfigWithDNSServers(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withNetworkAssociation(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithNetworkAssociation(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "network_association.#", "1"),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withAuthorizationRules(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithAuthorizationRules(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "authorization_rule.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "authorization_rule.2707333359.description"),
					resource.TestCheckResourceAttrSet(resourceName, "authorization_rule.2707333359.target_network_cidr"),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_withRoutes(t *testing.T) {
	rStr := acctest.RandString(5)
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithRoute(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithRoutes(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfigWithRoute(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnEndpoint_tags(t *testing.T) {
	resourceName := "aws_ec2_client_vpn_endpoint.test"
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAwsEc2ClientVpnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnEndpointConfig_tags(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig_tagsChanged(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: testAccEc2ClientVpnEndpointConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_client_vpn_endpoint" {
			continue
		}

		input := &ec2.DescribeClientVpnEndpointsInput{
			ClientVpnEndpointIds: []*string{aws.String(rs.Primary.ID)},
		}

		resp, _ := conn.DescribeClientVpnEndpoints(input)
		for _, v := range resp.ClientVpnEndpoints {
			if *v.ClientVpnEndpointId == rs.Primary.ID && !(*v.Status.Code == ec2.ClientVpnEndpointStatusCodeDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Client VPN endpoint (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsEc2ClientVpnEndpointExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

const testAccEc2ClientVpnEndpointBaseConfig = `
data "aws_availability_zones" "available" {}

resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}
`

func testAccEc2ClientVpnEndpointConfig(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithLogGroup(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "lg" {
  name = "terraform-testacc-clientvpn-loggroup-%s"
}

resource "aws_cloudwatch_log_stream" "ls" {
  name           = "${aws_cloudwatch_log_group.lg.name}-stream"
  log_group_name = "${aws_cloudwatch_log_group.lg.name}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled               = true
    cloudwatch_log_group  = "${aws_cloudwatch_log_group.lg.name}"
    cloudwatch_log_stream = "${aws_cloudwatch_log_stream.ls.name}"
  }
}
`, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithDNSServers(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  dns_servers = ["8.8.8.8", "8.8.4.4"]

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithMicrosoftAD(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block  = "10.0.0.0/16"
}

resource "aws_subnet" "test1" {
  vpc_id     				= "${aws_vpc.test.id}"
  cidr_block 				= "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
}

resource "aws_subnet" "test2" {
  vpc_id     				= "${aws_vpc.test.id}"
  cidr_block 				= "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
}

resource "aws_directory_service_directory" "test" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  vpc_settings {
    vpc_id = "${aws_vpc.test.id}"
    subnet_ids = ["${aws_subnet.test1.id}", "${aws_subnet.test2.id}"]
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                = "directory-service-authentication"
    active_directory_id = "${aws_directory_service_directory.test.id}"
  }

  connection_log_options {
    enabled = false
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfigWithNetworkAssociation(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-subnet-%s"
	}
}
resource "aws_subnet" "test" {
	cidr_block         = "10.1.1.0/24"
	vpc_id             = "${aws_vpc.test.id}"
	availability_zone  = "${data.aws_availability_zones.available.names[0]}"
	tags = {
		Name = "tf-acc-subnet-%s"
	}
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
	}
	
  network_association {
		subnet_id = "${aws_subnet.test.id}"
	}
}
`, rName, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithAuthorizationRules(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_vpc" "test" {
	cidr_block = "10.1.0.0/16"
	tags = {
		Name = "terraform-testacc-subnet-%s"
	}
}
resource "aws_subnet" "test" {
	cidr_block        = "10.1.1.0/24"
	vpc_id            = "${aws_vpc.test.id}"
	availability_zone = "${data.aws_availability_zones.available.names[0]}"
	tags = {
		Name = "tf-acc-subnet-%s"
	}
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
	}
	
  network_association {
		subnet_id = "${aws_subnet.test.id}"
	}

	authorization_rule {
		description          = "example auth rule"
		target_network_cidr  = "10.1.1.0/24"
	}
}
`, rName, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithRoute(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-subnet-%s"
  }
}
resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = "${aws_vpc.test.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-subnet-%s"
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }

  network_association {
    subnet_id = "${aws_subnet.test.id}"
  }

  authorization_rule {
    description          = "example auth rule"
    target_network_cidr  = "10.1.1.0/24"
  }

  route {
    description               = "example route 1"
    subnet_id                 = "${aws_subnet.test.id}"
    destination_network_cidr  = "192.168.1.0/24"
  }
}
`, rName, rName, rName)
}

func testAccEc2ClientVpnEndpointConfigWithRoutes(rName string) string {
	return testAccEc2ClientVpnEndpointBaseConfig + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
  tags = {
    Name = "terraform-testacc-subnet-%s"
  }
}
resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = "${aws_vpc.test.id}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-subnet-%s"
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }

  network_association {
    subnet_id = "${aws_subnet.test.id}"
  }

  authorization_rule {
    description          = "example auth rule"
    target_network_cidr  = "10.1.1.0/24"
  }

  route {
    description               = "example route 1"
    subnet_id                 = "${aws_subnet.test.id}"
    destination_network_cidr  = "192.168.1.0/24"
  }

  route {
    description               = "example route 2"
    subnet_id                 = "${aws_subnet.test.id}"
    destination_network_cidr  = "192.168.2.0/24"
  }
}
`, rName, rName, rName)
}

func testAccEc2ClientVpnEndpointConfig_tags(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block = "10.0.0.0/16"

  authentication_options {
    type = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Environment = "production"
    Usage = "original"
  }
}
`, rName)
}

func testAccEc2ClientVpnEndpointConfig_tagsChanged(rName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "example" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "example" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.example.private_key_pem}"

  subject {
    common_name  = "example.com"
    organization = "ACME Examples, Inc"
  }

  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "aws_acm_certificate" "cert" {
  private_key      = "${tls_private_key.example.private_key_pem}"
  certificate_body = "${tls_self_signed_cert.example.cert_pem}"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.cert.arn}"
  client_cidr_block = "10.0.0.0/16"

  authentication_options {
    type = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.cert.arn}"
  }

  connection_log_options {
    enabled = false
  }

  tags = {
    Usage = "changed"
  }
}
`, rName)
}
