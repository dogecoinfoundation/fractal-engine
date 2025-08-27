import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as rds from "aws-cdk-lib/aws-rds";
import * as secretsmanager from "aws-cdk-lib/aws-secretsmanager";

export interface DatabaseStackProps extends cdk.StackProps {
  // Network
  vpc: ec2.IVpc;
  rdsSecurityGroup: ec2.ISecurityGroup;

  // Database options
  databaseName?: string;
  dbSubnetSelection?: ec2.SubnetSelection;
  instanceType?: ec2.InstanceType;
  multiAz?: boolean;
  allocatedStorageGiB?: number;
  maxAllocatedStorageGiB?: number;
  deletionProtection?: boolean;
  backupRetentionDays?: number;
  credentialsSecretName?: string;
}

/**
 * DatabaseStack
 * - Provisions an Amazon RDS PostgreSQL instance for the Fractal Engine
 * - Generates and stores credentials in Secrets Manager
 * - Exposes the instance and secret as public readonly properties and stack outputs
 */
export class DatabaseStack extends cdk.Stack {
  public readonly rdsInstance: rds.DatabaseInstance;
  public readonly rdsSecret: secretsmanager.ISecret;

  constructor(scope: Construct, id: string, props: DatabaseStackProps) {
    super(scope, id, props);

    const databaseName = props.databaseName ?? "fractal";

    // Credentials stored in Secrets Manager
    const dbCredentials = rds.Credentials.fromGeneratedSecret(
      "fractal_engine",
      {
        secretName:
          props.credentialsSecretName ?? "FractalEngineRdsCredentials",
        excludeCharacters: '"@/\\',
      },
    );

    this.rdsInstance = new rds.DatabaseInstance(this, "FractalDb", {
      vpc: props.vpc,
      vpcSubnets:
        props.dbSubnetSelection ?? { subnetType: ec2.SubnetType.PRIVATE_ISOLATED },
      securityGroups: [props.rdsSecurityGroup],

      engine: rds.DatabaseInstanceEngine.postgres({
        // Keep consistent with existing EngineStack defaults
        version: rds.PostgresEngineVersion.of("15", "15"),
      }),
      credentials: dbCredentials,
      databaseName,

      instanceType:
        props.instanceType ??
        ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.MICRO),
      multiAz: props.multiAz ?? false,
      allocatedStorage: props.allocatedStorageGiB ?? 20,
      maxAllocatedStorage: props.maxAllocatedStorageGiB ?? 100,
      storageType: rds.StorageType.GP3,

      publiclyAccessible: false,
      deletionProtection: props.deletionProtection ?? false,
      removalPolicy: cdk.RemovalPolicy.DESTROY, // NOTE: adjust for production
      cloudwatchLogsExports: ["postgresql"],
      backupRetention: cdk.Duration.days(props.backupRetentionDays ?? 3),
    });

    this.rdsSecret = this.rdsInstance.secret as secretsmanager.ISecret;

    // Outputs
    new cdk.CfnOutput(this, "RdsEndpoint", {
      value: this.rdsInstance.instanceEndpoint.socketAddress,
    });

    new cdk.CfnOutput(this, "RdsSecretName", {
      value: this.rdsSecret.secretName,
    });
  }
}
