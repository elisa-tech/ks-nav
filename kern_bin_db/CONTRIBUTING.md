# Contributing to kern_bin_db

:+1::tada: First off, thanks for taking the time to 
contribute! :tada::+1:

The following is a set of guidelines for contributing to kern_bin_db.
The kern_bin_db utility is hosted in the 
[kern_bin_db repository](https://github.com/alessandrocarminati/kern_bin_db-) 
on GitHub. 
These are mostly guidelines, not rules. 
Use your best judgment, and feel free to propose changes to 
the code in a pull request.

# Bugs and new features

I intend to keep everything easy and track features and bugs 
by using the github infrastructure.
If you have issues or you want to purpose a new feature, feel 
free to just open issue or a pull request.

# Opening a PR
￼
To ensure that the code base stays clean and consistent, we are adhering to the following
rules:

* Follow Go conventions where applicable (e.g. naming using camelCase or PascalCase)
* Use `go fmt` to format the code 
* To sanity-check your code before submitting a PR, use [golangci-lint](https://golangci-lint.run/)
```
golangci-lint run
```
The provided `.golangci.yaml` file describes the linting rules that are used in this codebase. If you have
a good reason to skip a rule, you can add a `//nolint:rule` comment to the line that you want to skip.