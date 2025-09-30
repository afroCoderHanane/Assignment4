#!/bin/bash

# Configuration
CLUSTER="mapreduce-cluster"
REGION="us-west-2"

# Get networking info
VPC_ID=$(aws ec2 describe-vpcs --filters "Name=isDefault,Values=true" --query "Vpcs[0].VpcId" --output text --region $REGION)
SUBNET_ID=$(aws ec2 describe-subnets --filters "Name=vpc-id,Values=$VPC_ID" --query "Subnets[0].SubnetId" --output text --region $REGION)

# Create or get security group
SG_NAME="mapreduce-sg-$(date +%s)"
SG_ID=$(aws ec2 create-security-group \
    --group-name $SG_NAME \
    --description "MapReduce ECS tasks" \
    --vpc-id $VPC_ID \
    --query 'GroupId' \
    --output text \
    --region $REGION)

# Allow inbound on port 8080
aws ec2 authorize-security-group-ingress \
    --group-id $SG_ID \
    --protocol tcp \
    --port 8080 \
    --cidr 0.0.0.0/0 \
    --region $REGION

echo "Using Security Group: $SG_ID"

# Function to run task and store ARN
declare -a TASK_ARNS

run_task() {
    local task_def=$1
    local name=$2
    
    echo "Running $name..."
    TASK_ARN=$(aws ecs run-task \
        --cluster $CLUSTER \
        --task-definition $task_def \
        --launch-type FARGATE \
        --network-configuration "awsvpcConfiguration={subnets=[$SUBNET_ID],securityGroups=[$SG_ID],assignPublicIp=ENABLED}" \
        --query 'tasks[0].taskArn' \
        --output text \
        --region $REGION)
    
    if [ "$TASK_ARN" != "None" ]; then
        echo "$name started: $TASK_ARN"
        TASK_ARNS+=("$TASK_ARN")
    else
        echo "Failed to start $name"
    fi
}

# Run all tasks
run_task "mapreduce-splitter" "Splitter"
run_task "mapreduce-mapper" "Mapper 1"
run_task "mapreduce-mapper" "Mapper 2"
run_task "mapreduce-mapper" "Mapper 3"
run_task "mapreduce-reducer" "Reducer"

echo ""
echo "Waiting for tasks to reach RUNNING state (this takes 1-2 minutes)..."
sleep 90

# Get IPs
echo ""
echo "Task IPs:"
echo "========="

for TASK_ARN in "${TASK_ARNS[@]}"; do
    # Get task info
    TASK_INFO=$(aws ecs describe-tasks \
        --cluster $CLUSTER \
        --tasks $TASK_ARN \
        --region $REGION \
        --output json)
    
    # Extract task definition and ENI
    TASK_DEF=$(echo $TASK_INFO | jq -r '.tasks[0].taskDefinitionArn' | awk -F'/' '{print $2}' | awk -F':' '{print $1}')
    ENI_ID=$(echo $TASK_INFO | jq -r '.tasks[0].attachments[0].details[] | select(.name=="networkInterfaceId") | .value')
    
    # Get public IP
    if [ ! -z "$ENI_ID" ] && [ "$ENI_ID" != "null" ]; then
        PUBLIC_IP=$(aws ec2 describe-network-interfaces \
            --network-interface-ids $ENI_ID \
            --query 'NetworkInterfaces[0].Association.PublicIp' \
            --output text \
            --region $REGION)
        
        if [ ! -z "$PUBLIC_IP" ] && [ "$PUBLIC_IP" != "None" ]; then
            echo "$TASK_DEF: http://$PUBLIC_IP:8080"
        fi
    fi
done

echo ""
echo "Update your orchestrate.py with these IPs!"
echo ""
echo "To stop all tasks later, run:"
echo "aws ecs list-tasks --cluster $CLUSTER --region $REGION --query 'taskArns[]' --output text | xargs -I {} aws ecs stop-task --cluster $CLUSTER --task {} --region $REGION"