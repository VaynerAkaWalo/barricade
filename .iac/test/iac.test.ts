import * as cdk from "aws-cdk-lib/core";
import { Template, Match } from "aws-cdk-lib/assertions";
import { IacStack } from "../lib/iac-stack";

function createStack() {
  const app = new cdk.App();
  return new IacStack(app, "TestStack", {
    environment: "prod",
    env: { region: "eu-central-1" },
  });
}

describe("IacStack", () => {
  describe("EntitiesTable", () => {
    let template: Template;

    beforeEach(() => {
      template = Template.fromStack(createStack());
    });

    it("creates the entities DynamoDB table", () => {
      template.resourceCountIs("AWS::DynamoDB::Table", 2);
    });

    it("has the correct table name", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        TableName: "barricade-entities-prod",
      });
    });

    it("uses string partition key named id", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        KeySchema: Match.arrayWith([
          { AttributeName: "id", KeyType: "HASH" },
        ]),
        AttributeDefinitions: Match.arrayWith([
          { AttributeName: "id", AttributeType: "S" },
        ]),
      });
    });

    it("uses string sort key named type", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        KeySchema: Match.arrayWith([
          { AttributeName: "type", KeyType: "RANGE" },
        ]),
        AttributeDefinitions: Match.arrayWith([
          { AttributeName: "type", AttributeType: "S" },
        ]),
      });
    });

    it("uses pay per request billing mode", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        BillingMode: "PAY_PER_REQUEST",
      });
    });

    it("has the secondary lookup GSI", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        GlobalSecondaryIndexes: Match.arrayWith([
          Match.objectLike({
            IndexName: "secondary-lookup-index",
            KeySchema: [
              { AttributeName: "secondary-lookup", KeyType: "HASH" },
              { AttributeName: "secondary-lookup-sk", KeyType: "RANGE" },
            ],
            Projection: { ProjectionType: "ALL" },
          }),
        ]),
      });
    });
  });

  describe("OperationalEntitiesTable", () => {
    let template: Template;

    beforeEach(() => {
      template = Template.fromStack(createStack());
    });

    it("has the correct table name", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        TableName: "barricade-operational-entities-prod",
      });
    });

    it("uses string partition key named id", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        KeySchema: Match.arrayWith([
          { AttributeName: "id", KeyType: "HASH" },
        ]),
      });
    });

    it("uses string sort key named type", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        KeySchema: Match.arrayWith([
          { AttributeName: "type", KeyType: "RANGE" },
        ]),
      });
    });

    it("has TTL enabled on expireAt attribute", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        TimeToLiveSpecification: {
          AttributeName: "expireAt",
          Enabled: true,
        },
      });
    });

    it("uses pay per request billing mode", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        BillingMode: "PAY_PER_REQUEST",
      });
    });

    it("has the secondary lookup GSI", () => {
      template.hasResourceProperties("AWS::DynamoDB::Table", {
        GlobalSecondaryIndexes: Match.arrayWith([
          Match.objectLike({
            IndexName: "secondary-lookup-index",
            KeySchema: [
              { AttributeName: "secondary-lookup", KeyType: "HASH" },
              { AttributeName: "secondary-lookup-sk", KeyType: "RANGE" },
            ],
            Projection: { ProjectionType: "ALL" },
          }),
        ]),
      });
    });
  });

  describe("IAM Permissions", () => {
    let template: Template;

    beforeEach(() => {
      template = Template.fromStack(createStack());
    });

    it("creates an IAM policy granting read write access to both tables", () => {
      template.hasResourceProperties("AWS::IAM::Policy", {
        PolicyDocument: {
          Version: "2012-10-17",
          Statement: Match.arrayWith([
            Match.objectLike({
              Effect: "Allow",
            }),
          ]),
        },
      });

      const policy = template.findResources("AWS::IAM::Policy");
      const policyName = Object.keys(policy)[0];
      const statements = (policy[policyName].Properties as any).PolicyDocument.Statement as Array<any>;
      const tableActions = statements.find((s: any) => Array.isArray(s.Action) && s.Action.includes("dynamodb:PutItem"));
      expect(tableActions).toBeDefined();
      expect(tableActions!.Action).toContain("dynamodb:PutItem");
      expect(tableActions!.Action).toContain("dynamodb:GetItem");
      expect(tableActions!.Action).toContain("dynamodb:DeleteItem");
      expect(tableActions!.Action).toContain("dynamodb:UpdateItem");
      expect(tableActions!.Action).toContain("dynamodb:Query");
      expect(tableActions!.Action).toContain("dynamodb:Scan");
    });

    it("attaches the policy to the correct IAM user", () => {
      template.hasResourceProperties("AWS::IAM::Policy", {
        Users: Match.arrayWith(["barricade-stage"]),
      });
    });
  });
});