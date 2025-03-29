#!/bin/bash

# This script is useful for enabling the required services in GCP before running `terraform apply` the first time

gcloud services enable --project=$1 \
    cloudresourcemanager.googleapis.com \
    container.googleapis.com \
    compute.googleapis.com \
    servicenetworking.googleapis.com