import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";
import * as iam from "aws-cdk-lib/aws-iam";
import * as rds from "aws-cdk-lib/aws-rds";
import * as logs from "aws-cdk-lib/aws-logs";
import * as secretsmanager from "aws-cdk-lib/aws-secretsmanager";

export interface DogecoinConnection {
  host: string;
  rpcPort?: number; // default 22555
  zmqPort?: number; // default 28000
}

export interface EngineStackProps extends cdk.StackProps {
  // From NetworkStack
  vpc: ec2.IVpc;
  albSecurityGroup: ec2.ISecurityGroup;
  engineSecurityGroup: ec2.ISecurityGroup;
  rdsSecurityGroup: ec2.ISecurityGroup;

  // Optionally pass Dogecoin connection details (from DogecoinStack)
  dogecoin?: DogecoinConnection;

  // ECS/Service configuration
  desiredCount?: number;
  cpu?: number; // 256, 512, 1024...
  memoryMiB?: number; // 512, 1024, 2048...

  // Container image override
  engineContainerImage?: ecs.ContainerImage;

  // Database configuration
  databaseName?: string;

  // Subnets
  appSubnetSelection?: ec2.SubnetSelection; // defaults to PRIVATE_WITH_EGRESS
  albSubnetSelection?: ec2.SubnetSelection; // defaults to PUBLIC
  dbSubnetSelection?: ec2.SubnetSelection; // defaults to PRIVATE_ISOLATED
}

/**
 * EngineStack
 * - Provisions RDS Postgres for the Fractal Engine
 * - Creates ECS Fargate service for the engine, fronted by an ALB
 * - Accepts references to VPC and security groups from NetworkStack
 * - Optionally accepts Dogecoin connection details (from DogecoinStack) to set env vars
 */
export class EngineStack extends cdk.Stack {
  public readonly cluster: ecs.Cluster;
  public readonly service: ecs.FargateService;
  public readonly loadBalancer: elbv2.ApplicationLoadBalancer;
  public readonly rdsInstance: rds.DatabaseInstance;
  public readonly rdsSecret: secretsmanager.ISecret;

