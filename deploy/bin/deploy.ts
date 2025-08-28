#!/usr/bin/env node
import * as cdk from "aws-cdk-lib";
import { NetworkStack } from "../lib/network-stack";
import { DogecoinStack } from "../lib/dogecoin-stack";
import { DatabaseStack } from "../lib/database-stack";
import { EngineStack } from "../lib/engine-stack";

const app = new cdk.App();
const env = {
  account: process.env.CDK_DEFAULT_ACCOUNT,
  region: process.env.CDK_DEFAULT_REGION,
};

const network = new NetworkStack(app, "NetworkStack", { env });

const doge = new DogecoinStack(app, "DogecoinStack", {
  vpc: network.vpc,
  dogeSecurityGroup: network.dogeSg,
  namespace: network.namespace,
  env,
});

const db = new DatabaseStack(app, "DatabaseStack", {
  vpc: network.vpc,
  rdsSecurityGroup: network.rdsSg,
  env,
});

const engine = new EngineStack(app, "EngineStack", {
  vpc: network.vpc,
  albSecurityGroup: network.albSg,
  engineSecurityGroup: network.engineSg,
  dbHost: db.rdsInstance.instanceEndpoint.hostname,
  dbPort: db.rdsInstance.instanceEndpoint.port,
  dbSecret: db.rdsSecret,
  dogecoin: {
    host: doge.serviceDiscoveryName,
    rpcPort: doge.rpcPort,
    zmqPort: doge.zmqPort,
  },
  env,
});

// Ensure Dogecoin deploys before the Engine
engine.addDependency(doge);
engine.addDependency(db);
