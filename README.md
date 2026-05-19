# not-just-noise

## Summary

Provide an audio file link via an API. Download and store the file in AWS S3, then use AWS Bedrock to generate a summary including:

- General Summary
- Transcription
- Categorization
- Sentiment

Index the results in AWS OpenSearch to make the file searchable using the above information.

## Scope

This project is meant to show the following principles in action:

- Microservice Architecture
- AWS Platform Integration
- Serverless Computing
- Event-Driven Architecture
- Testability

## Project Components

The expected project structure is outlined in this diagram:
![not-just-noise Project Structure](https://github.com/jtenhave/not-just-noise/blob/main/not-just-noise.jpg)

This project is under construction. I will update this section as components are completed.

### ✅ RDS

RDS running MySQL 8.4 has been deployed. 

- Uses a public IP with a security group to lock it down, for now. May look to make this private within the VPC when services are deployed.
- Each service will have it's own dedicated user with the required permissions, e.g. `GRANT SELECT, INSERT, UPDATE ON audio_service.* TO 'audio_service_user'@'%';`
- Does not have any read-replicas and never will, since those will add cost, and this is just a portfolio project.

### 🚧 Audio Service

Audio service is under construction. This service will provide the basic CRUD operations for an audio resource, storing records in the RDS database.

### 🔜 Up Next

Set up the Index Service and integrate it into the Audio Service. This will allow records to exist in AWS OpenSearch, in an invisible state, until processing by AWS Bedrock is complete.

## Out of Scope

These are potentially coming to a future project soon:

- API Authentication
- Traceability & Logging

