package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/mutexkv"
	tfprovider "github.com/terraform-providers/terraform-provider-aws/aws/internal/provider"
)

// Provider returns a *schema.Provider.
func Provider() *schema.Provider {
	// TODO: Move the validation to this, requires conditional schemas
	// TODO: Move the configuration to this, requires validation

	// The actual provider
	provider := &schema.Provider{
		Schema: map[string]*schema.Schema{
			"access_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["access_key"],
			},

			"secret_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["secret_key"],
			},

			"profile": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["profile"],
			},

			"assume_role": assumeRoleSchema(),

			"shared_credentials_file": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["shared_credentials_file"],
			},

			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "",
				Description: descriptions["token"],
			},

			"region": {
				Type:     schema.TypeString,
				Required: true,
				DefaultFunc: schema.MultiEnvDefaultFunc([]string{
					"AWS_REGION",
					"AWS_DEFAULT_REGION",
				}, nil),
				Description:  descriptions["region"],
				InputDefault: "us-east-1", // lintignore:AWSAT003
			},

			"max_retries": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     25,
				Description: descriptions["max_retries"],
			},

			"allowed_account_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"forbidden_account_ids"},
				Set:           schema.HashString,
			},

			"forbidden_account_ids": {
				Type:          schema.TypeSet,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ConflictsWith: []string{"allowed_account_ids"},
				Set:           schema.HashString,
			},

			"default_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block with settings to default resource tags across all resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"tags": {
							Type:        schema.TypeMap,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Description: "Resource tags to default across all resources",
						},
					},
				},
			},

			"endpoints": endpointsSchema(),

			"ignore_tags": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Configuration block with settings to ignore resource tags across all resources.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"keys": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "Resource tag keys to ignore across all resources.",
						},
						"key_prefixes": {
							Type:        schema.TypeSet,
							Optional:    true,
							Elem:        &schema.Schema{Type: schema.TypeString},
							Set:         schema.HashString,
							Description: "Resource tag key prefixes to ignore across all resources.",
						},
					},
				},
			},

			"insecure": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["insecure"],
			},

			"skip_credentials_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_credentials_validation"],
			},

			"skip_get_ec2_platforms": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_get_ec2_platforms"],
			},

			"skip_region_validation": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_region_validation"],
			},

			"skip_requesting_account_id": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_requesting_account_id"],
			},

			"skip_metadata_api_check": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["skip_metadata_api_check"],
			},

			"s3_force_path_style": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: descriptions["s3_force_path_style"],
			},
		},

		DataSourcesMap: map[string]*schema.Resource{
			"aws_acm_certificate":                            dataSourceAwsAcmCertificate(),
			"aws_acmpca_certificate_authority":               dataSourceAwsAcmpcaCertificateAuthority(),
			"aws_acmpca_certificate":                         dataSourceAwsAcmpcaCertificate(),
			"aws_ami":                                        dataSourceAwsAmi(),
			"aws_ami_ids":                                    dataSourceAwsAmiIds(),
			"aws_api_gateway_api_key":                        dataSourceAwsApiGatewayApiKey(),
			"aws_api_gateway_domain_name":                    dataSourceAwsApiGatewayDomainName(),
			"aws_api_gateway_resource":                       dataSourceAwsApiGatewayResource(),
			"aws_api_gateway_rest_api":                       dataSourceAwsApiGatewayRestApi(),
			"aws_api_gateway_vpc_link":                       dataSourceAwsApiGatewayVpcLink(),
			"aws_apigatewayv2_api":                           dataSourceAwsApiGatewayV2Api(),
			"aws_apigatewayv2_apis":                          dataSourceAwsApiGatewayV2Apis(),
			"aws_appmesh_mesh":                               dataSourceAwsAppmeshMesh(),
			"aws_appmesh_virtual_service":                    dataSourceAwsAppmeshVirtualService(),
			"aws_arn":                                        dataSourceAwsArn(),
			"aws_autoscaling_group":                          dataSourceAwsAutoscalingGroup(),
			"aws_autoscaling_groups":                         dataSourceAwsAutoscalingGroups(),
			"aws_availability_zone":                          dataSourceAwsAvailabilityZone(),
			"aws_availability_zones":                         dataSourceAwsAvailabilityZones(),
			"aws_backup_plan":                                dataSourceAwsBackupPlan(),
			"aws_backup_selection":                           dataSourceAwsBackupSelection(),
			"aws_backup_vault":                               dataSourceAwsBackupVault(),
			"aws_batch_compute_environment":                  dataSourceAwsBatchComputeEnvironment(),
			"aws_batch_job_queue":                            dataSourceAwsBatchJobQueue(),
			"aws_billing_service_account":                    dataSourceAwsBillingServiceAccount(),
			"aws_caller_identity":                            dataSourceAwsCallerIdentity(),
			"aws_canonical_user_id":                          dataSourceAwsCanonicalUserId(),
			"aws_cloudformation_export":                      dataSourceAwsCloudFormationExport(),
			"aws_cloudformation_stack":                       dataSourceAwsCloudFormationStack(),
			"aws_cloudformation_type":                        dataSourceAwsCloudFormationType(),
			"aws_cloudfront_cache_policy":                    dataSourceAwsCloudFrontCachePolicy(),
			"aws_cloudfront_distribution":                    dataSourceAwsCloudFrontDistribution(),
			"aws_cloudfront_function":                        dataSourceAwsCloudFrontFunction(),
			"aws_cloudfront_origin_request_policy":           dataSourceAwsCloudFrontOriginRequestPolicy(),
			"aws_cloudhsm_v2_cluster":                        dataSourceCloudHsmV2Cluster(),
			"aws_cloudtrail_service_account":                 dataSourceAwsCloudTrailServiceAccount(),
			"aws_cloudwatch_event_connection":                dataSourceAwsCloudwatchEventConnection(),
			"aws_cloudwatch_event_source":                    dataSourceAwsCloudWatchEventSource(),
			"aws_cloudwatch_log_group":                       dataSourceAwsCloudwatchLogGroup(),
			"aws_codeartifact_authorization_token":           dataSourceAwsCodeArtifactAuthorizationToken(),
			"aws_codeartifact_repository_endpoint":           dataSourceAwsCodeArtifactRepositoryEndpoint(),
			"aws_cognito_user_pools":                         dataSourceAwsCognitoUserPools(),
			"aws_codecommit_repository":                      dataSourceAwsCodeCommitRepository(),
			"aws_codestarconnections_connection":             dataSourceAwsCodeStarConnectionsConnection(),
			"aws_cur_report_definition":                      dataSourceAwsCurReportDefinition(),
			"aws_default_tags":                               dataSourceAwsDefaultTags(),
			"aws_db_cluster_snapshot":                        dataSourceAwsDbClusterSnapshot(),
			"aws_db_event_categories":                        dataSourceAwsDbEventCategories(),
			"aws_db_instance":                                dataSourceAwsDbInstance(),
			"aws_db_snapshot":                                dataSourceAwsDbSnapshot(),
			"aws_db_subnet_group":                            dataSourceAwsDbSubnetGroup(),
			"aws_directory_service_directory":                dataSourceAwsDirectoryServiceDirectory(),
			"aws_docdb_engine_version":                       dataSourceAwsDocdbEngineVersion(),
			"aws_docdb_orderable_db_instance":                dataSourceAwsDocdbOrderableDbInstance(),
			"aws_dx_gateway":                                 dataSourceAwsDxGateway(),
			"aws_dynamodb_table":                             dataSourceAwsDynamoDbTable(),
			"aws_ebs_default_kms_key":                        dataSourceAwsEbsDefaultKmsKey(),
			"aws_ebs_encryption_by_default":                  dataSourceAwsEbsEncryptionByDefault(),
			"aws_ebs_snapshot":                               dataSourceAwsEbsSnapshot(),
			"aws_ebs_snapshot_ids":                           dataSourceAwsEbsSnapshotIds(),
			"aws_ebs_volume":                                 dataSourceAwsEbsVolume(),
			"aws_ebs_volumes":                                dataSourceAwsEbsVolumes(),
			"aws_ec2_coip_pool":                              dataSourceAwsEc2CoipPool(),
			"aws_ec2_coip_pools":                             dataSourceAwsEc2CoipPools(),
			"aws_ec2_instance_type":                          dataSourceAwsEc2InstanceType(),
			"aws_ec2_instance_type_offering":                 dataSourceAwsEc2InstanceTypeOffering(),
			"aws_ec2_instance_type_offerings":                dataSourceAwsEc2InstanceTypeOfferings(),
			"aws_ec2_local_gateway":                          dataSourceAwsEc2LocalGateway(),
			"aws_ec2_local_gateways":                         dataSourceAwsEc2LocalGateways(),
			"aws_ec2_local_gateway_route_table":              dataSourceAwsEc2LocalGatewayRouteTable(),
			"aws_ec2_local_gateway_route_tables":             dataSourceAwsEc2LocalGatewayRouteTables(),
			"aws_ec2_local_gateway_virtual_interface":        dataSourceAwsEc2LocalGatewayVirtualInterface(),
			"aws_ec2_local_gateway_virtual_interface_group":  dataSourceAwsEc2LocalGatewayVirtualInterfaceGroup(),
			"aws_ec2_local_gateway_virtual_interface_groups": dataSourceAwsEc2LocalGatewayVirtualInterfaceGroups(),
			"aws_ec2_managed_prefix_list":                    dataSourceAwsEc2ManagedPrefixList(),
			"aws_ec2_spot_price":                             dataSourceAwsEc2SpotPrice(),
			"aws_ec2_transit_gateway":                        dataSourceAwsEc2TransitGateway(),
			"aws_ec2_transit_gateway_dx_gateway_attachment":  dataSourceAwsEc2TransitGatewayDxGatewayAttachment(),
			"aws_ec2_transit_gateway_peering_attachment":     dataSourceAwsEc2TransitGatewayPeeringAttachment(),
			"aws_ec2_transit_gateway_route_table":            dataSourceAwsEc2TransitGatewayRouteTable(),
			"aws_ec2_transit_gateway_route_tables":           dataSourceAwsEc2TransitGatewayRouteTables(),
			"aws_ec2_transit_gateway_vpc_attachment":         dataSourceAwsEc2TransitGatewayVpcAttachment(),
			"aws_ec2_transit_gateway_vpn_attachment":         dataSourceAwsEc2TransitGatewayVpnAttachment(),
			"aws_ecr_authorization_token":                    dataSourceAwsEcrAuthorizationToken(),
			"aws_ecr_image":                                  dataSourceAwsEcrImage(),
			"aws_ecr_repository":                             dataSourceAwsEcrRepository(),
			"aws_ecs_cluster":                                dataSourceAwsEcsCluster(),
			"aws_ecs_container_definition":                   dataSourceAwsEcsContainerDefinition(),
			"aws_ecs_service":                                dataSourceAwsEcsService(),
			"aws_ecs_task_definition":                        dataSourceAwsEcsTaskDefinition(),
			"aws_customer_gateway":                           dataSourceAwsCustomerGateway(),
			"aws_efs_access_point":                           dataSourceAwsEfsAccessPoint(),
			"aws_efs_access_points":                          dataSourceAwsEfsAccessPoints(),
			"aws_efs_file_system":                            dataSourceAwsEfsFileSystem(),
			"aws_efs_mount_target":                           dataSourceAwsEfsMountTarget(),
			"aws_eip":                                        dataSourceAwsEip(),
			"aws_eks_addon":                                  dataSourceAwsEksAddon(),
			"aws_eks_cluster":                                dataSourceAwsEksCluster(),
			"aws_eks_cluster_auth":                           dataSourceAwsEksClusterAuth(),
			"aws_elastic_beanstalk_application":              dataSourceAwsElasticBeanstalkApplication(),
			"aws_elastic_beanstalk_hosted_zone":              dataSourceAwsElasticBeanstalkHostedZone(),
			"aws_elastic_beanstalk_solution_stack":           dataSourceAwsElasticBeanstalkSolutionStack(),
			"aws_elasticache_cluster":                        dataSourceAwsElastiCacheCluster(),
			"aws_elasticsearch_domain":                       dataSourceAwsElasticSearchDomain(),
			"aws_elb":                                        dataSourceAwsElb(),
			"aws_elasticache_replication_group":              dataSourceAwsElasticacheReplicationGroup(),
			"aws_elb_hosted_zone_id":                         dataSourceAwsElbHostedZoneId(),
			"aws_elb_service_account":                        dataSourceAwsElbServiceAccount(),
			"aws_globalaccelerator_accelerator":              dataSourceAwsGlobalAcceleratorAccelerator(),
			"aws_glue_connection":                            dataSourceAwsGlueConnection(),
			"aws_glue_data_catalog_encryption_settings":      dataSourceAwsGlueDataCatalogEncryptionSettings(),
			"aws_glue_script":                                dataSourceAwsGlueScript(),
			"aws_guardduty_detector":                         dataSourceAwsGuarddutyDetector(),
			"aws_iam_account_alias":                          dataSourceAwsIamAccountAlias(),
			"aws_iam_group":                                  dataSourceAwsIAMGroup(),
			"aws_iam_instance_profile":                       dataSourceAwsIAMInstanceProfile(),
			"aws_iam_policy":                                 dataSourceAwsIAMPolicy(),
			"aws_iam_policy_document":                        dataSourceAwsIamPolicyDocument(),
			"aws_iam_role":                                   dataSourceAwsIAMRole(),
			"aws_iam_server_certificate":                     dataSourceAwsIAMServerCertificate(),
			"aws_iam_user":                                   dataSourceAwsIAMUser(),
			"aws_identitystore_group":                        dataSourceAwsIdentityStoreGroup(),
			"aws_identitystore_user":                         dataSourceAwsIdentityStoreUser(),
			"aws_imagebuilder_component":                     dataSourceAwsImageBuilderComponent(),
			"aws_imagebuilder_distribution_configuration":    datasourceAwsImageBuilderDistributionConfiguration(),
			"aws_imagebuilder_image":                         dataSourceAwsImageBuilderImage(),
			"aws_imagebuilder_image_pipeline":                dataSourceAwsImageBuilderImagePipeline(),
			"aws_imagebuilder_image_recipe":                  dataSourceAwsImageBuilderImageRecipe(),
			"aws_imagebuilder_infrastructure_configuration":  datasourceAwsImageBuilderInfrastructureConfiguration(),
			"aws_inspector_rules_packages":                   dataSourceAwsInspectorRulesPackages(),
			"aws_instance":                                   dataSourceAwsInstance(),
			"aws_instances":                                  dataSourceAwsInstances(),
			"aws_internet_gateway":                           dataSourceAwsInternetGateway(),
			"aws_iot_endpoint":                               dataSourceAwsIotEndpoint(),
			"aws_ip_ranges":                                  dataSourceAwsIPRanges(),
			"aws_kinesis_stream":                             dataSourceAwsKinesisStream(),
			"aws_kinesis_stream_consumer":                    dataSourceAwsKinesisStreamConsumer(),
			"aws_kms_alias":                                  dataSourceAwsKmsAlias(),
			"aws_kms_ciphertext":                             dataSourceAwsKmsCiphertext(),
			"aws_kms_key":                                    dataSourceAwsKmsKey(),
			"aws_kms_public_key":                             dataSourceAwsKmsPublicKey(),
			"aws_kms_secret":                                 dataSourceAwsKmsSecret(),
			"aws_kms_secrets":                                dataSourceAwsKmsSecrets(),
			"aws_lakeformation_data_lake_settings":           dataSourceAwsLakeFormationDataLakeSettings(),
			"aws_lakeformation_permissions":                  dataSourceAwsLakeFormationPermissions(),
			"aws_lakeformation_resource":                     dataSourceAwsLakeFormationResource(),
			"aws_lambda_alias":                               dataSourceAwsLambdaAlias(),
			"aws_lambda_code_signing_config":                 dataSourceAwsLambdaCodeSigningConfig(),
			"aws_lambda_function":                            dataSourceAwsLambdaFunction(),
			"aws_lambda_invocation":                          dataSourceAwsLambdaInvocation(),
			"aws_lambda_layer_version":                       dataSourceAwsLambdaLayerVersion(),
			"aws_launch_configuration":                       dataSourceAwsLaunchConfiguration(),
			"aws_launch_template":                            dataSourceAwsLaunchTemplate(),
			"aws_lex_bot_alias":                              dataSourceAwsLexBotAlias(),
			"aws_lex_bot":                                    dataSourceAwsLexBot(),
			"aws_lex_intent":                                 dataSourceAwsLexIntent(),
			"aws_lex_slot_type":                              dataSourceAwsLexSlotType(),
			"aws_mq_broker":                                  dataSourceAwsMqBroker(),
			"aws_msk_cluster":                                dataSourceAwsMskCluster(),
			"aws_msk_configuration":                          dataSourceAwsMskConfiguration(),
			"aws_nat_gateway":                                dataSourceAwsNatGateway(),
			"aws_neptune_orderable_db_instance":              dataSourceAwsNeptuneOrderableDbInstance(),
			"aws_neptune_engine_version":                     dataSourceAwsNeptuneEngineVersion(),
			"aws_network_acls":                               dataSourceAwsNetworkAcls(),
			"aws_network_interface":                          dataSourceAwsNetworkInterface(),
			"aws_network_interfaces":                         dataSourceAwsNetworkInterfaces(),
			"aws_organizations_delegated_administrators":     dataSourceAwsOrganizationsDelegatedAdministrators(),
			"aws_organizations_delegated_services":           dataSourceAwsOrganizationsDelegatedServices(),
			"aws_organizations_organization":                 dataSourceAwsOrganizationsOrganization(),
			"aws_organizations_organizational_units":         dataSourceAwsOrganizationsOrganizationalUnits(),
			"aws_outposts_outpost":                           dataSourceAwsOutpostsOutpost(),
			"aws_outposts_outpost_instance_type":             dataSourceAwsOutpostsOutpostInstanceType(),
			"aws_outposts_outpost_instance_types":            dataSourceAwsOutpostsOutpostInstanceTypes(),
			"aws_outposts_outposts":                          dataSourceAwsOutpostsOutposts(),
			"aws_outposts_site":                              dataSourceAwsOutpostsSite(),
			"aws_outposts_sites":                             dataSourceAwsOutpostsSites(),
			"aws_partition":                                  dataSourceAwsPartition(),
			"aws_prefix_list":                                dataSourceAwsPrefixList(),
			"aws_pricing_product":                            dataSourceAwsPricingProduct(),
			"aws_qldb_ledger":                                dataSourceAwsQLDBLedger(),
			"aws_ram_resource_share":                         dataSourceAwsRamResourceShare(),
			"aws_rds_certificate":                            dataSourceAwsRdsCertificate(),
			"aws_rds_cluster":                                dataSourceAwsRdsCluster(),
			"aws_rds_engine_version":                         dataSourceAwsRdsEngineVersion(),
			"aws_rds_orderable_db_instance":                  dataSourceAwsRdsOrderableDbInstance(),
			"aws_redshift_cluster":                           dataSourceAwsRedshiftCluster(),
			"aws_redshift_orderable_cluster":                 dataSourceAwsRedshiftOrderableCluster(),
			"aws_redshift_service_account":                   dataSourceAwsRedshiftServiceAccount(),
			"aws_region":                                     dataSourceAwsRegion(),
			"aws_regions":                                    dataSourceAwsRegions(),
			"aws_resourcegroupstaggingapi_resources":         dataSourceAwsResourceGroupsTaggingAPIResources(),
			"aws_route":                                      dataSourceAwsRoute(),
			"aws_route_table":                                dataSourceAwsRouteTable(),
			"aws_route_tables":                               dataSourceAwsRouteTables(),
			"aws_route53_delegation_set":                     dataSourceAwsDelegationSet(),
			"aws_route53_resolver_endpoint":                  dataSourceAwsRoute53ResolverEndpoint(),
			"aws_route53_resolver_rule":                      dataSourceAwsRoute53ResolverRule(),
			"aws_route53_resolver_rules":                     dataSourceAwsRoute53ResolverRules(),
			"aws_route53_zone":                               dataSourceAwsRoute53Zone(),
			"aws_s3_bucket":                                  dataSourceAwsS3Bucket(),
			"aws_s3_bucket_object":                           dataSourceAwsS3BucketObject(),
			"aws_s3_bucket_objects":                          dataSourceAwsS3BucketObjects(),
			"aws_sagemaker_prebuilt_ecr_image":               dataSourceAwsSageMakerPrebuiltECRImage(),
			"aws_secretsmanager_secret":                      dataSourceAwsSecretsManagerSecret(),
			"aws_secretsmanager_secret_rotation":             dataSourceAwsSecretsManagerSecretRotation(),
			"aws_secretsmanager_secret_version":              dataSourceAwsSecretsManagerSecretVersion(),
			"aws_servicecatalog_constraint":                  dataSourceAwsServiceCatalogConstraint(),
			"aws_servicequotas_service":                      dataSourceAwsServiceQuotasService(),
			"aws_servicequotas_service_quota":                dataSourceAwsServiceQuotasServiceQuota(),
			"aws_service_discovery_dns_namespace":            dataSourceServiceDiscoveryDnsNamespace(),
			"aws_sfn_activity":                               dataSourceAwsSfnActivity(),
			"aws_sfn_state_machine":                          dataSourceAwsSfnStateMachine(),
			"aws_signer_signing_job":                         dataSourceAwsSignerSigningJob(),
			"aws_signer_signing_profile":                     dataSourceAwsSignerSigningProfile(),
			"aws_sns_topic":                                  dataSourceAwsSnsTopic(),
			"aws_sqs_queue":                                  dataSourceAwsSqsQueue(),
			"aws_ssm_document":                               dataSourceAwsSsmDocument(),
			"aws_ssm_parameter":                              dataSourceAwsSsmParameter(),
			"aws_ssm_patch_baseline":                         dataSourceAwsSsmPatchBaseline(),
			"aws_ssoadmin_instances":                         dataSourceAwsSsoAdminInstances(),
			"aws_ssoadmin_permission_set":                    dataSourceAwsSsoAdminPermissionSet(),
			"aws_storagegateway_local_disk":                  dataSourceAwsStorageGatewayLocalDisk(),
			"aws_subnet":                                     dataSourceAwsSubnet(),
			"aws_subnet_ids":                                 dataSourceAwsSubnetIDs(),
			"aws_transfer_server":                            dataSourceAwsTransferServer(),
			"aws_vpcs":                                       dataSourceAwsVpcs(),
			"aws_security_group":                             dataSourceAwsSecurityGroup(),
			"aws_security_groups":                            dataSourceAwsSecurityGroups(),
			"aws_vpc":                                        dataSourceAwsVpc(),
			"aws_vpc_dhcp_options":                           dataSourceAwsVpcDhcpOptions(),
			"aws_vpc_endpoint":                               dataSourceAwsVpcEndpoint(),
			"aws_vpc_endpoint_service":                       dataSourceAwsVpcEndpointService(),
			"aws_vpc_peering_connection":                     dataSourceAwsVpcPeeringConnection(),
			"aws_vpc_peering_connections":                    dataSourceAwsVpcPeeringConnections(),
			"aws_vpn_gateway":                                dataSourceAwsVpnGateway(),
			"aws_waf_ipset":                                  dataSourceAwsWafIpSet(),
			"aws_waf_rule":                                   dataSourceAwsWafRule(),
			"aws_waf_rate_based_rule":                        dataSourceAwsWafRateBasedRule(),
			"aws_waf_web_acl":                                dataSourceAwsWafWebAcl(),
			"aws_wafregional_ipset":                          dataSourceAwsWafRegionalIpSet(),
			"aws_wafregional_rule":                           dataSourceAwsWafRegionalRule(),
			"aws_wafregional_rate_based_rule":                dataSourceAwsWafRegionalRateBasedRule(),
			"aws_wafregional_web_acl":                        dataSourceAwsWafRegionalWebAcl(),
			"aws_wafv2_ip_set":                               dataSourceAwsWafv2IPSet(),
			"aws_wafv2_regex_pattern_set":                    dataSourceAwsWafv2RegexPatternSet(),
			"aws_wafv2_rule_group":                           dataSourceAwsWafv2RuleGroup(),
			"aws_wafv2_web_acl":                              dataSourceAwsWafv2WebACL(),
			"aws_workspaces_bundle":                          dataSourceAwsWorkspacesBundle(),
			"aws_workspaces_directory":                       dataSourceAwsWorkspacesDirectory(),
			"aws_workspaces_image":                           dataSourceAwsWorkspacesImage(),
			"aws_workspaces_workspace":                       dataSourceAwsWorkspacesWorkspace(),

			// Adding the Aliases for the ALB -> LB Rename
			"aws_lb":               dataSourceAwsLb(),
			"aws_alb":              dataSourceAwsLb(),
			"aws_lb_listener":      dataSourceAwsLbListener(),
			"aws_alb_listener":     dataSourceAwsLbListener(),
			"aws_lb_target_group":  dataSourceAwsLbTargetGroup(),
			"aws_alb_target_group": dataSourceAwsLbTargetGroup(),
		},

		ResourcesMap: map[string]*schema.Resource{
			"aws_accessanalyzer_analyzer":                             resourceAwsAccessAnalyzerAnalyzer(),
			"aws_acm_certificate":                                     resourceAwsAcmCertificate(),
			"aws_acm_certificate_validation":                          resourceAwsAcmCertificateValidation(),
			"aws_acmpca_certificate_authority":                        resourceAwsAcmpcaCertificateAuthority(),
			"aws_acmpca_certificate_authority_certificate":            resourceAwsAcmpcaCertificateAuthorityCertificate(),
			"aws_acmpca_certificate":                                  resourceAwsAcmpcaCertificate(),
			"aws_ami":                                                 resourceAwsAmi(),
			"aws_ami_copy":                                            resourceAwsAmiCopy(),
			"aws_ami_from_instance":                                   resourceAwsAmiFromInstance(),
			"aws_ami_launch_permission":                               resourceAwsAmiLaunchPermission(),
			"aws_api_gateway_account":                                 resourceAwsApiGatewayAccount(),
			"aws_api_gateway_api_key":                                 resourceAwsApiGatewayApiKey(),
			"aws_api_gateway_authorizer":                              resourceAwsApiGatewayAuthorizer(),
			"aws_api_gateway_base_path_mapping":                       resourceAwsApiGatewayBasePathMapping(),
			"aws_api_gateway_client_certificate":                      resourceAwsApiGatewayClientCertificate(),
			"aws_api_gateway_deployment":                              resourceAwsApiGatewayDeployment(),
			"aws_api_gateway_documentation_part":                      resourceAwsApiGatewayDocumentationPart(),
			"aws_api_gateway_documentation_version":                   resourceAwsApiGatewayDocumentationVersion(),
			"aws_api_gateway_domain_name":                             resourceAwsApiGatewayDomainName(),
			"aws_api_gateway_gateway_response":                        resourceAwsApiGatewayGatewayResponse(),
			"aws_api_gateway_integration":                             resourceAwsApiGatewayIntegration(),
			"aws_api_gateway_integration_response":                    resourceAwsApiGatewayIntegrationResponse(),
			"aws_api_gateway_method":                                  resourceAwsApiGatewayMethod(),
			"aws_api_gateway_method_response":                         resourceAwsApiGatewayMethodResponse(),
			"aws_api_gateway_method_settings":                         resourceAwsApiGatewayMethodSettings(),
			"aws_api_gateway_model":                                   resourceAwsApiGatewayModel(),
			"aws_api_gateway_request_validator":                       resourceAwsApiGatewayRequestValidator(),
			"aws_api_gateway_resource":                                resourceAwsApiGatewayResource(),
			"aws_api_gateway_rest_api":                                resourceAwsApiGatewayRestApi(),
			"aws_api_gateway_rest_api_policy":                         resourceAwsApiGatewayRestApiPolicy(),
			"aws_api_gateway_stage":                                   resourceAwsApiGatewayStage(),
			"aws_api_gateway_usage_plan":                              resourceAwsApiGatewayUsagePlan(),
			"aws_api_gateway_usage_plan_key":                          resourceAwsApiGatewayUsagePlanKey(),
			"aws_api_gateway_vpc_link":                                resourceAwsApiGatewayVpcLink(),
			"aws_apigatewayv2_api":                                    resourceAwsApiGatewayV2Api(),
			"aws_apigatewayv2_api_mapping":                            resourceAwsApiGatewayV2ApiMapping(),
			"aws_apigatewayv2_authorizer":                             resourceAwsApiGatewayV2Authorizer(),
			"aws_apigatewayv2_deployment":                             resourceAwsApiGatewayV2Deployment(),
			"aws_apigatewayv2_domain_name":                            resourceAwsApiGatewayV2DomainName(),
			"aws_apigatewayv2_integration":                            resourceAwsApiGatewayV2Integration(),
			"aws_apigatewayv2_integration_response":                   resourceAwsApiGatewayV2IntegrationResponse(),
			"aws_apigatewayv2_model":                                  resourceAwsApiGatewayV2Model(),
			"aws_apigatewayv2_route":                                  resourceAwsApiGatewayV2Route(),
			"aws_apigatewayv2_route_response":                         resourceAwsApiGatewayV2RouteResponse(),
			"aws_apigatewayv2_stage":                                  resourceAwsApiGatewayV2Stage(),
			"aws_apigatewayv2_vpc_link":                               resourceAwsApiGatewayV2VpcLink(),
			"aws_app_cookie_stickiness_policy":                        resourceAwsAppCookieStickinessPolicy(),
			"aws_appautoscaling_target":                               resourceAwsAppautoscalingTarget(),
			"aws_appautoscaling_policy":                               resourceAwsAppautoscalingPolicy(),
			"aws_appautoscaling_scheduled_action":                     resourceAwsAppautoscalingScheduledAction(),
			"aws_appmesh_gateway_route":                               resourceAwsAppmeshGatewayRoute(),
			"aws_appmesh_mesh":                                        resourceAwsAppmeshMesh(),
			"aws_appmesh_route":                                       resourceAwsAppmeshRoute(),
			"aws_appmesh_virtual_gateway":                             resourceAwsAppmeshVirtualGateway(),
			"aws_appmesh_virtual_node":                                resourceAwsAppmeshVirtualNode(),
			"aws_appmesh_virtual_router":                              resourceAwsAppmeshVirtualRouter(),
			"aws_appmesh_virtual_service":                             resourceAwsAppmeshVirtualService(),
			"aws_apprunner_auto_scaling_configuration_version":        resourceAwsAppRunnerAutoScalingConfigurationVersion(),
			"aws_apprunner_connection":                                resourceAwsAppRunnerConnection(),
			"aws_apprunner_custom_domain_association":                 resourceAwsAppRunnerCustomDomainAssociation(),
			"aws_apprunner_service":                                   resourceAwsAppRunnerService(),
			"aws_appsync_api_key":                                     resourceAwsAppsyncApiKey(),
			"aws_appsync_datasource":                                  resourceAwsAppsyncDatasource(),
			"aws_appsync_function":                                    resourceAwsAppsyncFunction(),
			"aws_appsync_graphql_api":                                 resourceAwsAppsyncGraphqlApi(),
			"aws_appsync_resolver":                                    resourceAwsAppsyncResolver(),
			"aws_athena_database":                                     resourceAwsAthenaDatabase(),
			"aws_athena_named_query":                                  resourceAwsAthenaNamedQuery(),
			"aws_athena_workgroup":                                    resourceAwsAthenaWorkgroup(),
			"aws_autoscaling_attachment":                              resourceAwsAutoscalingAttachment(),
			"aws_autoscaling_group":                                   resourceAwsAutoscalingGroup(),
			"aws_autoscaling_lifecycle_hook":                          resourceAwsAutoscalingLifecycleHook(),
			"aws_autoscaling_notification":                            resourceAwsAutoscalingNotification(),
			"aws_autoscaling_policy":                                  resourceAwsAutoscalingPolicy(),
			"aws_autoscaling_schedule":                                resourceAwsAutoscalingSchedule(),
			"aws_autoscalingplans_scaling_plan":                       resourceAwsAutoScalingPlansScalingPlan(),
			"aws_backup_global_settings":                              resourceAwsBackupGlobalSettings(),
			"aws_backup_plan":                                         resourceAwsBackupPlan(),
			"aws_backup_region_settings":                              resourceAwsBackupRegionSettings(),
			"aws_backup_selection":                                    resourceAwsBackupSelection(),
			"aws_backup_vault":                                        resourceAwsBackupVault(),
			"aws_backup_vault_notifications":                          resourceAwsBackupVaultNotifications(),
			"aws_backup_vault_policy":                                 resourceAwsBackupVaultPolicy(),
			"aws_budgets_budget":                                      resourceAwsBudgetsBudget(),
			"aws_budgets_budget_action":                               resourceAwsBudgetsBudgetAction(),
			"aws_cloud9_environment_ec2":                              resourceAwsCloud9EnvironmentEc2(),
			"aws_cloudformation_stack":                                resourceAwsCloudFormationStack(),
			"aws_cloudformation_stack_set":                            resourceAwsCloudFormationStackSet(),
			"aws_cloudformation_stack_set_instance":                   resourceAwsCloudFormationStackSetInstance(),
			"aws_cloudformation_type":                                 resourceAwsCloudFormationType(),
			"aws_cloudfront_cache_policy":                             resourceAwsCloudFrontCachePolicy(),
			"aws_cloudfront_distribution":                             resourceAwsCloudFrontDistribution(),
			"aws_cloudfront_function":                                 resourceAwsCloudFrontFunction(),
			"aws_cloudfront_key_group":                                resourceAwsCloudFrontKeyGroup(),
			"aws_cloudfront_origin_access_identity":                   resourceAwsCloudFrontOriginAccessIdentity(),
			"aws_cloudfront_origin_request_policy":                    resourceAwsCloudFrontOriginRequestPolicy(),
			"aws_cloudfront_public_key":                               resourceAwsCloudFrontPublicKey(),
			"aws_cloudfront_realtime_log_config":                      resourceAwsCloudFrontRealtimeLogConfig(),
			"aws_cloudtrail":                                          resourceAwsCloudTrail(),
			"aws_cloudwatch_event_bus":                                resourceAwsCloudWatchEventBus(),
			"aws_cloudwatch_event_permission":                         resourceAwsCloudWatchEventPermission(),
			"aws_cloudwatch_event_rule":                               resourceAwsCloudWatchEventRule(),
			"aws_cloudwatch_event_target":                             resourceAwsCloudWatchEventTarget(),
			"aws_cloudwatch_event_archive":                            resourceAwsCloudWatchEventArchive(),
			"aws_cloudwatch_event_connection":                         resourceAwsCloudWatchEventConnection(),
			"aws_cloudwatch_event_api_destination":                    resourceAwsCloudWatchEventApiDestination(),
			"aws_cloudwatch_log_destination":                          resourceAwsCloudWatchLogDestination(),
			"aws_cloudwatch_log_destination_policy":                   resourceAwsCloudWatchLogDestinationPolicy(),
			"aws_cloudwatch_log_group":                                resourceAwsCloudWatchLogGroup(),
			"aws_cloudwatch_log_metric_filter":                        resourceAwsCloudWatchLogMetricFilter(),
			"aws_cloudwatch_log_resource_policy":                      resourceAwsCloudWatchLogResourcePolicy(),
			"aws_cloudwatch_log_stream":                               resourceAwsCloudWatchLogStream(),
			"aws_cloudwatch_log_subscription_filter":                  resourceAwsCloudwatchLogSubscriptionFilter(),
			"aws_config_aggregate_authorization":                      resourceAwsConfigAggregateAuthorization(),
			"aws_config_config_rule":                                  resourceAwsConfigConfigRule(),
			"aws_config_configuration_aggregator":                     resourceAwsConfigConfigurationAggregator(),
			"aws_config_configuration_recorder":                       resourceAwsConfigConfigurationRecorder(),
			"aws_config_configuration_recorder_status":                resourceAwsConfigConfigurationRecorderStatus(),
			"aws_config_conformance_pack":                             resourceAwsConfigConformancePack(),
			"aws_config_delivery_channel":                             resourceAwsConfigDeliveryChannel(),
			"aws_config_organization_custom_rule":                     resourceAwsConfigOrganizationCustomRule(),
			"aws_config_organization_managed_rule":                    resourceAwsConfigOrganizationManagedRule(),
			"aws_config_remediation_configuration":                    resourceAwsConfigRemediationConfiguration(),
			"aws_cognito_identity_pool":                               resourceAwsCognitoIdentityPool(),
			"aws_cognito_identity_pool_roles_attachment":              resourceAwsCognitoIdentityPoolRolesAttachment(),
			"aws_cognito_identity_provider":                           resourceAwsCognitoIdentityProvider(),
			"aws_cognito_resource_server":                             resourceAwsCognitoResourceServer(),
			"aws_cognito_user_group":                                  resourceAwsCognitoUserGroup(),
			"aws_cognito_user_pool":                                   resourceAwsCognitoUserPool(),
			"aws_cognito_user_pool_client":                            resourceAwsCognitoUserPoolClient(),
			"aws_cognito_user_pool_domain":                            resourceAwsCognitoUserPoolDomain(),
			"aws_cognito_user_pool_ui_customization":                  resourceAwsCognitoUserPoolUICustomization(),
			"aws_cloudhsm_v2_cluster":                                 resourceAwsCloudHsmV2Cluster(),
			"aws_cloudhsm_v2_hsm":                                     resourceAwsCloudHsmV2Hsm(),
			"aws_cloudwatch_composite_alarm":                          resourceAwsCloudWatchCompositeAlarm(),
			"aws_cloudwatch_metric_alarm":                             resourceAwsCloudWatchMetricAlarm(),
			"aws_cloudwatch_dashboard":                                resourceAwsCloudWatchDashboard(),
			"aws_cloudwatch_metric_stream":                            resourceAwsCloudWatchMetricStream(),
			"aws_cloudwatch_query_definition":                         resourceAwsCloudWatchQueryDefinition(),
			"aws_codedeploy_app":                                      resourceAwsCodeDeployApp(),
			"aws_codedeploy_deployment_config":                        resourceAwsCodeDeployDeploymentConfig(),
			"aws_codedeploy_deployment_group":                         resourceAwsCodeDeployDeploymentGroup(),
			"aws_codecommit_repository":                               resourceAwsCodeCommitRepository(),
			"aws_codecommit_trigger":                                  resourceAwsCodeCommitTrigger(),
			"aws_codeartifact_domain":                                 resourceAwsCodeArtifactDomain(),
			"aws_codeartifact_domain_permissions_policy":              resourceAwsCodeArtifactDomainPermissionsPolicy(),
			"aws_codeartifact_repository":                             resourceAwsCodeArtifactRepository(),
			"aws_codeartifact_repository_permissions_policy":          resourceAwsCodeArtifactRepositoryPermissionsPolicy(),
			"aws_codebuild_project":                                   resourceAwsCodeBuildProject(),
			"aws_codebuild_report_group":                              resourceAwsCodeBuildReportGroup(),
			"aws_codebuild_source_credential":                         resourceAwsCodeBuildSourceCredential(),
			"aws_codebuild_webhook":                                   resourceAwsCodeBuildWebhook(),
			"aws_codepipeline":                                        resourceAwsCodePipeline(),
			"aws_codepipeline_webhook":                                resourceAwsCodePipelineWebhook(),
			"aws_codestarconnections_connection":                      resourceAwsCodeStarConnectionsConnection(),
			"aws_codestarconnections_host":                            resourceAwsCodeStarConnectionsHost(),
			"aws_codestarnotifications_notification_rule":             resourceAwsCodeStarNotificationsNotificationRule(),
			"aws_cur_report_definition":                               resourceAwsCurReportDefinition(),
			"aws_customer_gateway":                                    resourceAwsCustomerGateway(),
			"aws_datapipeline_pipeline":                               resourceAwsDataPipelinePipeline(),
			"aws_datasync_agent":                                      resourceAwsDataSyncAgent(),
			"aws_datasync_location_efs":                               resourceAwsDataSyncLocationEfs(),
			"aws_datasync_location_fsx_windows_file_system":           resourceAwsDataSyncLocationFsxWindowsFileSystem(),
			"aws_datasync_location_nfs":                               resourceAwsDataSyncLocationNfs(),
			"aws_datasync_location_s3":                                resourceAwsDataSyncLocationS3(),
			"aws_datasync_location_smb":                               resourceAwsDataSyncLocationSmb(),
			"aws_datasync_task":                                       resourceAwsDataSyncTask(),
			"aws_dax_cluster":                                         resourceAwsDaxCluster(),
			"aws_dax_parameter_group":                                 resourceAwsDaxParameterGroup(),
			"aws_dax_subnet_group":                                    resourceAwsDaxSubnetGroup(),
			"aws_db_cluster_snapshot":                                 resourceAwsDbClusterSnapshot(),
			"aws_db_event_subscription":                               resourceAwsDbEventSubscription(),
			"aws_db_instance":                                         resourceAwsDbInstance(),
			"aws_db_instance_role_association":                        resourceAwsDbInstanceRoleAssociation(),
			"aws_db_option_group":                                     resourceAwsDbOptionGroup(),
			"aws_db_parameter_group":                                  resourceAwsDbParameterGroup(),
			"aws_db_proxy":                                            resourceAwsDbProxy(),
			"aws_db_proxy_default_target_group":                       resourceAwsDbProxyDefaultTargetGroup(),
			"aws_db_proxy_endpoint":                                   resourceAwsDbProxyEndpoint(),
			"aws_db_proxy_target":                                     resourceAwsDbProxyTarget(),
			"aws_db_security_group":                                   resourceAwsDbSecurityGroup(),
			"aws_db_snapshot":                                         resourceAwsDbSnapshot(),
			"aws_db_subnet_group":                                     resourceAwsDbSubnetGroup(),
			"aws_devicefarm_project":                                  resourceAwsDevicefarmProject(),
			"aws_directory_service_directory":                         resourceAwsDirectoryServiceDirectory(),
			"aws_directory_service_conditional_forwarder":             resourceAwsDirectoryServiceConditionalForwarder(),
			"aws_directory_service_log_subscription":                  resourceAwsDirectoryServiceLogSubscription(),
			"aws_dlm_lifecycle_policy":                                resourceAwsDlmLifecyclePolicy(),
			"aws_dms_certificate":                                     resourceAwsDmsCertificate(),
			"aws_dms_endpoint":                                        resourceAwsDmsEndpoint(),
			"aws_dms_event_subscription":                              resourceAwsDmsEventSubscription(),
			"aws_dms_replication_instance":                            resourceAwsDmsReplicationInstance(),
			"aws_dms_replication_subnet_group":                        resourceAwsDmsReplicationSubnetGroup(),
			"aws_dms_replication_task":                                resourceAwsDmsReplicationTask(),
			"aws_docdb_cluster":                                       resourceAwsDocDBCluster(),
			"aws_docdb_cluster_instance":                              resourceAwsDocDBClusterInstance(),
			"aws_docdb_cluster_parameter_group":                       resourceAwsDocDBClusterParameterGroup(),
			"aws_docdb_cluster_snapshot":                              resourceAwsDocDBClusterSnapshot(),
			"aws_docdb_subnet_group":                                  resourceAwsDocDBSubnetGroup(),
			"aws_dx_bgp_peer":                                         resourceAwsDxBgpPeer(),
			"aws_dx_connection":                                       resourceAwsDxConnection(),
			"aws_dx_connection_association":                           resourceAwsDxConnectionAssociation(),
			"aws_dx_gateway":                                          resourceAwsDxGateway(),
			"aws_dx_gateway_association":                              resourceAwsDxGatewayAssociation(),
			"aws_dx_gateway_association_proposal":                     resourceAwsDxGatewayAssociationProposal(),
			"aws_dx_hosted_private_virtual_interface":                 resourceAwsDxHostedPrivateVirtualInterface(),
			"aws_dx_hosted_private_virtual_interface_accepter":        resourceAwsDxHostedPrivateVirtualInterfaceAccepter(),
			"aws_dx_hosted_public_virtual_interface":                  resourceAwsDxHostedPublicVirtualInterface(),
			"aws_dx_hosted_public_virtual_interface_accepter":         resourceAwsDxHostedPublicVirtualInterfaceAccepter(),
			"aws_dx_hosted_transit_virtual_interface":                 resourceAwsDxHostedTransitVirtualInterface(),
			"aws_dx_hosted_transit_virtual_interface_accepter":        resourceAwsDxHostedTransitVirtualInterfaceAccepter(),
			"aws_dx_lag":                                              resourceAwsDxLag(),
			"aws_dx_private_virtual_interface":                        resourceAwsDxPrivateVirtualInterface(),
			"aws_dx_public_virtual_interface":                         resourceAwsDxPublicVirtualInterface(),
			"aws_dx_transit_virtual_interface":                        resourceAwsDxTransitVirtualInterface(),
			"aws_dynamodb_table":                                      resourceAwsDynamoDbTable(),
			"aws_dynamodb_table_item":                                 resourceAwsDynamoDbTableItem(),
			"aws_dynamodb_global_table":                               resourceAwsDynamoDbGlobalTable(),
			"aws_dynamodb_kinesis_streaming_destination":              resourceAwsDynamoDbKinesisStreamingDestination(),
			"aws_ebs_default_kms_key":                                 resourceAwsEbsDefaultKmsKey(),
			"aws_ebs_encryption_by_default":                           resourceAwsEbsEncryptionByDefault(),
			"aws_ebs_snapshot":                                        resourceAwsEbsSnapshot(),
			"aws_ebs_snapshot_copy":                                   resourceAwsEbsSnapshotCopy(),
			"aws_ebs_volume":                                          resourceAwsEbsVolume(),
			"aws_ec2_availability_zone_group":                         resourceAwsEc2AvailabilityZoneGroup(),
			"aws_ec2_capacity_reservation":                            resourceAwsEc2CapacityReservation(),
			"aws_ec2_carrier_gateway":                                 resourceAwsEc2CarrierGateway(),
			"aws_ec2_client_vpn_authorization_rule":                   resourceAwsEc2ClientVpnAuthorizationRule(),
			"aws_ec2_client_vpn_endpoint":                             resourceAwsEc2ClientVpnEndpoint(),
			"aws_ec2_client_vpn_network_association":                  resourceAwsEc2ClientVpnNetworkAssociation(),
			"aws_ec2_client_vpn_route":                                resourceAwsEc2ClientVpnRoute(),
			"aws_ec2_fleet":                                           resourceAwsEc2Fleet(),
			"aws_ec2_local_gateway_route":                             resourceAwsEc2LocalGatewayRoute(),
			"aws_ec2_local_gateway_route_table_vpc_association":       resourceAwsEc2LocalGatewayRouteTableVpcAssociation(),
			"aws_ec2_managed_prefix_list":                             resourceAwsEc2ManagedPrefixList(),
			"aws_ec2_tag":                                             resourceAwsEc2Tag(),
			"aws_ec2_traffic_mirror_filter":                           resourceAwsEc2TrafficMirrorFilter(),
			"aws_ec2_traffic_mirror_filter_rule":                      resourceAwsEc2TrafficMirrorFilterRule(),
			"aws_ec2_traffic_mirror_target":                           resourceAwsEc2TrafficMirrorTarget(),
			"aws_ec2_traffic_mirror_session":                          resourceAwsEc2TrafficMirrorSession(),
			"aws_ec2_transit_gateway":                                 resourceAwsEc2TransitGateway(),
			"aws_ec2_transit_gateway_peering_attachment":              resourceAwsEc2TransitGatewayPeeringAttachment(),
			"aws_ec2_transit_gateway_peering_attachment_accepter":     resourceAwsEc2TransitGatewayPeeringAttachmentAccepter(),
			"aws_ec2_transit_gateway_prefix_list_reference":           resourceAwsEc2TransitGatewayPrefixListReference(),
			"aws_ec2_transit_gateway_route":                           resourceAwsEc2TransitGatewayRoute(),
			"aws_ec2_transit_gateway_route_table":                     resourceAwsEc2TransitGatewayRouteTable(),
			"aws_ec2_transit_gateway_route_table_association":         resourceAwsEc2TransitGatewayRouteTableAssociation(),
			"aws_ec2_transit_gateway_route_table_propagation":         resourceAwsEc2TransitGatewayRouteTablePropagation(),
			"aws_ec2_transit_gateway_vpc_attachment":                  resourceAwsEc2TransitGatewayVpcAttachment(),
			"aws_ec2_transit_gateway_vpc_attachment_accepter":         resourceAwsEc2TransitGatewayVpcAttachmentAccepter(),
			"aws_ecr_lifecycle_policy":                                resourceAwsEcrLifecyclePolicy(),
			"aws_ecrpublic_repository":                                resourceAwsEcrPublicRepository(),
			"aws_ecr_registry_policy":                                 resourceAwsEcrRegistryPolicy(),
			"aws_ecr_replication_configuration":                       resourceAwsEcrReplicationConfiguration(),
			"aws_ecr_repository":                                      resourceAwsEcrRepository(),
			"aws_ecr_repository_policy":                               resourceAwsEcrRepositoryPolicy(),
			"aws_ecs_capacity_provider":                               resourceAwsEcsCapacityProvider(),
			"aws_ecs_cluster":                                         resourceAwsEcsCluster(),
			"aws_ecs_service":                                         resourceAwsEcsService(),
			"aws_ecs_task_definition":                                 resourceAwsEcsTaskDefinition(),
			"aws_efs_access_point":                                    resourceAwsEfsAccessPoint(),
			"aws_efs_file_system":                                     resourceAwsEfsFileSystem(),
			"aws_efs_file_system_policy":                              resourceAwsEfsFileSystemPolicy(),
			"aws_efs_mount_target":                                    resourceAwsEfsMountTarget(),
			"aws_egress_only_internet_gateway":                        resourceAwsEgressOnlyInternetGateway(),
			"aws_eip":                                                 resourceAwsEip(),
			"aws_eip_association":                                     resourceAwsEipAssociation(),
			"aws_eks_cluster":                                         resourceAwsEksCluster(),
			"aws_eks_addon":                                           resourceAwsEksAddon(),
			"aws_eks_fargate_profile":                                 resourceAwsEksFargateProfile(),
			"aws_eks_node_group":                                      resourceAwsEksNodeGroup(),
			"aws_elasticache_cluster":                                 resourceAwsElasticacheCluster(),
			"aws_elasticache_global_replication_group":                resourceAwsElasticacheGlobalReplicationGroup(),
			"aws_elasticache_parameter_group":                         resourceAwsElasticacheParameterGroup(),
			"aws_elasticache_replication_group":                       resourceAwsElasticacheReplicationGroup(),
			"aws_elasticache_security_group":                          resourceAwsElasticacheSecurityGroup(),
			"aws_elasticache_subnet_group":                            resourceAwsElasticacheSubnetGroup(),
			"aws_elastic_beanstalk_application":                       resourceAwsElasticBeanstalkApplication(),
			"aws_elastic_beanstalk_application_version":               resourceAwsElasticBeanstalkApplicationVersion(),
			"aws_elastic_beanstalk_configuration_template":            resourceAwsElasticBeanstalkConfigurationTemplate(),
			"aws_elastic_beanstalk_environment":                       resourceAwsElasticBeanstalkEnvironment(),
			"aws_elasticsearch_domain":                                resourceAwsElasticSearchDomain(),
			"aws_elasticsearch_domain_policy":                         resourceAwsElasticSearchDomainPolicy(),
			"aws_elastictranscoder_pipeline":                          resourceAwsElasticTranscoderPipeline(),
			"aws_elastictranscoder_preset":                            resourceAwsElasticTranscoderPreset(),
			"aws_elb":                                                 resourceAwsElb(),
			"aws_elb_attachment":                                      resourceAwsElbAttachment(),
			"aws_emr_cluster":                                         resourceAwsEMRCluster(),
			"aws_emr_instance_group":                                  resourceAwsEMRInstanceGroup(),
			"aws_emr_instance_fleet":                                  resourceAwsEMRInstanceFleet(),
			"aws_emr_managed_scaling_policy":                          resourceAwsEMRManagedScalingPolicy(),
			"aws_emr_security_configuration":                          resourceAwsEMRSecurityConfiguration(),
			"aws_flow_log":                                            resourceAwsFlowLog(),
			"aws_fsx_lustre_file_system":                              resourceAwsFsxLustreFileSystem(),
			"aws_fsx_windows_file_system":                             resourceAwsFsxWindowsFileSystem(),
			"aws_fms_admin_account":                                   resourceAwsFmsAdminAccount(),
			"aws_fms_policy":                                          resourceAwsFmsPolicy(),
			"aws_gamelift_alias":                                      resourceAwsGameliftAlias(),
			"aws_gamelift_build":                                      resourceAwsGameliftBuild(),
			"aws_gamelift_fleet":                                      resourceAwsGameliftFleet(),
			"aws_gamelift_game_session_queue":                         resourceAwsGameliftGameSessionQueue(),
			"aws_glacier_vault":                                       resourceAwsGlacierVault(),
			"aws_glacier_vault_lock":                                  resourceAwsGlacierVaultLock(),
			"aws_globalaccelerator_accelerator":                       resourceAwsGlobalAcceleratorAccelerator(),
			"aws_globalaccelerator_endpoint_group":                    resourceAwsGlobalAcceleratorEndpointGroup(),
			"aws_globalaccelerator_listener":                          resourceAwsGlobalAcceleratorListener(),
			"aws_glue_catalog_database":                               resourceAwsGlueCatalogDatabase(),
			"aws_glue_catalog_table":                                  resourceAwsGlueCatalogTable(),
			"aws_glue_classifier":                                     resourceAwsGlueClassifier(),
			"aws_glue_connection":                                     resourceAwsGlueConnection(),
			"aws_glue_dev_endpoint":                                   resourceAwsGlueDevEndpoint(),
			"aws_glue_crawler":                                        resourceAwsGlueCrawler(),
			"aws_glue_data_catalog_encryption_settings":               resourceAwsGlueDataCatalogEncryptionSettings(),
			"aws_glue_job":                                            resourceAwsGlueJob(),
			"aws_glue_ml_transform":                                   resourceAwsGlueMLTransform(),
			"aws_glue_partition":                                      resourceAwsGluePartition(),
			"aws_glue_registry":                                       resourceAwsGlueRegistry(),
			"aws_glue_resource_policy":                                resourceAwsGlueResourcePolicy(),
			"aws_glue_schema":                                         resourceAwsGlueSchema(),
			"aws_glue_security_configuration":                         resourceAwsGlueSecurityConfiguration(),
			"aws_glue_trigger":                                        resourceAwsGlueTrigger(),
			"aws_glue_user_defined_function":                          resourceAwsGlueUserDefinedFunction(),
			"aws_glue_workflow":                                       resourceAwsGlueWorkflow(),
			"aws_guardduty_detector":                                  resourceAwsGuardDutyDetector(),
			"aws_guardduty_filter":                                    resourceAwsGuardDutyFilter(),
			"aws_guardduty_invite_accepter":                           resourceAwsGuardDutyInviteAccepter(),
			"aws_guardduty_ipset":                                     resourceAwsGuardDutyIpset(),
			"aws_guardduty_member":                                    resourceAwsGuardDutyMember(),
			"aws_guardduty_organization_admin_account":                resourceAwsGuardDutyOrganizationAdminAccount(),
			"aws_guardduty_organization_configuration":                resourceAwsGuardDutyOrganizationConfiguration(),
			"aws_guardduty_publishing_destination":                    resourceAwsGuardDutyPublishingDestination(),
			"aws_guardduty_threatintelset":                            resourceAwsGuardDutyThreatintelset(),
			"aws_iam_access_key":                                      resourceAwsIamAccessKey(),
			"aws_iam_account_alias":                                   resourceAwsIamAccountAlias(),
			"aws_iam_account_password_policy":                         resourceAwsIamAccountPasswordPolicy(),
			"aws_iam_group_policy":                                    resourceAwsIamGroupPolicy(),
			"aws_iam_group":                                           resourceAwsIamGroup(),
			"aws_iam_group_membership":                                resourceAwsIamGroupMembership(),
			"aws_iam_group_policy_attachment":                         resourceAwsIamGroupPolicyAttachment(),
			"aws_iam_instance_profile":                                resourceAwsIamInstanceProfile(),
			"aws_iam_openid_connect_provider":                         resourceAwsIamOpenIDConnectProvider(),
			"aws_iam_policy":                                          resourceAwsIamPolicy(),
			"aws_iam_policy_attachment":                               resourceAwsIamPolicyAttachment(),
			"aws_iam_role_policy_attachment":                          resourceAwsIamRolePolicyAttachment(),
			"aws_iam_role_policy":                                     resourceAwsIamRolePolicy(),
			"aws_iam_role":                                            resourceAwsIamRole(),
			"aws_iam_saml_provider":                                   resourceAwsIamSamlProvider(),
			"aws_iam_server_certificate":                              resourceAwsIAMServerCertificate(),
			"aws_iam_service_linked_role":                             resourceAwsIamServiceLinkedRole(),
			"aws_iam_user_group_membership":                           resourceAwsIamUserGroupMembership(),
			"aws_iam_user_policy_attachment":                          resourceAwsIamUserPolicyAttachment(),
			"aws_iam_user_policy":                                     resourceAwsIamUserPolicy(),
			"aws_iam_user_ssh_key":                                    resourceAwsIamUserSshKey(),
			"aws_iam_user":                                            resourceAwsIamUser(),
			"aws_iam_user_login_profile":                              resourceAwsIamUserLoginProfile(),
			"aws_imagebuilder_component":                              resourceAwsImageBuilderComponent(),
			"aws_imagebuilder_distribution_configuration":             resourceAwsImageBuilderDistributionConfiguration(),
			"aws_imagebuilder_image":                                  resourceAwsImageBuilderImage(),
			"aws_imagebuilder_image_pipeline":                         resourceAwsImageBuilderImagePipeline(),
			"aws_imagebuilder_image_recipe":                           resourceAwsImageBuilderImageRecipe(),
			"aws_imagebuilder_infrastructure_configuration":           resourceAwsImageBuilderInfrastructureConfiguration(),
			"aws_inspector_assessment_target":                         resourceAWSInspectorAssessmentTarget(),
			"aws_inspector_assessment_template":                       resourceAWSInspectorAssessmentTemplate(),
			"aws_inspector_resource_group":                            resourceAWSInspectorResourceGroup(),
			"aws_instance":                                            resourceAwsInstance(),
			"aws_internet_gateway":                                    resourceAwsInternetGateway(),
			"aws_iot_certificate":                                     resourceAwsIotCertificate(),
			"aws_iot_policy":                                          resourceAwsIotPolicy(),
			"aws_iot_policy_attachment":                               resourceAwsIotPolicyAttachment(),
			"aws_iot_thing":                                           resourceAwsIotThing(),
			"aws_iot_thing_principal_attachment":                      resourceAwsIotThingPrincipalAttachment(),
			"aws_iot_thing_type":                                      resourceAwsIotThingType(),
			"aws_iot_topic_rule":                                      resourceAwsIotTopicRule(),
			"aws_iot_role_alias":                                      resourceAwsIotRoleAlias(),
			"aws_key_pair":                                            resourceAwsKeyPair(),
			"aws_kinesis_analytics_application":                       resourceAwsKinesisAnalyticsApplication(),
			"aws_kinesisanalyticsv2_application":                      resourceAwsKinesisAnalyticsV2Application(),
			"aws_kinesisanalyticsv2_application_snapshot":             resourceAwsKinesisAnalyticsV2ApplicationSnapshot(),
			"aws_kinesis_firehose_delivery_stream":                    resourceAwsKinesisFirehoseDeliveryStream(),
			"aws_kinesis_stream":                                      resourceAwsKinesisStream(),
			"aws_kinesis_stream_consumer":                             resourceAwsKinesisStreamConsumer(),
			"aws_kinesis_video_stream":                                resourceAwsKinesisVideoStream(),
			"aws_kms_alias":                                           resourceAwsKmsAlias(),
			"aws_kms_external_key":                                    resourceAwsKmsExternalKey(),
			"aws_kms_grant":                                           resourceAwsKmsGrant(),
			"aws_kms_key":                                             resourceAwsKmsKey(),
			"aws_kms_ciphertext":                                      resourceAwsKmsCiphertext(),
			"aws_lakeformation_data_lake_settings":                    resourceAwsLakeFormationDataLakeSettings(),
			"aws_lakeformation_permissions":                           resourceAwsLakeFormationPermissions(),
			"aws_lakeformation_resource":                              resourceAwsLakeFormationResource(),
			"aws_lambda_alias":                                        resourceAwsLambdaAlias(),
			"aws_lambda_code_signing_config":                          resourceAwsLambdaCodeSigningConfig(),
			"aws_lambda_event_source_mapping":                         resourceAwsLambdaEventSourceMapping(),
			"aws_lambda_function_event_invoke_config":                 resourceAwsLambdaFunctionEventInvokeConfig(),
			"aws_lambda_function":                                     resourceAwsLambdaFunction(),
			"aws_lambda_layer_version":                                resourceAwsLambdaLayerVersion(),
			"aws_lambda_permission":                                   resourceAwsLambdaPermission(),
			"aws_lambda_provisioned_concurrency_config":               resourceAwsLambdaProvisionedConcurrencyConfig(),
			"aws_launch_configuration":                                resourceAwsLaunchConfiguration(),
			"aws_launch_template":                                     resourceAwsLaunchTemplate(),
			"aws_lex_bot":                                             resourceAwsLexBot(),
			"aws_lex_bot_alias":                                       resourceAwsLexBotAlias(),
			"aws_lex_intent":                                          resourceAwsLexIntent(),
			"aws_lex_slot_type":                                       resourceAwsLexSlotType(),
			"aws_licensemanager_association":                          resourceAwsLicenseManagerAssociation(),
			"aws_licensemanager_license_configuration":                resourceAwsLicenseManagerLicenseConfiguration(),
			"aws_lightsail_domain":                                    resourceAwsLightsailDomain(),
			"aws_lightsail_instance":                                  resourceAwsLightsailInstance(),
			"aws_lightsail_instance_public_ports":                     resourceAwsLightsailInstancePublicPorts(),
			"aws_lightsail_key_pair":                                  resourceAwsLightsailKeyPair(),
			"aws_lightsail_static_ip":                                 resourceAwsLightsailStaticIp(),
			"aws_lightsail_static_ip_attachment":                      resourceAwsLightsailStaticIpAttachment(),
			"aws_lb_cookie_stickiness_policy":                         resourceAwsLBCookieStickinessPolicy(),
			"aws_load_balancer_policy":                                resourceAwsLoadBalancerPolicy(),
			"aws_load_balancer_backend_server_policy":                 resourceAwsLoadBalancerBackendServerPolicies(),
			"aws_load_balancer_listener_policy":                       resourceAwsLoadBalancerListenerPolicies(),
			"aws_lb_ssl_negotiation_policy":                           resourceAwsLBSSLNegotiationPolicy(),
			"aws_macie2_account":                                      resourceAwsMacie2Account(),
			"aws_macie2_classification_job":                           resourceAwsMacie2ClassificationJob(),
			"aws_macie2_custom_data_identifier":                       resourceAwsMacie2CustomDataIdentifier(),
			"aws_macie2_findings_filter":                              resourceAwsMacie2FindingsFilter(),
			"aws_macie2_invitation_accepter":                          resourceAwsMacie2InvitationAccepter(),
			"aws_macie2_member":                                       resourceAwsMacie2Member(),
			"aws_macie2_organization_admin_account":                   resourceAwsMacie2OrganizationAdminAccount(),
			"aws_macie_member_account_association":                    resourceAwsMacieMemberAccountAssociation(),
			"aws_macie_s3_bucket_association":                         resourceAwsMacieS3BucketAssociation(),
			"aws_main_route_table_association":                        resourceAwsMainRouteTableAssociation(),
			"aws_mq_broker":                                           resourceAwsMqBroker(),
			"aws_mq_configuration":                                    resourceAwsMqConfiguration(),
			"aws_media_convert_queue":                                 resourceAwsMediaConvertQueue(),
			"aws_media_package_channel":                               resourceAwsMediaPackageChannel(),
			"aws_media_store_container":                               resourceAwsMediaStoreContainer(),
			"aws_media_store_container_policy":                        resourceAwsMediaStoreContainerPolicy(),
			"aws_msk_cluster":                                         resourceAwsMskCluster(),
			"aws_msk_configuration":                                   resourceAwsMskConfiguration(),
			"aws_msk_scram_secret_association":                        resourceAwsMskScramSecretAssociation(),
			"aws_mwaa_environment":                                    resourceAwsMwaaEnvironment(),
			"aws_nat_gateway":                                         resourceAwsNatGateway(),
			"aws_network_acl":                                         resourceAwsNetworkAcl(),
			"aws_default_network_acl":                                 resourceAwsDefaultNetworkAcl(),
			"aws_neptune_cluster":                                     resourceAwsNeptuneCluster(),
			"aws_neptune_cluster_instance":                            resourceAwsNeptuneClusterInstance(),
			"aws_neptune_cluster_parameter_group":                     resourceAwsNeptuneClusterParameterGroup(),
			"aws_neptune_cluster_snapshot":                            resourceAwsNeptuneClusterSnapshot(),
			"aws_neptune_event_subscription":                          resourceAwsNeptuneEventSubscription(),
			"aws_neptune_parameter_group":                             resourceAwsNeptuneParameterGroup(),
			"aws_neptune_subnet_group":                                resourceAwsNeptuneSubnetGroup(),
			"aws_network_acl_rule":                                    resourceAwsNetworkAclRule(),
			"aws_network_interface":                                   resourceAwsNetworkInterface(),
			"aws_network_interface_attachment":                        resourceAwsNetworkInterfaceAttachment(),
			"aws_networkfirewall_firewall":                            resourceAwsNetworkFirewallFirewall(),
			"aws_networkfirewall_firewall_policy":                     resourceAwsNetworkFirewallFirewallPolicy(),
			"aws_networkfirewall_logging_configuration":               resourceAwsNetworkFirewallLoggingConfiguration(),
			"aws_networkfirewall_resource_policy":                     resourceAwsNetworkFirewallResourcePolicy(),
			"aws_networkfirewall_rule_group":                          resourceAwsNetworkFirewallRuleGroup(),
			"aws_opsworks_application":                                resourceAwsOpsworksApplication(),
			"aws_opsworks_stack":                                      resourceAwsOpsworksStack(),
			"aws_opsworks_java_app_layer":                             resourceAwsOpsworksJavaAppLayer(),
			"aws_opsworks_haproxy_layer":                              resourceAwsOpsworksHaproxyLayer(),
			"aws_opsworks_static_web_layer":                           resourceAwsOpsworksStaticWebLayer(),
			"aws_opsworks_php_app_layer":                              resourceAwsOpsworksPhpAppLayer(),
			"aws_opsworks_rails_app_layer":                            resourceAwsOpsworksRailsAppLayer(),
			"aws_opsworks_nodejs_app_layer":                           resourceAwsOpsworksNodejsAppLayer(),
			"aws_opsworks_memcached_layer":                            resourceAwsOpsworksMemcachedLayer(),
			"aws_opsworks_mysql_layer":                                resourceAwsOpsworksMysqlLayer(),
			"aws_opsworks_ganglia_layer":                              resourceAwsOpsworksGangliaLayer(),
			"aws_opsworks_custom_layer":                               resourceAwsOpsworksCustomLayer(),
			"aws_opsworks_instance":                                   resourceAwsOpsworksInstance(),
			"aws_opsworks_user_profile":                               resourceAwsOpsworksUserProfile(),
			"aws_opsworks_permission":                                 resourceAwsOpsworksPermission(),
			"aws_opsworks_rds_db_instance":                            resourceAwsOpsworksRdsDbInstance(),
			"aws_organizations_organization":                          resourceAwsOrganizationsOrganization(),
			"aws_organizations_account":                               resourceAwsOrganizationsAccount(),
			"aws_organizations_delegated_administrator":               resourceAwsOrganizationsDelegatedAdministrator(),
			"aws_organizations_policy":                                resourceAwsOrganizationsPolicy(),
			"aws_organizations_policy_attachment":                     resourceAwsOrganizationsPolicyAttachment(),
			"aws_organizations_organizational_unit":                   resourceAwsOrganizationsOrganizationalUnit(),
			"aws_placement_group":                                     resourceAwsPlacementGroup(),
			"aws_prometheus_workspace":                                resourceAwsPrometheusWorkspace(),
			"aws_proxy_protocol_policy":                               resourceAwsProxyProtocolPolicy(),
			"aws_qldb_ledger":                                         resourceAwsQLDBLedger(),
			"aws_quicksight_group":                                    resourceAwsQuickSightGroup(),
			"aws_quicksight_user":                                     resourceAwsQuickSightUser(),
			"aws_ram_principal_association":                           resourceAwsRamPrincipalAssociation(),
			"aws_ram_resource_association":                            resourceAwsRamResourceAssociation(),
			"aws_ram_resource_share":                                  resourceAwsRamResourceShare(),
			"aws_ram_resource_share_accepter":                         resourceAwsRamResourceShareAccepter(),
			"aws_rds_cluster":                                         resourceAwsRDSCluster(),
			"aws_rds_cluster_endpoint":                                resourceAwsRDSClusterEndpoint(),
			"aws_rds_cluster_instance":                                resourceAwsRDSClusterInstance(),
			"aws_rds_cluster_parameter_group":                         resourceAwsRDSClusterParameterGroup(),
			"aws_rds_global_cluster":                                  resourceAwsRDSGlobalCluster(),
			"aws_redshift_cluster":                                    resourceAwsRedshiftCluster(),
			"aws_redshift_security_group":                             resourceAwsRedshiftSecurityGroup(),
			"aws_redshift_parameter_group":                            resourceAwsRedshiftParameterGroup(),
			"aws_redshift_subnet_group":                               resourceAwsRedshiftSubnetGroup(),
			"aws_redshift_snapshot_copy_grant":                        resourceAwsRedshiftSnapshotCopyGrant(),
			"aws_redshift_snapshot_schedule":                          resourceAwsRedshiftSnapshotSchedule(),
			"aws_redshift_snapshot_schedule_association":              resourceAwsRedshiftSnapshotScheduleAssociation(),
			"aws_redshift_event_subscription":                         resourceAwsRedshiftEventSubscription(),
			"aws_resourcegroups_group":                                resourceAwsResourceGroupsGroup(),
			"aws_route53_delegation_set":                              resourceAwsRoute53DelegationSet(),
			"aws_route53_hosted_zone_dnssec":                          resourceAwsRoute53HostedZoneDnssec(),
			"aws_route53_key_signing_key":                             resourceAwsRoute53KeySigningKey(),
			"aws_route53_query_log":                                   resourceAwsRoute53QueryLog(),
			"aws_route53_record":                                      resourceAwsRoute53Record(),
			"aws_route53_zone_association":                            resourceAwsRoute53ZoneAssociation(),
			"aws_route53_vpc_association_authorization":               resourceAwsRoute53VPCAssociationAuthorization(),
			"aws_route53_zone":                                        resourceAwsRoute53Zone(),
			"aws_route53_health_check":                                resourceAwsRoute53HealthCheck(),
			"aws_route53_resolver_dnssec_config":                      resourceAwsRoute53ResolverDnssecConfig(),
			"aws_route53_resolver_endpoint":                           resourceAwsRoute53ResolverEndpoint(),
			"aws_route53_resolver_firewall_domain_list":               resourceAwsRoute53ResolverFirewallDomainList(),
			"aws_route53_resolver_firewall_rule":                      resourceAwsRoute53ResolverFirewallRule(),
			"aws_route53_resolver_firewall_rule_group":                resourceAwsRoute53ResolverFirewallRuleGroup(),
			"aws_route53_resolver_firewall_rule_group_association":    resourceAwsRoute53ResolverFirewallRuleGroupAssociation(),
			"aws_route53_resolver_query_log_config":                   resourceAwsRoute53ResolverQueryLogConfig(),
			"aws_route53_resolver_query_log_config_association":       resourceAwsRoute53ResolverQueryLogConfigAssociation(),
			"aws_route53_resolver_rule_association":                   resourceAwsRoute53ResolverRuleAssociation(),
			"aws_route53_resolver_rule":                               resourceAwsRoute53ResolverRule(),
			"aws_route":                                               resourceAwsRoute(),
			"aws_route_table":                                         resourceAwsRouteTable(),
			"aws_default_route_table":                                 resourceAwsDefaultRouteTable(),
			"aws_route_table_association":                             resourceAwsRouteTableAssociation(),
			"aws_sagemaker_app":                                       resourceAwsSagemakerApp(),
			"aws_sagemaker_app_image_config":                          resourceAwsSagemakerAppImageConfig(),
			"aws_sagemaker_code_repository":                           resourceAwsSagemakerCodeRepository(),
			"aws_sagemaker_domain":                                    resourceAwsSagemakerDomain(),
			"aws_sagemaker_endpoint":                                  resourceAwsSagemakerEndpoint(),
			"aws_sagemaker_endpoint_configuration":                    resourceAwsSagemakerEndpointConfiguration(),
			"aws_sagemaker_feature_group":                             resourceAwsSagemakerFeatureGroup(),
			"aws_sagemaker_image":                                     resourceAwsSagemakerImage(),
			"aws_sagemaker_image_version":                             resourceAwsSagemakerImageVersion(),
			"aws_sagemaker_model":                                     resourceAwsSagemakerModel(),
			"aws_sagemaker_model_package_group":                       resourceAwsSagemakerModelPackageGroup(),
			"aws_sagemaker_notebook_instance_lifecycle_configuration": resourceAwsSagemakerNotebookInstanceLifeCycleConfiguration(),
			"aws_sagemaker_notebook_instance":                         resourceAwsSagemakerNotebookInstance(),
			"aws_sagemaker_user_profile":                              resourceAwsSagemakerUserProfile(),
			"aws_schemas_discoverer":                                  resourceAwsSchemasDiscoverer(),
			"aws_schemas_registry":                                    resourceAwsSchemasRegistry(),
			"aws_schemas_schema":                                      resourceAwsSchemasSchema(),
			"aws_secretsmanager_secret":                               resourceAwsSecretsManagerSecret(),
			"aws_secretsmanager_secret_policy":                        resourceAwsSecretsManagerSecretPolicy(),
			"aws_secretsmanager_secret_version":                       resourceAwsSecretsManagerSecretVersion(),
			"aws_secretsmanager_secret_rotation":                      resourceAwsSecretsManagerSecretRotation(),
			"aws_ses_active_receipt_rule_set":                         resourceAwsSesActiveReceiptRuleSet(),
			"aws_ses_domain_identity":                                 resourceAwsSesDomainIdentity(),
			"aws_ses_domain_identity_verification":                    resourceAwsSesDomainIdentityVerification(),
			"aws_ses_domain_dkim":                                     resourceAwsSesDomainDkim(),
			"aws_ses_domain_mail_from":                                resourceAwsSesDomainMailFrom(),
			"aws_ses_email_identity":                                  resourceAwsSesEmailIdentity(),
			"aws_ses_identity_policy":                                 resourceAwsSesIdentityPolicy(),
			"aws_ses_receipt_filter":                                  resourceAwsSesReceiptFilter(),
			"aws_ses_receipt_rule":                                    resourceAwsSesReceiptRule(),
			"aws_ses_receipt_rule_set":                                resourceAwsSesReceiptRuleSet(),
			"aws_ses_configuration_set":                               resourceAwsSesConfigurationSet(),
			"aws_ses_event_destination":                               resourceAwsSesEventDestination(),
			"aws_ses_identity_notification_topic":                     resourceAwsSesNotificationTopic(),
			"aws_ses_template":                                        resourceAwsSesTemplate(),
			"aws_s3_access_point":                                     resourceAwsS3AccessPoint(),
			"aws_s3_account_public_access_block":                      resourceAwsS3AccountPublicAccessBlock(),
			"aws_s3_bucket":                                           resourceAwsS3Bucket(),
			"aws_s3_bucket_analytics_configuration":                   resourceAwsS3BucketAnalyticsConfiguration(),
			"aws_s3_bucket_policy":                                    resourceAwsS3BucketPolicy(),
			"aws_s3_bucket_public_access_block":                       resourceAwsS3BucketPublicAccessBlock(),
			"aws_s3_bucket_object":                                    resourceAwsS3BucketObject(),
			"aws_s3_bucket_ownership_controls":                        resourceAwsS3BucketOwnershipControls(),
			"aws_s3_bucket_notification":                              resourceAwsS3BucketNotification(),
			"aws_s3_bucket_metric":                                    resourceAwsS3BucketMetric(),
			"aws_s3_bucket_inventory":                                 resourceAwsS3BucketInventory(),
			"aws_s3_object_copy":                                      resourceAwsS3ObjectCopy(),
			"aws_s3control_bucket":                                    resourceAwsS3ControlBucket(),
			"aws_s3control_bucket_policy":                             resourceAwsS3ControlBucketPolicy(),
			"aws_s3control_bucket_lifecycle_configuration":            resourceAwsS3ControlBucketLifecycleConfiguration(),
			"aws_s3outposts_endpoint":                                 resourceAwsS3OutpostsEndpoint(),
			"aws_security_group":                                      resourceAwsSecurityGroup(),
			"aws_network_interface_sg_attachment":                     resourceAwsNetworkInterfaceSGAttachment(),
			"aws_default_security_group":                              resourceAwsDefaultSecurityGroup(),
			"aws_security_group_rule":                                 resourceAwsSecurityGroupRule(),
			"aws_securityhub_account":                                 resourceAwsSecurityHubAccount(),
			"aws_securityhub_action_target":                           resourceAwsSecurityHubActionTarget(),
			"aws_securityhub_insight":                                 resourceAwsSecurityHubInsight(),
			"aws_securityhub_invite_accepter":                         resourceAwsSecurityHubInviteAccepter(),
			"aws_securityhub_member":                                  resourceAwsSecurityHubMember(),
			"aws_securityhub_organization_admin_account":              resourceAwsSecurityHubOrganizationAdminAccount(),
			"aws_securityhub_product_subscription":                    resourceAwsSecurityHubProductSubscription(),
			"aws_securityhub_standards_subscription":                  resourceAwsSecurityHubStandardsSubscription(),
			"aws_servicecatalog_budget_resource_association":          resourceAwsServiceCatalogBudgetResourceAssociation(),
			"aws_servicecatalog_constraint":                           resourceAwsServiceCatalogConstraint(),
			"aws_servicecatalog_organizations_access":                 resourceAwsServiceCatalogOrganizationsAccess(),
			"aws_servicecatalog_portfolio":                            resourceAwsServiceCatalogPortfolio(),
			"aws_servicecatalog_portfolio_share":                      resourceAwsServiceCatalogPortfolioShare(),
			"aws_servicecatalog_product":                              resourceAwsServiceCatalogProduct(),
			"aws_servicecatalog_service_action":                       resourceAwsServiceCatalogServiceAction(),
			"aws_servicecatalog_tag_option":                           resourceAwsServiceCatalogTagOption(),
			"aws_servicecatalog_tag_option_resource_association":      resourceAwsServiceCatalogTagOptionResourceAssociation(),
			"aws_servicecatalog_principal_portfolio_association":      resourceAwsServiceCatalogPrincipalPortfolioAssociation(),
			"aws_servicecatalog_product_portfolio_association":        resourceAwsServiceCatalogProductPortfolioAssociation(),
			"aws_servicecatalog_provisioning_artifact":                resourceAwsServiceCatalogProvisioningArtifact(),
			"aws_service_discovery_http_namespace":                    resourceAwsServiceDiscoveryHttpNamespace(),
			"aws_service_discovery_private_dns_namespace":             resourceAwsServiceDiscoveryPrivateDnsNamespace(),
			"aws_service_discovery_public_dns_namespace":              resourceAwsServiceDiscoveryPublicDnsNamespace(),
			"aws_service_discovery_service":                           resourceAwsServiceDiscoveryService(),
			"aws_servicequotas_service_quota":                         resourceAwsServiceQuotasServiceQuota(),
			"aws_shield_protection":                                   resourceAwsShieldProtection(),
			"aws_signer_signing_job":                                  resourceAwsSignerSigningJob(),
			"aws_signer_signing_profile":                              resourceAwsSignerSigningProfile(),
			"aws_signer_signing_profile_permission":                   resourceAwsSignerSigningProfilePermission(),
			"aws_simpledb_domain":                                     resourceAwsSimpleDBDomain(),
			"aws_ssm_activation":                                      resourceAwsSsmActivation(),
			"aws_ssm_association":                                     resourceAwsSsmAssociation(),
			"aws_ssm_document":                                        resourceAwsSsmDocument(),
			"aws_ssm_maintenance_window":                              resourceAwsSsmMaintenanceWindow(),
			"aws_ssm_maintenance_window_target":                       resourceAwsSsmMaintenanceWindowTarget(),
			"aws_ssm_maintenance_window_task":                         resourceAwsSsmMaintenanceWindowTask(),
			"aws_ssm_patch_baseline":                                  resourceAwsSsmPatchBaseline(),
			"aws_ssm_patch_group":                                     resourceAwsSsmPatchGroup(),
			"aws_ssm_parameter":                                       resourceAwsSsmParameter(),
			"aws_ssm_resource_data_sync":                              resourceAwsSsmResourceDataSync(),
			"aws_ssoadmin_account_assignment":                         resourceAwsSsoAdminAccountAssignment(),
			"aws_ssoadmin_managed_policy_attachment":                  resourceAwsSsoAdminManagedPolicyAttachment(),
			"aws_ssoadmin_permission_set":                             resourceAwsSsoAdminPermissionSet(),
			"aws_ssoadmin_permission_set_inline_policy":               resourceAwsSsoAdminPermissionSetInlinePolicy(),
			"aws_storagegateway_cache":                                resourceAwsStorageGatewayCache(),
			"aws_storagegateway_cached_iscsi_volume":                  resourceAwsStorageGatewayCachedIscsiVolume(),
			"aws_storagegateway_gateway":                              resourceAwsStorageGatewayGateway(),
			"aws_storagegateway_nfs_file_share":                       resourceAwsStorageGatewayNfsFileShare(),
			"aws_storagegateway_smb_file_share":                       resourceAwsStorageGatewaySmbFileShare(),
			"aws_storagegateway_stored_iscsi_volume":                  resourceAwsStorageGatewayStoredIscsiVolume(),
			"aws_storagegateway_tape_pool":                            resourceAwsStorageGatewayTapePool(),
			"aws_storagegateway_upload_buffer":                        resourceAwsStorageGatewayUploadBuffer(),
			"aws_storagegateway_working_storage":                      resourceAwsStorageGatewayWorkingStorage(),
			"aws_spot_datafeed_subscription":                          resourceAwsSpotDataFeedSubscription(),
			"aws_spot_instance_request":                               resourceAwsSpotInstanceRequest(),
			"aws_spot_fleet_request":                                  resourceAwsSpotFleetRequest(),
			"aws_sqs_queue":                                           resourceAwsSqsQueue(),
			"aws_sqs_queue_policy":                                    resourceAwsSqsQueuePolicy(),
			"aws_snapshot_create_volume_permission":                   resourceAwsSnapshotCreateVolumePermission(),
			"aws_sns_platform_application":                            resourceAwsSnsPlatformApplication(),
			"aws_sns_sms_preferences":                                 resourceAwsSnsSmsPreferences(),
			"aws_sns_topic":                                           resourceAwsSnsTopic(),
			"aws_sns_topic_policy":                                    resourceAwsSnsTopicPolicy(),
			"aws_sns_topic_subscription":                              resourceAwsSnsTopicSubscription(),
			"aws_sfn_activity":                                        resourceAwsSfnActivity(),
			"aws_sfn_state_machine":                                   resourceAwsSfnStateMachine(),
			"aws_default_subnet":                                      resourceAwsDefaultSubnet(),
			"aws_subnet":                                              resourceAwsSubnet(),
			"aws_swf_domain":                                          resourceAwsSwfDomain(),
			"aws_synthetics_canary":                                   resourceAwsSyntheticsCanary(),
			"aws_timestreamwrite_database":                            resourceAwsTimestreamWriteDatabase(),
			"aws_timestreamwrite_table":                               resourceAwsTimestreamWriteTable(),
			"aws_transfer_server":                                     resourceAwsTransferServer(),
			"aws_transfer_ssh_key":                                    resourceAwsTransferSshKey(),
			"aws_transfer_user":                                       resourceAwsTransferUser(),
			"aws_volume_attachment":                                   resourceAwsVolumeAttachment(),
			"aws_vpc_dhcp_options_association":                        resourceAwsVpcDhcpOptionsAssociation(),
			"aws_default_vpc_dhcp_options":                            resourceAwsDefaultVpcDhcpOptions(),
			"aws_vpc_dhcp_options":                                    resourceAwsVpcDhcpOptions(),
			"aws_vpc_peering_connection":                              resourceAwsVpcPeeringConnection(),
			"aws_vpc_peering_connection_accepter":                     resourceAwsVpcPeeringConnectionAccepter(),
			"aws_vpc_peering_connection_options":                      resourceAwsVpcPeeringConnectionOptions(),
			"aws_default_vpc":                                         resourceAwsDefaultVpc(),
			"aws_vpc":                                                 resourceAwsVpc(),
			"aws_vpc_endpoint":                                        resourceAwsVpcEndpoint(),
			"aws_vpc_endpoint_connection_notification":                resourceAwsVpcEndpointConnectionNotification(),
			"aws_vpc_endpoint_route_table_association":                resourceAwsVpcEndpointRouteTableAssociation(),
			"aws_vpc_endpoint_subnet_association":                     resourceAwsVpcEndpointSubnetAssociation(),
			"aws_vpc_endpoint_service":                                resourceAwsVpcEndpointService(),
			"aws_vpc_endpoint_service_allowed_principal":              resourceAwsVpcEndpointServiceAllowedPrincipal(),
			"aws_vpc_ipv4_cidr_block_association":                     resourceAwsVpcIpv4CidrBlockAssociation(),
			"aws_vpn_connection":                                      resourceAwsVpnConnection(),
			"aws_vpn_connection_route":                                resourceAwsVpnConnectionRoute(),
			"aws_vpn_gateway":                                         resourceAwsVpnGateway(),
			"aws_vpn_gateway_attachment":                              resourceAwsVpnGatewayAttachment(),
			"aws_vpn_gateway_route_propagation":                       resourceAwsVpnGatewayRoutePropagation(),
			"aws_waf_byte_match_set":                                  resourceAwsWafByteMatchSet(),
			"aws_waf_ipset":                                           resourceAwsWafIPSet(),
			"aws_waf_rate_based_rule":                                 resourceAwsWafRateBasedRule(),
			"aws_waf_regex_match_set":                                 resourceAwsWafRegexMatchSet(),
			"aws_waf_regex_pattern_set":                               resourceAwsWafRegexPatternSet(),
			"aws_waf_rule":                                            resourceAwsWafRule(),
			"aws_waf_rule_group":                                      resourceAwsWafRuleGroup(),
			"aws_waf_size_constraint_set":                             resourceAwsWafSizeConstraintSet(),
			"aws_waf_web_acl":                                         resourceAwsWafWebAcl(),
			"aws_waf_xss_match_set":                                   resourceAwsWafXssMatchSet(),
			"aws_waf_sql_injection_match_set":                         resourceAwsWafSqlInjectionMatchSet(),
			"aws_waf_geo_match_set":                                   resourceAwsWafGeoMatchSet(),
			"aws_wafregional_byte_match_set":                          resourceAwsWafRegionalByteMatchSet(),
			"aws_wafregional_geo_match_set":                           resourceAwsWafRegionalGeoMatchSet(),
			"aws_wafregional_ipset":                                   resourceAwsWafRegionalIPSet(),
			"aws_wafregional_rate_based_rule":                         resourceAwsWafRegionalRateBasedRule(),
			"aws_wafregional_regex_match_set":                         resourceAwsWafRegionalRegexMatchSet(),
			"aws_wafregional_regex_pattern_set":                       resourceAwsWafRegionalRegexPatternSet(),
			"aws_wafregional_rule":                                    resourceAwsWafRegionalRule(),
			"aws_wafregional_rule_group":                              resourceAwsWafRegionalRuleGroup(),
			"aws_wafregional_size_constraint_set":                     resourceAwsWafRegionalSizeConstraintSet(),
			"aws_wafregional_sql_injection_match_set":                 resourceAwsWafRegionalSqlInjectionMatchSet(),
			"aws_wafregional_xss_match_set":                           resourceAwsWafRegionalXssMatchSet(),
			"aws_wafregional_web_acl":                                 resourceAwsWafRegionalWebAcl(),
			"aws_wafregional_web_acl_association":                     resourceAwsWafRegionalWebAclAssociation(),
			"aws_wafv2_ip_set":                                        resourceAwsWafv2IPSet(),
			"aws_wafv2_regex_pattern_set":                             resourceAwsWafv2RegexPatternSet(),
			"aws_wafv2_rule_group":                                    resourceAwsWafv2RuleGroup(),
			"aws_wafv2_web_acl":                                       resourceAwsWafv2WebACL(),
			"aws_wafv2_web_acl_association":                           resourceAwsWafv2WebACLAssociation(),
			"aws_wafv2_web_acl_logging_configuration":                 resourceAwsWafv2WebACLLoggingConfiguration(),
			"aws_worklink_fleet":                                      resourceAwsWorkLinkFleet(),
			"aws_worklink_website_certificate_authority_association":  resourceAwsWorkLinkWebsiteCertificateAuthorityAssociation(),
			"aws_workspaces_directory":                                resourceAwsWorkspacesDirectory(),
			"aws_workspaces_workspace":                                resourceAwsWorkspacesWorkspace(),
			"aws_batch_compute_environment":                           resourceAwsBatchComputeEnvironment(),
			"aws_batch_job_definition":                                resourceAwsBatchJobDefinition(),
			"aws_batch_job_queue":                                     resourceAwsBatchJobQueue(),
			"aws_pinpoint_app":                                        resourceAwsPinpointApp(),
			"aws_pinpoint_adm_channel":                                resourceAwsPinpointADMChannel(),
			"aws_pinpoint_apns_channel":                               resourceAwsPinpointAPNSChannel(),
			"aws_pinpoint_apns_sandbox_channel":                       resourceAwsPinpointAPNSSandboxChannel(),
			"aws_pinpoint_apns_voip_channel":                          resourceAwsPinpointAPNSVoipChannel(),
			"aws_pinpoint_apns_voip_sandbox_channel":                  resourceAwsPinpointAPNSVoipSandboxChannel(),
			"aws_pinpoint_baidu_channel":                              resourceAwsPinpointBaiduChannel(),
			"aws_pinpoint_email_channel":                              resourceAwsPinpointEmailChannel(),
			"aws_pinpoint_event_stream":                               resourceAwsPinpointEventStream(),
			"aws_pinpoint_gcm_channel":                                resourceAwsPinpointGCMChannel(),
			"aws_pinpoint_sms_channel":                                resourceAwsPinpointSMSChannel(),
			"aws_xray_encryption_config":                              resourceAwsXrayEncryptionConfig(),
			"aws_xray_group":                                          resourceAwsXrayGroup(),
			"aws_xray_sampling_rule":                                  resourceAwsXraySamplingRule(),
			"aws_workspaces_ip_group":                                 resourceAwsWorkspacesIpGroup(),

			// ALBs are actually LBs because they can be type `network` or `application`
			// To avoid regressions, we will add a new resource for each and they both point
			// back to the old ALB version. IF the Terraform supported aliases for resources
			// this would be a whole lot simpler
			"aws_alb":                         resourceAwsLb(),
			"aws_lb":                          resourceAwsLb(),
			"aws_alb_listener":                resourceAwsLbListener(),
			"aws_lb_listener":                 resourceAwsLbListener(),
			"aws_alb_listener_certificate":    resourceAwsLbListenerCertificate(),
			"aws_lb_listener_certificate":     resourceAwsLbListenerCertificate(),
			"aws_alb_listener_rule":           resourceAwsLbbListenerRule(),
			"aws_lb_listener_rule":            resourceAwsLbbListenerRule(),
			"aws_alb_target_group":            resourceAwsLbTargetGroup(),
			"aws_lb_target_group":             resourceAwsLbTargetGroup(),
			"aws_alb_target_group_attachment": resourceAwsLbTargetGroupAttachment(),
			"aws_lb_target_group_attachment":  resourceAwsLbTargetGroupAttachment(),
		},
	}

	// Avoid Go formatting churn and Git conflicts
	// You probably should not do this
	provider.DataSourcesMap["aws_serverlessapplicationrepository_application"] = dataSourceAwsServerlessApplicationRepositoryApplication()
	provider.ResourcesMap["aws_serverlessapplicationrepository_cloudformation_stack"] = resourceAwsServerlessApplicationRepositoryCloudFormationStack()

	// Add in service package data sources and resources.
	servicePackages, err := tfprovider.ServicePackages()

	if err != nil {
		panic(err)
	}

	for serviceName, servicePackage := range servicePackages {
		for name, ds := range servicePackage.DataSources() {
			if _, exists := provider.DataSourcesMap[name]; exists {
				panic(fmt.Sprintf("(%s) A data source named %q is already registered", serviceName, name))
			}

			provider.DataSourcesMap[name] = ds
		}

		for name, res := range servicePackage.Resources() {
			if _, exists := provider.ResourcesMap[name]; exists {
				panic(fmt.Sprintf("(%s) A resource named %q is already registered", serviceName, name))
			}

			provider.ResourcesMap[name] = res
		}
	}

	// Custom endpoints.
	customEndpoints := make(map[string]struct{})

	for _, endpointServiceName := range endpointServiceNames {
		if _, ok := customEndpoints[endpointServiceName]; ok {
			panic(fmt.Sprintf("A service named %q is already registered for custom endpoints", endpointServiceName))
		}

		customEndpoints[endpointServiceName] = struct{}{}
	}

	for serviceName, servicePackage := range servicePackages {
		endpointServiceName := servicePackage.CustomEndpointKey()

		if _, ok := customEndpoints[endpointServiceName]; ok {
			panic(fmt.Sprintf("(%s) A service named %q is already registered for custom endpoints", serviceName, endpointServiceName))
		}

		customEndpoints[endpointServiceName] = struct{}{}
	}

	endpointServiceNames = make([]string, len(customEndpoints))

	for endpointServiceName := range customEndpoints {
		endpointServiceNames = append(endpointServiceNames, endpointServiceName)
	}

	provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
		terraformVersion := provider.TerraformVersion
		if terraformVersion == "" {
			// Terraform 0.12 introduced this field to the protocol
			// We can therefore assume that if it's missing it's 0.10 or 0.11
			terraformVersion = "0.11+compatible"
		}
		return providerConfigure(ctx, d, terraformVersion)
	}

	return provider
}