  constructor(scope: Construct, id: string, props: EngineStackProps) {
    super(scope, id, props);

    const databaseName = props.databaseName ?? "fractal";

    //
    // Data layer: RDS PostgreSQL
    //
    const dbCredentials = rds.Credentials.fromGeneratedSecret(
      "fractal_engine",
      {
        secretName: "FractalEngineRdsCredentials",
        excludeCharacters: '"@/\\',
      },
    );

    this.rdsInstance = new rds.DatabaseInstance(this, "FractalDb", {
      vpc: props.vpc,
      vpcSubnets: props.dbSubnetSelection ?? {
        subnetType: ec2.SubnetType.PRIVATE_ISOLATED,
      },
      securityGroups: [props.rdsSecurityGroup],
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.of("15", "15"),
      }),
      credentials: dbCredentials,
      databaseName,
      instanceType: ec2.InstanceType.of(
        ec2.InstanceClass.T3,
        ec2.InstanceSize.MICRO,
      ),
      multiAz: false,
      allocatedStorage: 20,
      maxAllocatedStorage: 100,
      storageType: rds.StorageType.GP3,
      publiclyAccessible: false,
      deletionProtection: false,
      removalPolicy: cdk.RemovalPolicy.DESTROY, // TODO: Change for production usage
      cloudwatchLogsExports: ["postgresql"],
      backupRetention: cdk.Duration.days(3),
    });

    this.rdsSecret = this.rdsInstance.secret as secretsmanager.ISecret;

    //
    // Compute: ECS Cluster + TaskDefinition + FargateService behind ALB
    //
    this.cluster = new ecs.Cluster(this, "FractalCluster", {
      vpc: props.vpc,
      containerInsights: true,
    });

    // IAM Roles
    const taskExecutionRole = new iam.Role(this, "TaskExecutionRole", {
      assumedBy: new iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "service-role/AmazonECSTaskExecutionRolePolicy",
        ),
      ],
    });

    const taskRole = new iam.Role(this, "TaskRole", {
      assumedBy: new iam.ServicePrincipal("ecs-tasks.amazonaws.com"),
      description:
        "Task role for Fractal Engine (access to Secrets Manager, etc.)",
    });

    // Allow the task to read the RDS generated secret
    this.rdsSecret.grantRead(taskRole);

    const taskDef = new ecs.FargateTaskDefinition(this, "FractalTaskDef", {
      memoryLimitMiB: props.memoryMiB ?? 1024,
      cpu: props.cpu ?? 512,
      executionRole: taskExecutionRole,
      taskRole,
    });

    const logGroup = new logs.LogGroup(this, "EngineLogs", {
      retention: logs.RetentionDays.ONE_WEEK,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    const image =
      props.engineContainerImage ??
      ecs.ContainerImage.fromRegistry(
        "ghcr.io/dogecoinfoundation/fractal-engine:v0.0.1",
      );

    const container = taskDef.addContainer("Engine", {
      image,
      logging: ecs.LogDrivers.awsLogs({ streamPrefix: "engine", logGroup }),
      environment: {
        RPC_SERVER_HOST: "0.0.0.0",
        RPC_SERVER_PORT: "8891",
        CORS_ALLOWED_ORIGINS: "*",
        DATABASE_HOST: this.rdsInstance.instanceEndpoint.hostname,
        DATABASE_PORT: this.rdsInstance.instanceEndpoint.port.toString(),
        DATABASE_NAME: databaseName,
        ...(props.dogecoin?.host
          ? {
              DOGECOIN_RPC_HOST: props.dogecoin.host,
              DOGECOIN_RPC_PORT: String(props.dogecoin.rpcPort ?? 22555),
              DOGECOIN_ZMQ_PORT: String(props.dogecoin.zmqPort ?? 28000),
            }
          : {}),
      },
      secrets: {
        DATABASE_USERNAME: ecs.Secret.fromSecretsManager(
          this.rdsSecret,
          "username",
        ),
        DATABASE_PASSWORD: ecs.Secret.fromSecretsManager(
          this.rdsSecret,
          "password",
        ),
      },
      essential: true,
    });

    container.addPortMappings({
      containerPort: 8891,
      protocol: ecs.Protocol.TCP,
    });

    this.service = new ecs.FargateService(this, "FractalService", {
      cluster: this.cluster,
      taskDefinition: taskDef,
      desiredCount: props.desiredCount ?? 1,
      securityGroups: [props.engineSecurityGroup],
      vpcSubnets: props.appSubnetSelection ?? {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      assignPublicIp: false,
      minHealthyPercent: 100,
      maxHealthyPercent: 200,
    });

    // Application Load Balancer in public subnets
    this.loadBalancer = new elbv2.ApplicationLoadBalancer(this, "FractalAlb", {
      vpc: props.vpc,
      internetFacing: true,
      securityGroup: props.albSecurityGroup,
      vpcSubnets: props.albSubnetSelection ?? {
        subnetType: ec2.SubnetType.PUBLIC,
      },
    });

    const httpListener = this.loadBalancer.addListener("HttpListener", {
      port: 80,
      open: true,
    });

    const tg = new elbv2.ApplicationTargetGroup(this, "EngineTargetGroup", {
      vpc: props.vpc,
      port: 8891,
      protocol: elbv2.ApplicationProtocol.HTTP,
      targetType: elbv2.TargetType.IP,
      healthCheck: {
        path: "/health",
        healthyHttpCodes: "200",
        interval: cdk.Duration.seconds(30),
      },
      deregistrationDelay: cdk.Duration.seconds(10),
    });

    // Attach the service to the Target Group
    this.service.attachToApplicationTargetGroup(tg);

    // Add TG to Listener
    httpListener.addTargetGroups("AttachEngineTg", {
      targetGroups: [tg],
    });

    //
    // Outputs
    //
    new cdk.CfnOutput(this, "AlbDnsName", {
      value: this.loadBalancer.loadBalancerDnsName,
    });

    new cdk.CfnOutput(this, "RdsEndpoint", {
      value: this.rdsInstance.instanceEndpoint.socketAddress,
    });

    new cdk.CfnOutput(this, "RdsSecretName", {
      value: this.rdsSecret.secretName,
    });
  }
}
