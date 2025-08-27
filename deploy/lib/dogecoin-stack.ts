import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as logs from "aws-cdk-lib/aws-logs";
import * as iam from "aws-cdk-lib/aws-iam";
import * as servicediscovery from "aws-cdk-lib/aws-servicediscovery";

export interface DogecoinStackProps extends cdk.StackProps {
  // Network resources from NetworkStack
  vpc: ec2.IVpc;
  dogeSecurityGroup: ec2.ISecurityGroup;

  // Optional networking
  subnetSelection?: ec2.SubnetSelection; // defaults to PRIVATE_WITH_EGRESS

  // ECS/Service configuration
  desiredCount?: number; // default 1
  cpu?: number; // default 512
  memoryMiB?: number; // default 1024
  ephemeralStorageGiB?: number; // default 50

  // Container image override (defaults to docker.io/danielwhelansb/dogecoin)
  containerImage?: ecs.ContainerImage;

  // Service discovery
  namespaceName?: string; // defaults to "fractal.local"
  serviceName?: string; // defaults to "dogecoin"

  // Dogecoin ports
  rpcPort?: number; // default 22555
  p2pPort?: number; // default 22556
  zmqPort?: number; // default 28000

  // Extra environment for the container
  environment?: Record<string, string>;
}

/**
 * DogecoinStack (ECS Fargate)
 * - Runs Dogecoin in ECS Fargate
 * - Registers the service in AWS Cloud Map (private DNS) for discovery by Engine
 * - Publishes container logs to CloudWatch Logs
 */
export class DogecoinStack extends cdk.Stack {
  public readonly cluster: ecs.Cluster;
  public readonly service: ecs.FargateService;
  public readonly namespace: servicediscovery.PrivateDnsNamespace;
  public readonly serviceDiscoveryName: string;
  public readonly rpcPort: number;
  public readonly zmqPort: number;

  constructor(scope: Construct, id: string, props: DogecoinStackProps) {
    super(scope, id, props);

    const desiredCount = props.desiredCount ?? 1;
    const cpu = props.cpu ?? 512;
    const memoryMiB = props.memoryMiB ?? 1024;
    const ephemeralStorageGiB = props.ephemeralStorageGiB ?? 50;

    const rpcPort = props.rpcPort ?? 22555;
    const p2pPort = props.p2pPort ?? 22556;
    const zmqPort = props.zmqPort ?? 28000;
    this.rpcPort = rpcPort;
    this.zmqPort = zmqPort;

    const namespaceName = props.namespaceName ?? "fractal.local";
    const serviceName = props.serviceName ?? "dogecoin";
    const vpcCidr = props.vpc.vpcCidrBlock;

    //
    // ECS Cluster
    //
    this.cluster = new ecs.Cluster(this, "DogecoinCluster", {
      vpc: props.vpc,
      containerInsights: true,
    });

    //
    // IAM roles
    //
    const executionRole = new iam.Role(this, "DogecoinTaskExecutionRole", {
      assumedBy: new iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "service-role/AmazonECSTaskExecutionRolePolicy",
        ),
      ],
    });

    const taskRole = new iam.Role(this, "DogecoinTaskRole", {
      assumedBy: new iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
      description: "Task role for Dogecoin node",
    });

    //
    // Task Definition
    //
    const taskDef = new ecs.FargateTaskDefinition(this, "DogecoinTaskDef", {
      cpu,
      memoryLimitMiB: memoryMiB,
      executionRole,
      taskRole,
      ephemeralStorageGiB,
    });

    const logGroup = new logs.LogGroup(this, "DogecoinLogs", {
      retention: logs.RetentionDays.ONE_WEEK,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    //
    // Container
    //
    const image =
      props.containerImage ??
      ecs.ContainerImage.fromRegistry(
        "docker.io/danielwhelansb/dogecoin:v1.14.9",
      );

    const container = taskDef.addContainer("Dogecoin", {
      image,
      logging: ecs.LogDrivers.awsLogs({ streamPrefix: "dogecoin", logGroup }),
      // You can extend with additional env if your image supports it:
      // e.g., CHAIN, RPC_USER/PASSWORD, etc.
      environment: {
        ...props.environment,
      },
      // Ensure RPC listens on the VPC and accepts connections from the VPC CIDR
      command: [
        "dogecoind",
        "-server=1",
        "-printtoconsole",
        `-rpcbind=0.0.0.0`,
        `-rpcallowip=${vpcCidr}`,
        `-rpcuser=test`,
        `-rpcpassword=test`,
        `-rpcport=${rpcPort}`,
        "-listen=1",
        `-port=${p2pPort}`,
        "-txindex=1",
        `-zmqpubrawblock=tcp://0.0.0.0:${zmqPort}`,
        `-zmqpubrawtx=tcp://0.0.0.0:${zmqPort}`,
        `-zmqpubhashtx=tcp://0.0.0.0:${zmqPort}`,
        `-zmqpubhashblock=tcp://0.0.0.0:${zmqPort}`,
      ],
      essential: true,
    });

    // Expose the typical Dogecoin ports
    container.addPortMappings(
      { containerPort: rpcPort, protocol: ecs.Protocol.TCP }, // RPC
      { containerPort: p2pPort, protocol: ecs.Protocol.TCP }, // P2P
      { containerPort: zmqPort, protocol: ecs.Protocol.TCP }, // ZMQ
    );

    //
    // Service Discovery namespace (Private DNS in the VPC)
    //
    this.namespace = new servicediscovery.PrivateDnsNamespace(
      this,
      "DogecoinNamespace",
      {
        name: namespaceName,
        vpc: props.vpc,
      },
    );

    //
    // Fargate Service
    //
    this.service = new ecs.FargateService(this, "DogecoinService", {
      cluster: this.cluster,
      taskDefinition: taskDef,
      desiredCount,
      securityGroups: [props.dogeSecurityGroup],
      vpcSubnets: props.subnetSelection ?? {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      assignPublicIp: false,
      minHealthyPercent: 100,
      maxHealthyPercent: 200,
      cloudMapOptions: {
        name: serviceName,
        cloudMapNamespace: this.namespace,
        dnsRecordType: servicediscovery.DnsRecordType.A,
        dnsTtl: cdk.Duration.seconds(30),
      },
    });

    // Full service discovery DNS name host: service.namespace
    this.serviceDiscoveryName = `${serviceName}.${namespaceName}`;

    //
    // Outputs
    //
    new cdk.CfnOutput(this, "DogecoinServiceDiscoveryName", {
      value: this.serviceDiscoveryName,
    });
    new cdk.CfnOutput(this, "DogecoinRpcPort", {
      value: String(rpcPort),
    });
    new cdk.CfnOutput(this, "DogecoinZmqPort", {
      value: String(zmqPort),
    });
  }
}
