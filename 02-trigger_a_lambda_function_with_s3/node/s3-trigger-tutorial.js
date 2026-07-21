import { S3Client, HeadObjectCommand } from "@aws-sdk/client-s3";

// S3 client used to read the uploaded object's metadata
const client = new S3Client();

export const handler = async (event, context) => {

    // Step 1: read bucket name and object key out of the S3 event record
    const bucket = event.Records[0].s3.bucket.name;

    // Step 2: decode the key (S3 URL-encodes spaces/special chars in event keys)
    const key = decodeURIComponent(event.Records[0].s3.object.key.replace(/\+/g, ' '));

    try {
        // Step 3: fetch the object's metadata (headers only, no body) and read its content type
        const { ContentType } = await client.send(new HeadObjectCommand({
            Bucket: bucket,
            Key: key,
        }));

        console.log('CONTENT TYPE:', ContentType);
        return ContentType;

    } catch (err) {

        // Step 4: log and re-throw so the Lambda invocation shows as failed
        console.log(err);
        const message = `Error getting object ${key} from bucket ${bucket}. Make sure they exist and your bucket is in the same region as this function.`;
        
        console.log(message);
        throw new Error(message);
    }
};

