# GRACE-PaaS RDS [![GoDoc](https://godoc.org/github.com/GSA/grace-paas-rds?status.svg)](https://godoc.org/github.com/GSA/grace-paas-rds) [![Go Report Card](https://goreportcard.com/badge/gojp/goreportcard)](https://goreportcard.com/report/github.com/GSA/grace-paas-rds)[![CircleCI](https://circleci.com/gh/GSA/grace-paas-rds.svg?style=shield)](https://circleci.com/gh/GSA/grace-paas-rds)

Command to generate Terrafrom JSON Configuration Syntax to create an AWS
Relational Database Service (RDS) resource from a ServiceNow ticket (json)

## Usage

Download the binaries for your OS from the latest release artifacts. Unzip the
executable file in a directory in your `$PATH` and make executable. The command
takes two arguments: the input file containing the JSON from the ServiceNow
requested item, and the path/name of the output file for the JSON syntax
configuration file. The JSON syntax file should have a `.tf.json` extension.

For example:

```
$ grace-paas-rds RITM.json rds.tf.json
```

## Public domain

This project is in the worldwide [public domain](LICENSE.md). As stated in [CONTRIBUTING](CONTRIBUTING.md):

> This project is in the public domain within the United States, and copyright and related rights in the work worldwide are waived through the [CC0 1.0 Universal public domain dedication](https://creativecommons.org/publicdomain/zero/1.0/).
>
> All contributions to this project will be released under the CC0 dedication. By submitting a pull request, you are agreeing to comply with this waiver of copyright interest.
