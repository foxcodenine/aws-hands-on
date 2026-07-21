# Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
# SPDX-License-Identifier: Apache-2.0
import json
import urllib.parse
import boto3

print('Loading function')

# S3 client used to read the uploaded object
s3 = boto3.client('s3')


def lambda_handler(event, context):
    #print("Received event: " + json.dumps(event, indent=2))

    # Step 1: read bucket name and object key out of the S3 event record
    bucket = event['Records'][0]['s3']['bucket']['name']

    # Step 2: decode the key (S3 URL-encodes spaces/special chars in event keys)
    key = urllib.parse.unquote_plus(event['Records'][0]['s3']['object']['key'], encoding='utf-8')
    try:
        # Step 3: fetch the object and read its content type
        response = s3.get_object(Bucket=bucket, Key=key)
        print("CONTENT TYPE: " + response['ContentType'])
        return response['ContentType']
    
    except Exception as e:
        # Step 4: log and re-raise so the Lambda invocation shows as failed
        print(e)
        print('Error getting object {} from bucket {}. Make sure they exist and your bucket is in the same region as this function.'.format(key, bucket))
        raise e