var descriptions map[string]string
var endpointServiceNames []string

func init() {
	descriptions = map[string]string{
		"region": "The region where AWS operations will take place. Examples\n" +
			"are us-east-1, us-west-2, etc.", // lintignore:AWSAT003

		"access_key": "The access key for API operations. You can retrieve this\n" +
			"from the 'Security & Credentials' section of the AWS console.",

		"secret_key": "The secret key for API operations. You can retrieve this\n" +
			"from the 'Security & Credentials' section of the AWS console.",

		"profile": "The profile for API operations. If not set, the default profile\n" +
			"created with `aws configure` will be used.",

		"shared_credentials_file": "The path to the shared credentials file. If not set\n" +
			"this defaults to ~/.aws/credentials.",

		"token": "session token. A session token is only required if you are\n" +
			"using temporary security credentials.",

		"max_retries": "The maximum number of times an AWS API request is\n" +
			"being executed. If the API request still fails, an error is\n" +
			"thrown.",

		"endpoint": "Use this to override the default service endpoint URL",

		"insecure": "Explicitly allow the provider to perform \"insecure\" SSL requests. If omitted," +
			"default value is `false`",

		"skip_credentials_validation": "Skip the credentials validation via STS API. " +
			"Used for AWS API implementations that do not have STS available/implemented.",

		"skip_get_ec2_platforms": "Skip getting the supported EC2 platforms. " +
			"Used by users that don't have ec2:DescribeAccountAttributes permissions.",

		"skip_region_validation": "Skip static validation of region name. " +
			"Used by users of alternative AWS-like APIs or users w/ access to regions that are not public (yet).",

		"skip_requesting_account_id": "Skip requesting the account ID. " +
			"Used for AWS API implementations that do not have IAM/STS API and/or metadata API.",

		"skip_medatadata_api_check": "Skip the AWS Metadata API check. " +
			"Used for AWS API implementations that do not have a metadata api endpoint.",

		"s3_force_path_style": "Set this to true to force the request to use path-style addressing,\n" +
			"i.e., http://s3.amazonaws.com/BUCKET/KEY. By default, the S3 client will\n" +
			"use virtual hosted bucket addressing when possible\n" +
			"(http://BUCKET.s3.amazonaws.com/KEY). Specific to the Amazon S3 service.",
	}

	endpointServiceNames = []string{
		"accessanalyzer",
		"acm",
		"acmpca",
		"amplify",
		"apigateway",
		"appconfig",
		"applicationautoscaling",
		"applicationinsights",
		"appmesh",
		"apprunner",
		"appstream",
		"appsync",
		"athena",
		"auditmanager",
		"autoscaling",
		"autoscalingplans",
		"backup",
		"batch",
		"budgets",
		"chime",
		"cloud9",
		"cloudformation",
		"cloudfront",
		"cloudhsm",
		"cloudsearch",
		"cloudtrail",
		"cloudwatch",
		"cloudwatchevents",
		"cloudwatchlogs",
		"codeartifact",
		"codebuild",
		"codecommit",
		"codedeploy",
		"codepipeline",
		"codestarconnections",
		"cognitoidentity",
		"cognitoidp",
		"configservice",
		"connect",
		"cur",
		"dataexchange",
		"datapipeline",
		"datasync",
		"dax",
		"detective",
		"devicefarm",
		"directconnect",
		"dlm",
		"dms",
		"docdb",
		"ds",
		"dynamodb",
		"ec2",
		"ecr",
		"ecrpublic",
		"ecs",
		"efs",
		"eks",
		"elasticache",
		"elasticbeanstalk",
		"elastictranscoder",
		"elb",
		"emr",
		"emrcontainers",
		"es",
		"firehose",
		"fms",
		"forecast",
		"fsx",
		"gamelift",
		"glacier",
		"globalaccelerator",
		"glue",
		"greengrass",
		"guardduty",
		"iam",
		"identitystore",
		"imagebuilder",
		"inspector",
		"iot",
		"iotanalytics",
		"iotevents",
		"kafka",
		"kinesis",
		"kinesisanalytics",
		"kinesisanalyticsv2",
		"kinesisvideo",
		"kms",
		"lakeformation",
		"lambda",
		"lexmodels",
		"licensemanager",
		"lightsail",
		"location",
		"macie",
		"macie2",
		"managedblockchain",
		"marketplacecatalog",
		"mediaconnect",
		"mediaconvert",
		"medialive",
		"mediapackage",
		"mediastore",
		"mediastoredata",
		"mq",
		"mwaa",
		"neptune",
		"networkfirewall",
		"networkmanager",
		"opsworks",
		"organizations",
		"outposts",
		"personalize",
		"pinpoint",
		"pricing",
		"qldb",
		"quicksight",
		"ram",
		"rds",
		"redshift",
		"resourcegroups",
		"resourcegroupstaggingapi",
		"route53",
		"route53domains",
		"route53resolver",
		"s3",
		"s3control",
		"s3outposts",
		"sagemaker",
		"schemas",
		"sdb",
		"secretsmanager",
		"securityhub",
		"serverlessrepo",
		"servicecatalog",
		"servicediscovery",
		"servicequotas",
		"ses",
		"shield",
		"signer",
		"sns",
		"sqs",
		"ssm",
		"ssoadmin",
		"stepfunctions",
		"storagegateway",
		"sts",
		"swf",
		"synthetics",
		"timestreamwrite",
		"transfer",
		"waf",
		"wafregional",
		"wafv2",
		"worklink",
		"workmail",
		"workspaces",
		"xray",
	}
}

