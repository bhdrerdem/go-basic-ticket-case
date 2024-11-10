import * as cdk from "aws-cdk-lib";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as ecsp from "aws-cdk-lib/aws-ecs-patterns";
import * as rds from "aws-cdk-lib/aws-rds";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as elasticache from "aws-cdk-lib/aws-elasticache";
import { Construct } from "constructs";
import { Config } from "../config/default";

export interface CustomStackProps extends cdk.StackProps, Config {}

export class InfraStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: CustomStackProps) {
    super(scope, id, props);

    const vpc = new ec2.Vpc(this, "GowitVpc", {
      maxAzs: 2,
      subnetConfiguration: [
        {
          cidrMask: 24,
          name: "gowitPublicSubnet",
          subnetType: ec2.SubnetType.PUBLIC,
        },
        {
          cidrMask: 24,
          name: "gowitPrivateSubnet",
          subnetType: ec2.SubnetType.PRIVATE_WITH_EGRESS,
        },
      ],
    });

    const rdsSecurityGroup = new ec2.SecurityGroup(this, "RdsSecurityGroup", {
      vpc,
      allowAllOutbound: true,
    });

    const postgres = new rds.DatabaseInstance(this, "PostgresqlInstance", {
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.VER_13,
      }),
      instanceType: new ec2.InstanceType("t3.micro"),
      allocatedStorage: 20,
      vpc,
      vpcSubnets: {
        subnetType: ec2.SubnetType.PUBLIC,
      },
      credentials: rds.Credentials.fromGeneratedSecret("postgres"),
      databaseName: "gowit",
      securityGroups: [rdsSecurityGroup],
      publiclyAccessible: true,
    });

    // Not recommended but for the sake of simplicity
    postgres.connections.allowFrom(
      ec2.Peer.anyIpv4(),
      ec2.Port.tcp(5432),
      "Allow server to PostgreSQL"
    );

    const redisSecurityGroup = new ec2.SecurityGroup(
      this,
      "RedisSecurityGroup",
      {
        vpc,
        allowAllOutbound: true,
      }
    );

    const redisSubnetGroup = new elasticache.CfnSubnetGroup(
      this,
      "RedisSubnetGroup",
      {
        description: "Subnet group for Redis",
        subnetIds: vpc.privateSubnets.map((subnet) => subnet.subnetId),
      }
    );

    const redis = new elasticache.CfnCacheCluster(this, "RedisCluster", {
      engine: "redis",
      cacheNodeType: "cache.t3.micro",
      numCacheNodes: 1,
      vpcSecurityGroupIds: [redisSecurityGroup.securityGroupId],
      cacheSubnetGroupName: redisSubnetGroup.ref,
    });

    const fargateService = new ecsp.ApplicationLoadBalancedFargateService(
      this,
      "GoWitCaseServer",
      {
        taskImageOptions: {
          image: ecs.ContainerImage.fromRegistry(props.imageUrl),
          containerPort: 8080,
          environment: {
            DB_HOST: postgres.dbInstanceEndpointAddress,
            DB_USER: "postgres",
            DB_PASSWORD: postgres
              .secret!.secretValueFromJson("password")
              .unsafeUnwrap(),
            DB_NAME: "gowit",
            DB_PORT: postgres.dbInstanceEndpointPort,
            REDIS_ADDR: redis.attrRedisEndpointAddress + ":6379",
            REDIS_PASS: "",
          },
        },
        publicLoadBalancer: true,
        desiredCount: 1,
        memoryLimitMiB: 512,
        cpu: 256,
        vpc,
      }
    );

    fargateService.targetGroup.configureHealthCheck({
      path: "/health",
    });

    redisSecurityGroup.addIngressRule(
      fargateService.service.connections.securityGroups[0],
      ec2.Port.tcp(6379),
      "Allow server to Redis"
    );
  }
}
