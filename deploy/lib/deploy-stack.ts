import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as elbv2 from "aws-cdk-lib/aws-elasticloadbalancingv2";
import * as iam from "aws-cdk-lib/aws-iam";
import * as rds from "aws-cdk-lib/aws-rds";
import * as logs from "aws-cdk-lib/aws-logs";
import * as secretsmanager from "aws-cdk-lib/aws-secretsmanager";

export class DeployStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    //
    // Networking: VPC with Public / Private App / Private Data subnets
    //
    const vpc = new ec2.Vpc(this, "FractalVpc", {
      ipAddresses: ec2.IpAddresses.cidr("10.0.0.0/16"),
      natGateways: 1,
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

    // VPC endpoints (recommended)
    vpc.addGatewayEndpoint("S3Endpoint", {
      service: ec2.GatewayVpcEndpointAwsService.S3,
      // route via both private subnets
      // (PRIVATE_WITH_EGRESS & PRIVATE_ISOLATED route tables)
      subnets: [
        { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
        { subnetType: ec2.SubnetType.PRIVATE_ISOLATED },
      ],
    });

    vpc.addInterfaceEndpoint("EcrDockerEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.ECR_DOCKER,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });
    vpc.addInterfaceEndpoint("EcrApiEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.ECR,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });
    vpc.addInterfaceEndpoint("LogsEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.CLOUDWATCH_LOGS,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });
    vpc.addInterfaceEndpoint("SecretsManagerEndpoint", {
      service: ec2.InterfaceVpcEndpointAwsService.SECRETS_MANAGER,
      subnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
    });

    //
    // Security Groups
    //
    const albSg = new ec2.SecurityGroup(this, "AlbSg", {
      vpc,
      description: "ALB security group",
      allowAllOutbound: true,
    });
    albSg.addIngressRule(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(80),
      "HTTP from anywhere",
    );
    // If/when TLS is used, open 443 as well (disabled until certs are configured):
    // albSg.addIngressRule(ec2.Peer.anyIpv4(), ec2.Port.tcp(443), 'HTTPS from anywhere');

    const engineSg = new ec2.SecurityGroup(this, "EngineSg", {
      vpc,
      description: "Fractal Engine security group",
      allowAllOutbound: true,
    });

    const rdsSg = new ec2.SecurityGroup(this, "RdsSg", {
      vpc,
      description: "RDS security group",
      allowAllOutbound: true,
    });

    const dogeSg = new ec2.SecurityGroup(this, "DogeSg", {
      vpc,
      description: "Dogecoin Node security group",
      allowAllOutbound: true,
    });

    // engine-sg inbound only from ALB
    engineSg.addIngressRule(albSg, ec2.Port.tcp(8891), "RPC from ALB only");

    // rds-sg inbound from engine-sg
    rdsSg.addIngressRule(
      engineSg,
      ec2.Port.tcp(5432),
      "Postgres from engine only",
    );

    // doge-sg inbound from engine-sg only
    dogeSg.addIngressRule(
      engineSg,
      ec2.Port.tcp(22555),
      "Dogecoin P2P from engine only",
    );

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

    const dbInstance = new rds.DatabaseInstance(this, "FractalDb", {
      vpc,
      vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_ISOLATED },
      securityGroups: [rdsSg],
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.of("15", "15"),
      }),
      credentials: dbCredentials,
      databaseName: "fractal",
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
      removalPolicy: cdk.RemovalPolicy.DESTROY, // TODO: Change this for prod
      cloudwatchLogsExports: ["postgresql"], // optional
      backupRetention: cdk.Duration.days(3),
    });

    // RDS secret for app usage
    const rdsSecret = dbInstance.secret as secretsmanager.ISecret;

    //
    // Compute: ECS Cluster + TaskDefinition + FargateService behind ALB
    //
    const cluster = new ecs.Cluster(this, "FractalCluster", {
      vpc,
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
    rdsSecret.grantRead(taskRole);

    const taskDef = new ecs.FargateTaskDefinition(this, "FractalTaskDef", {
      memoryLimitMiB: 1024,
      cpu: 512,
      executionRole: taskExecutionRole,
      taskRole,
    });

    const logGroup = new logs.LogGroup(this, "EngineLogs", {
      retention: logs.RetentionDays.ONE_WEEK,
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Note: Replace the container image below with the actual Fractal Engine image.
    const container = taskDef.addContainer("Engine", {
      image: ecs.ContainerImage.fromRegistry(
        "public.ecr.aws/docker/library/nginx:1.25",
      ),
      logging: ecs.LogDrivers.awsLogs({ streamPrefix: "engine", logGroup }),
      environment: {
        // Example environment variables for the engine
        RPC_SERVER_HOST: "0.0.0.0",
        RPC_SERVER_PORT: "8891",
        CORS_ALLOWED_ORIGINS: "*",
        // Database connection (host/port/name via env; credentials via secrets)
        DATABASE_HOST: dbInstance.instanceEndpoint.hostname,
        DATABASE_PORT: dbInstance.instanceEndpoint.port.toString(),
        DATABASE_NAME: "fractal",
      },
      secrets: {
        DATABASE_USERNAME: ecs.Secret.fromSecretsManager(rdsSecret, "username"),
        DATABASE_PASSWORD: ecs.Secret.fromSecretsManager(rdsSecret, "password"),
      },
      essential: true,
      // You might want to set command/entryPoint if your container image is generic
      // command: ['fractal-engine', '--database-url', '...'],
    });

    // Expose the engine's RPC port (8891)
    container.addPortMappings({
      containerPort: 8891,
      protocol: ecs.Protocol.TCP,
    });

    const service = new ecs.FargateService(this, "FractalService", {
      cluster,
      taskDefinition: taskDef,
      desiredCount: 1,
      securityGroups: [engineSg],
      vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
      assignPublicIp: false,
      minHealthyPercent: 100,
      maxHealthyPercent: 200,
    });

    // Application Load Balancer in public subnets
    const alb = new elbv2.ApplicationLoadBalancer(this, "FractalAlb", {
      vpc,
      internetFacing: true,
      securityGroup: albSg,
      vpcSubnets: { subnetType: ec2.SubnetType.PUBLIC },
    });

    const httpListener = alb.addListener("HttpListener", {
      port: 80,
      open: true,
    });

    // Target Group to ECS Service (IP mode for Fargate)
    const tg = new elbv2.ApplicationTargetGroup(this, "EngineTargetGroup", {
      vpc,
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
    service.attachToApplicationTargetGroup(tg);

    // Add TG to Listener
    httpListener.addTargetGroups("AttachEngineTg", {
      targetGroups: [tg],
    });

    //
    // Compute: Dogecoin L1 node on EC2 with EBS storage
    //
    const ec2Role = new iam.Role(this, "DogeEc2Role", {
      assumedBy: new iam.ServicePrincipal("ec2.amazonaws.com"),
      managedPolicies: [
        // SSM + CloudWatch agent policies
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "AmazonSSMManagedInstanceCore",
        ),
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "CloudWatchAgentServerPolicy",
        ),
      ],
    });

    // UserData to format/mount EBS and stub dogecoin service bootstrap
    const userData = ec2.UserData.forLinux();
    userData.addCommands(
      "set -euxo pipefail",
      "DEVICE_NAME=/dev/xvdb",
      "MOUNT_POINT=/var/lib/dogecoin",
      // Create FS if not already present
      "if ! file -s ${DEVICE_NAME} | grep -q filesystem; then mkfs -t xfs ${DEVICE_NAME}; fi",
      "mkdir -p ${MOUNT_POINT}",
      "blkid ${DEVICE_NAME} || true",
      'echo "${DEVICE_NAME} ${MOUNT_POINT} xfs defaults,nofail 0 2" >> /etc/fstab',
      "mount -a",
      // Placeholder for installing and starting dogecoin daemon
      'echo "Bootstrapping Dogecoin node (placeholder)..."',
      // e.g. install packages, docker, pull dogecoin image and run; or install native binary
      // 'dnf install -y docker && systemctl enable --now docker',
      // 'docker run -d --name dogecoind -v /var/lib/dogecoin:/data -p 22555:22555 -p 22556:22556 dogecoin/dogecoin:latest',
      'echo "Doge node started (placeholder)"',
    );

    const dogeInstance = new ec2.Instance(this, "DogeNode", {
      vpc,
      vpcSubnets: { subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS },
      instanceType: ec2.InstanceType.of(
        ec2.InstanceClass.T3,
        ec2.InstanceSize.SMALL,
      ),
      machineImage: ec2.MachineImage.latestAmazonLinux2023({
        cachedInContext: true,
      }),
      securityGroup: dogeSg,
      userData,
      role: ec2Role,
      blockDevices: [
        {
          deviceName: "/dev/xvdb",
          volume: ec2.BlockDeviceVolume.ebs(200, {
            encrypted: true,
            volumeType: ec2.EbsDeviceVolumeType.GP3,
          }),
        },
      ],
    });

    //
    // Permissions between Engine and Dogecoin node (already covered by SGs):
    // engineSg egress allowed, dogeSg ingress 22555 from engineSg
    //

    //
    // Outputs
    //
    new cdk.CfnOutput(this, "AlbDnsName", {
      value: alb.loadBalancerDnsName,
    });

    new cdk.CfnOutput(this, "RdsEndpoint", {
      value: dbInstance.instanceEndpoint.socketAddress,
    });

    new cdk.CfnOutput(this, "RdsSecretName", {
      value: rdsSecret.secretName,
    });

    new cdk.CfnOutput(this, "DogeInstanceId", {
      value: dogeInstance.instanceId,
    });

    new cdk.CfnOutput(this, "VpcId", { value: vpc.vpcId });
  }
}