func providerConfigure(ctx context.Context, d *schema.ResourceData, terraformVersion string) (interface{}, []diag.Diagnostic) {
	config := Config{
		AccessKey:               d.Get("access_key").(string),
		SecretKey:               d.Get("secret_key").(string),
		Profile:                 d.Get("profile").(string),
		Token:                   d.Get("token").(string),
		Region:                  d.Get("region").(string),
		CredsFilename:           d.Get("shared_credentials_file").(string),
		DefaultTagsConfig:       expandProviderDefaultTags(d.Get("default_tags").([]interface{})),
		Endpoints:               make(map[string]string),
		MaxRetries:              d.Get("max_retries").(int),
		IgnoreTagsConfig:        expandProviderIgnoreTags(d.Get("ignore_tags").([]interface{})),
		Insecure:                d.Get("insecure").(bool),
		SkipCredsValidation:     d.Get("skip_credentials_validation").(bool),
		SkipGetEC2Platforms:     d.Get("skip_get_ec2_platforms").(bool),
		SkipRegionValidation:    d.Get("skip_region_validation").(bool),
		SkipRequestingAccountId: d.Get("skip_requesting_account_id").(bool),
		SkipMetadataApiCheck:    d.Get("skip_metadata_api_check").(bool),
		S3ForcePathStyle:        d.Get("s3_force_path_style").(bool),
		terraformVersion:        terraformVersion,
	}

	if l, ok := d.Get("assume_role").([]interface{}); ok && len(l) > 0 && l[0] != nil {
		m := l[0].(map[string]interface{})

		if v, ok := m["duration_seconds"].(int); ok && v != 0 {
			config.AssumeRoleDurationSeconds = v
		}

		if v, ok := m["external_id"].(string); ok && v != "" {
			config.AssumeRoleExternalID = v
		}

		if v, ok := m["policy"].(string); ok && v != "" {
			config.AssumeRolePolicy = v
		}

		if policyARNSet, ok := m["policy_arns"].(*schema.Set); ok && policyARNSet.Len() > 0 {
			for _, policyARNRaw := range policyARNSet.List() {
				policyARN, ok := policyARNRaw.(string)

				if !ok {
					continue
				}

				config.AssumeRolePolicyARNs = append(config.AssumeRolePolicyARNs, policyARN)
			}
		}

		if v, ok := m["role_arn"].(string); ok && v != "" {
			config.AssumeRoleARN = v
		}

		if v, ok := m["session_name"].(string); ok && v != "" {
			config.AssumeRoleSessionName = v
		}

		if tagMapRaw, ok := m["tags"].(map[string]interface{}); ok && len(tagMapRaw) > 0 {
			config.AssumeRoleTags = make(map[string]string)

			for k, vRaw := range tagMapRaw {
				v, ok := vRaw.(string)

				if !ok {
					continue
				}

				config.AssumeRoleTags[k] = v
			}
		}

		if transitiveTagKeySet, ok := m["transitive_tag_keys"].(*schema.Set); ok && transitiveTagKeySet.Len() > 0 {
			for _, transitiveTagKeyRaw := range transitiveTagKeySet.List() {
				transitiveTagKey, ok := transitiveTagKeyRaw.(string)

				if !ok {
					continue
				}

				config.AssumeRoleTransitiveTagKeys = append(config.AssumeRoleTransitiveTagKeys, transitiveTagKey)
			}
		}

		log.Printf("[INFO] assume_role configuration set: (ARN: %q, SessionID: %q, ExternalID: %q)", config.AssumeRoleARN, config.AssumeRoleSessionName, config.AssumeRoleExternalID)
	}

	endpointsSet := d.Get("endpoints").(*schema.Set)

	for _, endpointsSetI := range endpointsSet.List() {
		endpoints := endpointsSetI.(map[string]interface{})
		for _, endpointServiceName := range endpointServiceNames {
			config.Endpoints[endpointServiceName] = endpoints[endpointServiceName].(string)
		}
	}

	if v, ok := d.GetOk("allowed_account_ids"); ok {
		for _, accountIDRaw := range v.(*schema.Set).List() {
			config.AllowedAccountIds = append(config.AllowedAccountIds, accountIDRaw.(string))
		}
	}

	if v, ok := d.GetOk("forbidden_account_ids"); ok {
		for _, accountIDRaw := range v.(*schema.Set).List() {
			config.ForbiddenAccountIds = append(config.ForbiddenAccountIds, accountIDRaw.(string))
		}
	}

	awsClient, err := config.Client()

	if err != nil {
		return nil, diag.FromErr(err)
	}

	return awsClient, nil
}

