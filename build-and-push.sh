#!/bin/bash

# Exit on error
set -e

# Configuration
REGION=us-west-2
ACCOUNT_ID=$(aws sts get-caller-identity --query Account --output text)
ECR_REGISTRY=$ACCOUNT_ID.dkr.ecr.$REGION.amazonaws.com

echo "Building MapReduce services for linux/amd64..."

# Create buildx builder if it doesn't exist
docker buildx create --name multiarch --use 2>/dev/null || docker buildx use multiarch

# Build all images for linux/amd64
echo "Building splitter..."
docker buildx build --platform linux/amd64 -t mapreduce-splitter:latest --load ./splitter

echo "Building mapper..."
docker buildx build --platform linux/amd64 -t mapreduce-mapper:latest --load ./mapper

echo "Building reducer..."
docker buildx build --platform linux/amd64 -t mapreduce-reducer:latest --load ./reducer

echo "Logging into ECR..."
aws ecr get-login-password --region $REGION | docker login --username AWS --password-stdin $ECR_REGISTRY

# Create repositories if they don't exist
for repo in mapreduce-splitter mapreduce-mapper mapreduce-reducer; do
    aws ecr describe-repositories --repository-names $repo --region $REGION 2>/dev/null || \
    aws ecr create-repository --repository-name $repo --region $REGION
done

echo "Pushing images to ECR..."

# Tag and push
for service in splitter mapper reducer; do
    docker tag mapreduce-$service:latest $ECR_REGISTRY/mapreduce-$service:latest
    docker push $ECR_REGISTRY/mapreduce-$service:latest
    echo "Pushed mapreduce-$service"
done

echo "All images pushed successfully!"
echo "ECR URIs:"
echo "  Splitter: $ECR_REGISTRY/mapreduce-splitter:latest"
echo "  Mapper:   $ECR_REGISTRY/mapreduce-mapper:latest"
echo "  Reducer:  $ECR_REGISTRY/mapreduce-reducer:latest"