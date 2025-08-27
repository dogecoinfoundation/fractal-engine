import * as cdk from "aws-cdk-lib";
import { Construct } from "constructs";

import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as iam from "aws-cdk-lib/aws-iam";
import * as s3assets from "aws-cdk-lib/aws-s3-assets";
import * as path from "path";

export interface DogecoinStackProps extends cdk.StackProps {
  // Network resources from NetworkStack
  vpc: ec2.IVpc;
  dogeSecurityGroup: ec2.ISecurityGroup;

  // Optional overrides
  subnetSelection?: ec2.SubnetSelection;
  instanceType?: ec2.InstanceType;
  volumeSizeGiB?: number;
  volumeDeviceName?: string;
}

/**
 * DogecoinStack
 * - Launches an EC2 instance for Dogecoin with an attached EBS volume
 * - Uses UserData to install Nix, build the Dogecoin wrapper, and run dogecoind as a systemd service
 * - Expects networking (VPC and security groups) to be created by NetworkStack
 */
export class DogecoinStack extends cdk.Stack {
  public readonly instance: ec2.Instance;

  constructor(scope: Construct, id: string, props: DogecoinStackProps) {
    super(scope, id, props);

    //
    // IAM Role for EC2 (SSM + CloudWatch agent)
    //
    const ec2Role = new iam.Role(this, "DogeEc2Role", {
      assumedBy: new iam.ServicePrincipal("ec2.amazonaws.com"),
      managedPolicies: [
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "AmazonSSMManagedInstanceCore",
        ),
        iam.ManagedPolicy.fromAwsManagedPolicyName(
          "CloudWatchAgentServerPolicy",
        ),
      ],
    });

    //
    // Assets: bootstrap script and Nix expressions
    //
    const bootstrapAsset = new s3assets.Asset(this, "DogeBootstrapAsset", {
      path: path.join(__dirname, "../assets/dogecoin/bootstrap.sh"),
    });
    const dogecoinNixAsset = new s3assets.Asset(this, "DogecoinNixAsset", {
      path: path.join(__dirname, "../../nix/dogecoin.nix"),
    });

    // Allow instance role to read assets from S3
    bootstrapAsset.grantRead(ec2Role);
    dogecoinNixAsset.grantRead(ec2Role);

    //
    // UserData: download assets and run bootstrap
    //
    const userData = ec2.UserData.forLinux();
    userData.addCommands(
      "set -euxo pipefail",
      "install -d -m 0755 /opt/dogecoin-nix",
    );

    userData.addS3DownloadCommand({
      bucket: bootstrapAsset.bucket,
      bucketKey: bootstrapAsset.s3ObjectKey,
      localFile: "/opt/dogecoin-nix/bootstrap.sh",
    });
    userData.addS3DownloadCommand({
      bucket: dogecoinNixAsset.bucket,
      bucketKey: dogecoinNixAsset.s3ObjectKey,
      localFile: "/opt/dogecoin-nix/dogecoin.nix",
    });

    userData.addCommands("chmod +x /opt/dogecoin-nix/bootstrap.sh");

    // Optional: you can export environment values for the bootstrap here
    // userData.addCommands('export CHAIN=mainnet', 'export RPC_USER=test', ...);

    userData.addExecuteFileCommand({
      filePath: "/opt/dogecoin-nix/bootstrap.sh",
    });

    //
    // EC2 instance with EBS volume for Dogecoin data
    //
    const deviceName = props.volumeDeviceName ?? "/dev/xvdb";
    const volumeSize = props.volumeSizeGiB ?? 200;

    this.instance = new ec2.Instance(this, "DogeNode", {
      vpc: props.vpc,
      vpcSubnets: props.subnetSelection ?? {
        subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
      },
      instanceType:
        props.instanceType ??
        ec2.InstanceType.of(ec2.InstanceClass.T3, ec2.InstanceSize.SMALL),
      machineImage: ec2.MachineImage.latestAmazonLinux2023({
        cachedInContext: true,
      }),
      securityGroup: props.dogeSecurityGroup,
      role: ec2Role,
      userData,
      blockDevices: [
        {
          deviceName,
          volume: ec2.BlockDeviceVolume.ebs(volumeSize, {
            encrypted: true,
            volumeType: ec2.EbsDeviceVolumeType.GP3,
          }),
        },
      ],
    });

    //
    // Outputs
    //
    new cdk.CfnOutput(this, "DogeInstanceId", {
      value: this.instance.instanceId,
    });
    new cdk.CfnOutput(this, "DogePrivateIp", {
      value: this.instance.instancePrivateIp,
    });
  }
}