// This is a global MutexKV for use within this plugin.
var awsMutexKV = mutexkv.NewMutexKV()

func assumeRoleSchema() *schema.Schema {
	return &schema.Schema{
		Type:     schema.TypeList,
		Optional: true,
		MaxItems: 1,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"duration_seconds": {
					Type:        schema.TypeInt,
					Optional:    true,
					Description: "Seconds to restrict the assume role session duration.",
				},
				"external_id": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Unique identifier that might be required for assuming a role in another account.",
				},
				"policy": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "IAM Policy JSON describing further restricting permissions for the IAM Role being assumed.",
					ValidateFunc: validation.StringIsJSON,
				},
				"policy_arns": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Amazon Resource Names (ARNs) of IAM Policies describing further restricting permissions for the IAM Role being assumed.",
					Elem: &schema.Schema{
						Type:         schema.TypeString,
						ValidateFunc: validateArn,
					},
				},
				"role_arn": {
					Type:         schema.TypeString,
					Optional:     true,
					Description:  "Amazon Resource Name of an IAM Role to assume prior to making API calls.",
					ValidateFunc: validateArn,
				},
				"session_name": {
					Type:        schema.TypeString,
					Optional:    true,
					Description: "Identifier for the assumed role session.",
				},
				"tags": {
					Type:        schema.TypeMap,
					Optional:    true,
					Description: "Assume role session tags.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
				"transitive_tag_keys": {
					Type:        schema.TypeSet,
					Optional:    true,
					Description: "Assume role session tag keys to pass to any subsequent sessions.",
					Elem:        &schema.Schema{Type: schema.TypeString},
				},
			},
		},
	}
}

