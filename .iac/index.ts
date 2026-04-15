import * as cdk from 'aws-cdk-lib/core';
import { IacStack } from './lib/iac-stack';

const app = new cdk.App();

new IacStack(app, 'IacStack', {
  environment: 'prod',
  env: { region: 'eu-central-1' },
});

app.synth();