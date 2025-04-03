# GCP IP List

[![godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/mark-adams/gcp-ip-list) [![license](http://img.shields.io/badge/license-MIT-red.svg?style=flat)](https://raw.githubusercontent.com/mark-adams/gcp-ip-list/main/LICENSE) [![Build Status](https://github.com/mark-adams/gcp-ip-list/actions/workflows/test.yml/badge.svg)](https://github.com/mark-adams/gcp-ip-list/actions/workflows/test.yml) 


`gcp-ip-list` is a CLI tool (and library) written in Go to simplify the process of retrieving IP addresses from infrastructure hosted on Google Cloud Platform (GCP).

Most enumeration tooling today uses the normal CRUD REST APIs provided by Google to retrieve GCP assets and their IP address information. This is less than ideal because it typically involves interacting with several different Google APIs and puts additional load on the very same APIs that are used as the control plane for GCP customers. In addition, it is quite slow especially if you have a large number of projects.

This tool takes a different approach and queries information about assets from Google's [Cloud Asset Inventory API](https://cloud.google.com/asset-inventory/docs/overview) instead. This allows us to use a single API to pull down all the data about assets that could potentially have public IP addresses assigned to them which allows us to download data for organizations of any size much more efficiently.

# Installation

<table>
    <tr>
        <td>Homebrew (macOS or Linux)</td>
        <td>
            <code>brew tap mark-adams/gcp-ip-list && brew install gcp-ip-list</code>
        </td>
    </tr>
</table>

Pre-built binaries are also avalable from the [Releases page](https://github.com/mark-adams/gcp-ip-list/releases)

If your system has a [supported version of Go](https://go.dev/dl/), you can build from source.

```
go install github.com/mark-adams/gcp-ip-list/cmd/gcp-ip-list@latest
```

# Running the tool

This application authenticates with GCP using [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials). Some examples of how you might authneticate include:

- Run the application on a GCP resource (VM, Cloud Function, etc.) with an attached service account
- Run the application on your workstation after using `gcloud auth application-default login` to use your user account's credentials
- Run the application on your workstation using a service account's credentials by running `gcloud auth activate-service-account`

For more information on authenticating, see the [Application Default Credentials](https://cloud.google.com/docs/authentication/application-default-credentials) documentation.

Since this application uses the Cloud Asset Inventory APIs, your user account / service account will need to have the Cloud Asset Viewer (`roles/cloudasset.viewer`) IAM role assigned for the targeted scope's (i.e. organization, folder, or project) IAM policy.

## Usage
```
$ gcp-ip-list -h       
Usage of gcp-ip-list:
  -format string
        The output format (csv, json, table, list) (default "table")
  -private
        Include private IPs only
  -public
        Include public IPs only
  -scope string
        The scope (organization, folder, or project) to search (i.e. projects/abc-123 or organizations/123456)
  -version
        Display the current version
```

### Use as a library
Core functionality of the CLI is exposed via Go APIs as well in the `github.com/mark-adams/gcp-ip-list/pkg/go` package via the `GetAllAddressesFromAssetInventory()` and `GetAddressesFromAssetInventory()` functions in case you want to incoporate this functionality into your own application.

## Examples

### Table output
```
$ gcp-ip-list --scope=projects/sample-project -public
+----------------+--------------+---------------------------------------+-------------------------------------------------------------------------------------------------------------------------+
|    ADDRESS     | ADDRESS TYPE |             RESOURCE TYPE             |                                                      RESOURCE NAME                                                      |
+----------------+--------------+---------------------------------------+-------------------------------------------------------------------------------------------------------------------------+
| 35.244.150.176 | public       | compute.googleapis.com/ForwardingRule | //compute.googleapis.com/projects/sample-project/global/forwardingRules/ip-list-test-forwarding-rule-external        |
| 34.54.75.78    | public       | compute.googleapis.com/ForwardingRule | //compute.googleapis.com/projects/sample-project/global/forwardingRules/ip-list-test-forwarding-rule-external-static |
| 34.83.163.216  | public       | compute.googleapis.com/Instance       | //compute.googleapis.com/projects/sample-project/zones/us-west1-a/instances/ip-list-test-vm                          |
| 34.105.8.244   | public       | compute.googleapis.com/Router         | //compute.googleapis.com/projects/sample-project/regions/us-west1/routers/ip-list-test-router                        |
| 34.19.43.198   | public       | container.googleapis.com/Cluster      | //container.googleapis.com/projects/sample-project/locations/us-west1/clusters/ip-list-test-cluster                  |
| 34.127.47.18   | public       | sqladmin.googleapis.com/Instance      | //cloudsql.googleapis.com/projects/sample-project/instances/ip-list-test-db                                          |
+----------------+--------------+---------------------------------------+-------------------------------------------------------------------------------------------------------------------------+
```

### List output

```
$ gcp-ip-list --scope=projects/sample-project -public -format=list
35.244.150.176
34.54.75.78
34.83.163.216
34.105.8.244
34.19.43.198
34.127.47.18
```

This mode is handy for piping to your favorite port scanning tool like `nmap` or `naabu`:
```
gcp-ip-list --scope=projects/sample-project -public -format=list | nmap -iL -
```

### CSV & JSON output

You can get the same output as the default table format but in CSV or JSON as well:

```
gcp-ip-list --scope=projects/sample-project -public -format=csv
```

```
gcp-ip-list --scope=projects/sample-project -public -format=json
```

# Contributing
See our [Contribution guidelines](CONTRIBUTING.md)

## Terraform resources
The `terraform` directory contains sample resources that are handy when doing local development on `gcp-ip-list`.
If you add support for a new resource type, please add the appropriate Terraform resources in the same PR.

# Releases
New releases can be found on the [Releases](page).

## Verifying signatures
Binaries built by this project are signed using Sigstore.

To verify the signature for a given binary, you can use [cosign](https://github.com/sigstore/cosign):

```
$ cosign verify-blob gcp-ip-list_Darwin_x86_64/gcp-ip-list \                                                 
       --bundle gcp-ip-list_Darwin_x86_64.cosign.bundle \
       --certificate-oidc-issuer=https://token.actions.githubusercontent.com \
       --certificate-identity=https://github.com/mark-adams/gcp-ip-list/.github/workflows/release.yml@refs/tags/<version>
Verified OK
```

# Troubleshooting

## Could not find default credentials

> error getting public addresses: error setting up client: credentials: could not find default credentials. See https://cloud.google.com/docs/authentication/external/set-up-adc for more information

This means that you're likely running the tool locally from your workstation without having application default credentials set up. You can follow the link in the message or run `gcloud auth application-default login` to authenticate with GCP and obtain the proper credentials.

## Cloud Asset API has not been used in project X before

This tool depends on the Cloud Asset Inventory API being enabled. Luckily, the error message points you in the right direction. Look for "Enable it by visiting https://..." in the error message and visit that page to enable the API.