func endpointsSchema() *schema.Schema {
	endpointsAttributes := make(map[string]*schema.Schema)

	for _, endpointServiceName := range endpointServiceNames {
		endpointsAttributes[endpointServiceName] = &schema.Schema{
			Type:        schema.TypeString,
			Optional:    true,
			Default:     "",
			Description: descriptions["endpoint"],
		}
	}

	return &schema.Schema{
		Type:     schema.TypeSet,
		Optional: true,
		Elem: &schema.Resource{
			Schema: endpointsAttributes,
		},
	}
}

func expandProviderDefaultTags(l []interface{}) *keyvaluetags.DefaultConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	defaultConfig := &keyvaluetags.DefaultConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["tags"].(map[string]interface{}); ok {
		defaultConfig.Tags = keyvaluetags.New(v)
	}
	return defaultConfig
}

func expandProviderIgnoreTags(l []interface{}) *keyvaluetags.IgnoreConfig {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	ignoreConfig := &keyvaluetags.IgnoreConfig{}
	m := l[0].(map[string]interface{})

	if v, ok := m["keys"].(*schema.Set); ok {
		ignoreConfig.Keys = keyvaluetags.New(v.List())
	}

	if v, ok := m["key_prefixes"].(*schema.Set); ok {
		ignoreConfig.KeyPrefixes = keyvaluetags.New(v.List())
	}

	return ignoreConfig
}

// ReverseDns switches a DNS hostname to reverse DNS and vice-versa.
func ReverseDns(hostname string) string {
	parts := strings.Split(hostname, ".")

	for i, j := 0, len(parts)-1; i < j; i, j = i+1, j-1 {
		parts[i], parts[j] = parts[j], parts[i]
	}

	return strings.Join(parts, ".")
}
