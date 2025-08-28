import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as servicediscovery from "aws-cdk-lib/aws-servicediscovery";

/**
 * NetworkStack sets up shared networking primitives:
 * - VPC with public, private-with-egress (app), and private-isolated (data) subnets
 * - Useful VPC endpoints for private networking
 * - Security groups for ALB, Engine, RDS, and Dogecoin EC2
 *
 * Deploy this stack first. Subsequent stacks (Dogecoin, Engine) should accept
 * references to the VPC and security groups from this stack.
 */
export interface NetworkStackProps extends cdk.StackProps {
  /**
   * CIDR block for the VPC. Defaults to 10.0.0.0/16.
   */
  cidr?: string;

  /**
   * Number of NAT Gateways. Defaults to 1.
   */
  natGateways?: number;
}

export class NetworkStack extends cdk.Stack {
  // Shared network resources to be consumed by other stacks
  public readonly vpc: ec2.Vpc;
  public readonly namespace: cdk.aws_servicediscovery.PrivateDnsNamespace;

  public readonly albSg: ec2.SecurityGroup;
  public readonly engineSg: ec2.SecurityGroup;
  public readonly rdsSg: ec2.SecurityGroup;
  public readonly dogeSg: ec2.SecurityGroup;

  constructor(scope: Construct, id: string, props?: NetworkStackProps) {
    super(scope, id, props);

    //
    // VPC
    //
    this.vpc = new ec2.Vpc(this, "FractalVpc", {
      ipAddresses: ec2.IpAddresses.cidr(props?.cidr ?? "10.0.0.0/16"),
      natGateways: props?.natGateways ?? 1,
      subnetConfiguration: [
        {
          name: "public",
          subnetType: ec2.SubnetType.PUBLIC,
          cidrMask: 24,
        },
        {
          name: "app",
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
          cidrMask: 24,
        },
        {
          name: "data",
          subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
          cidrMask: 24,
        },
      ],
    });

    this.namespace = new servicediscovery.PrivateDnsNamespace(
      this,
      "ServiceNamespace",
      {
        name: "fractal.local",
        vpc: this.vpc,
        description: "Private DNS for Fractal services",
      },
    );

    //
    // VPC Endpoints (recommended for private networking)
    //
    this.vpc.addGatewayEndpoint("S3Endpoint", {
      service: ec2.GatewayVpcEndpointAwsService.S3,
      subnets: [
        { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
        { subnetType: ec2.SubnetType.PRIVATE_ISOLATED },
      ],
    });

    this.vpc.addInterfaceEndpoint("EcrDockerEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.ECR_DOCKER,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    this.vpc.addInterfaceEndpoint("EcrApiEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.ECR,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    this.vpc.addInterfaceEndpoint("LogsEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.CLOUDWATCH_LOGS,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    this.vpc.addInterfaceEndpoint("SecretsManagerEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    this.vpc.addInterfaceEndpoint("SsmEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.SSM,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });
    this.vpc.addInterfaceEndpoint("SsmMessagesEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.SSM_MESSAGES,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });
    this.vpc.addInterfaceEndpoint("Ec2MessagesEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.EC2_MESSAGES,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    //
    // Security Groups
    //
    this.albSg = new ec2.SecurityGroup(this, "AlbSg", {
      vpc: this.vpc,
      description: "ALB security group",
      allowAllOutbound: true,
    });
    this.albSg.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(80),
      "HTTP from anywhere",
    );
    // If/when TLS is used, open 443 as well (uncomment below once certs are configured)
    // this.albSg.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(443), "HTTPS from anywhere");

    this.engineSg = new ec2.SecurityGroup(this, "EngineSg", {
      vpc: this.vpc,
      description: "Fractal Engine security group",
      allowAllOutbound: true,
    });

    this.rdsSg = new ec2.SecurityGroup(this, "RdsSg", {
      vpc: this.vpc,
      description: "RDS security group",
      allowAllOutbound: true,
    });

    this.dogeSg = new ec2.SecurityGroup(this, "DogeSg", {
      vpc: this.vpc,
      description: "Dogecoin Node security group",
      allowAllOutbound: true,
    });

    // Engine inbound from ALB (Engine RPC)
    this.engineSg.addIngressRule(
      this.albSg,
      ec2.Port.tcp(8891),
      "Engine RPC from ALB only",
    );

    // RDS inbound from Engine
    this.rdsSg.addIngressRule(
      this.engineSg,
      ec2.Port.tcp(5432),
      "Postgres from Engine only",
    );

    // Dogecoin inbound from Engine
    this.dogeSg.addIngressRule(
      this.engineSg,
      ec2.Port.tcp(22555),
      "Dogecoin RPC from Engine only",
    );
    this.dogeSg.addIngressRule(
      this.engineSg,
      ec2.Port.tcp(22556),
      "Dogecoin P2P from Engine only",
    );
    // Allow ZMQ from Engine
    this.dogeSg.addIngressRule(
      this.engineSg,
      ec2.Port.tcp(28000),
      "Dogecoin ZMQ from Engine only",
    );

    //
    // Outputs
    //
    new cdk.CfnOutput(this, "VpcId", { value: this.vpc.vpcId });
    new cdk.CfnOutput(this, "AlbSgId", { value: this.albSg.securityGroupId });
    new cdk.CfnOutput(this, "EngineSgId", {
      value: this.engineSg.securityGroupId,
    });
    new cdk.CfnOutput(this, "RdsSgId", { value: this.rdsSg.securityGroupId });
    new cdk.CfnOutput(this, "DogeSgId", { value: this.dogeSg.securityGroupId });
    new cdk.CfnOutput(this, "NamespaceId", {
      value: this.namespace.namespaceId,
    });
  }
}
