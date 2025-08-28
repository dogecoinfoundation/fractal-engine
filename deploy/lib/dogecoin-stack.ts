import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as logs from "aws-cdk-lib/aws-logs";
import * as iam from "aws-cdk-lib/aws-iam";
import * as servicediscovery from "aws-cdk-lib/aws-servicediscovery";
import * as efs from "aws-cdk-lib/aws-efs";

export interface DogecoinStackProps extends cdk.StackProps {
  // Network resources from NetworkStack
  vpc: ec2.IVpc;
  dogeSecurityGroup: ec2.ISecurityGroup;
  namespace: cdk.aws_servicediscovery.PrivateDnsNamespace;

  // Optional networking
  subnetSelection?: ec2.SubnetSelection; // defaults to PRIVATE_WITH_EGRESS

  // ECS/Service configuration
  desiredCount?: number; // default 1
  cpu?: number; // default 512
  memoryMiB?: number; // default 1024

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
  public readonly serviceDiscoveryName: string;
  public readonly rpcPort: number;
  public readonly zmqPort: number;

  constructor(scope: Construct, id: string, props: DogecoinStackProps) {
    super(scope, id, props);

    const desiredCount = props.desiredCount ?? 1;
    const cpu = props.cpu ?? 512;
    const memoryMiB = props.memoryMiB ?? 1024;

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

    const subnets = props.vpc.selectSubnets({
      subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
    });

    // Storage: EFS for persistent Dogecoin data
    const dogeEfs = new efs.FileSystem(this, "DogecoinEfs", {
      vpc: props.vpc,
      vpcSubnets: subnets,
      lifecyclePolicy: efs.LifecyclePolicy.AFTER_14_DAYS,
      performanceMode: efs.PerformanceMode.GENERAL_PURPOSE,
      throughputMode: efs.ThroughputMode.BURSTING,
      removalPolicy: cdk.RemovalPolicy.RETAIN,
    });
    // Allow NFS from Dogecoin tasks
    dogeEfs.connections.allowDefaultPortFrom(
      props.dogeSecurityGroup,
      "Allow NFS from Dogecoin tasks",
    );
    // Access Point for the container
    const dogeEfsAp = new efs.AccessPoint(this, "DogecoinEfsAp", {
      fileSystem: dogeEfs,
      path: "/dogecoin",
      createAcl: { ownerUid: "0", ownerGid: "0", permissions: "0777" },
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
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "AmazonSSMManagedInstanceCore",
        ),
      ],
    });

    const taskRole = new iam.Role(this, "DogecoinTaskRole", {
      assumedBy: new iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
      description: "Task role for Dogecoin node",
    });

    // Minimal SSM messages permissions on the task role for ECS Exec
    taskRole.addToPolicy(
      new iam.PolicyStatement({
        actions: [
          "ssm:CreateControlChannel",
          "ssm:CreateDataChannel",
          "ssm:OpenControlChannel",
          "ssm:OpenDataChannel",
          "elasticfilesystem:ClientMount",
          "elasticfilesystem:ClientWrite",
          "elasticfilesystem:DescribeMountTargets",
          "elasticfilesystem:DescribeAccessPoints",
        ],
        resources: ["*"],
      }),
    );

    // Managed policy for operators to run ECS Exec against this service's tasks.
    // Attach this policy to your human/operator IAM users or roles.
    const ecsExecOperatorPolicy = new iam.ManagedPolicy(
      this,
      "EcsExecOperatorPolicy",
      {
        description: "Allows operators to run ECS Exec and manage SSM sessions",
        statements: [
          new iam.PolicyStatement({
            actions: ["ecs:ExecuteCommand"],
            resources: ["*"],
          }),
          new iam.PolicyStatement({
            actions: [
              "ssm:StartSession",
              "ssm:DescribeSessions",
              "ssm:TerminateSession",
            ],
            resources: ["*"],
          }),
          new iam.PolicyStatement({
            actions: ["kms:Decrypt"],
            resources: ["*"],
          }),
        ],
      },
    );

    //
    // Task Definition
    //
    const taskDef = new ecs.FargateTaskDefinition(this, "DogecoinTaskDef", {
      cpu,
      memoryLimitMiB: memoryMiB,
      executionRole,
      taskRole,
    });

    // Grant ECS task permission to mount the EFS access point
    executionRole.addToPolicy(
      new iam.PolicyStatement({
        actions: [
          "elasticfilesystem:ClientMount",
          "elasticfilesystem:ClientWrite",
          "elasticfilesystem:DescribeMountTargets",
          "elasticfilesystem:DescribeAccessPoints",
        ],
        resources: [dogeEfs.fileSystemArn, dogeEfsAp.accessPointArn],
      }),
    );

    // EFS volume for persistent blockchain data
    taskDef.addVolume({
      name: "dogecoin-data",
      efsVolumeConfiguration: {
        fileSystemId: dogeEfs.fileSystemId,
        transitEncryption: "ENABLED",
        authorizationConfig: {
          accessPointId: dogeEfsAp.accessPointId,
          iam: "ENABLED",
        },
      },
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
      cpu,
      memoryLimitMiB: memoryMiB,
      logging: ecs.LogDrivers.awsLogs({ streamPrefix: "dogecoin", logGroup }),
      // You can extend with additional env if your image supports it:
      // e.g., CHAIN, RPC_USER/PASSWORD, etc.
      environment: {
        ...props.environment,
        // Ensure dogecoind writes data to the EBS-backed host mount
        DATADIR: "/data",
      },
      essential: true,
    });

    // Mount EFS volume for persistent blockchain data
    container.addMountPoints({
      containerPath: "/data",
      sourceVolume: "dogecoin-data",
      readOnly: false,
    });
    // Expose the typical Dogecoin ports
    container.addPortMappings(
      { containerPort: rpcPort, protocol: ecs.Protocol.TCP }, // RPC
      { containerPort: p2pPort, protocol: ecs.Protocol.TCP }, // P2P
      { containerPort: zmqPort, protocol: ecs.Protocol.TCP }, // ZMQ
    );

    //
    // Fargate Service
    //
    this.service = new ecs.FargateService(this, "DogecoinService", {
      cluster: this.cluster,
      taskDefinition: taskDef,
      desiredCount,
      enableExecuteCommand: true,
      securityGroups: [props.dogeSecurityGroup],
      vpcSubnets: subnets,
      assignPublicIp: false,
      minHealthyPercent: 100,
      maxHealthyPercent: 200,
      cloudMapOptions: {
        name: serviceName,
        cloudMapNamespace: props.namespace,
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
