import * as cdk from "aws-cdk-lib/core";
import * as dynamodb from "aws-cdk-lib/aws-dynamodb";
import { Construct } from "constructs";
import { aws_iam } from "aws-cdk-lib";

interface IacStackProps extends cdk.StackProps {
  environment: string;
}

export class IacStack extends cdk.Stack {
  constructor(scope: Construct, id: string, props: IacStackProps) {
    super(scope, id, props);

    const entitiesTable = new dynamodb.Table(this, "EntitiesTable", {
      tableName: `barricade-entities-${props.environment}`,
      partitionKey: { name: "id", type: dynamodb.AttributeType.STRING },
      sortKey: { name: "type", type: dynamodb.AttributeType.STRING },
      billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
    });

    entitiesTable.addGlobalSecondaryIndex({
      indexName: "secondary-lookup-index",
      partitionKey: {
        name: "secondary-lookup",
        type: dynamodb.AttributeType.STRING,
      },
      sortKey: {
        name: "secondary-lookup-sk",
        type: dynamodb.AttributeType.STRING,
      },
    });

    const operationalTable = new dynamodb.Table(
      this,
      "OperationalEntitiesTable",
      {
        tableName: `barricade-operational-entities-${props.environment}`,
        partitionKey: { name: "id", type: dynamodb.AttributeType.STRING },
        sortKey: { name: "type", type: dynamodb.AttributeType.STRING },
        timeToLiveAttribute: "expireAt",
        billingMode: dynamodb.BillingMode.PAY_PER_REQUEST,
      },
    );

    operationalTable.addGlobalSecondaryIndex({
      indexName: "secondary-lookup-index",
      partitionKey: {
        name: "secondary-lookup",
        type: dynamodb.AttributeType.STRING,
      },
      sortKey: {
        name: "secondary-lookup-sk",
        type: dynamodb.AttributeType.STRING,
      },
    });

    const user = aws_iam.User.fromUserArn(
      this,
      "barricadeAccess",
      "arn:aws:iam::823070154118:user/barricade-stage",
    );

    entitiesTable.grantReadWriteData(user);
    operationalTable.grantReadWriteData(user);
  }
}
