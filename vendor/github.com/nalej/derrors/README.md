# derrors - Extended Errors

This repository contains the definition of an extended error structure for Go applications.

## General overview

The main purpose of this repository is to improve error reporting facilitating the communication of error states to the
users and allowing deeper reporting of the errors for the developers at the same time.

The Error interfaces defines a set of basic methods that makes a Error compatible with the GolangError but
provides extra functions to track the error origin.

## Building and testing

| Action  | Command |
| ------------- | ------------- |
| Build  | `go build`  |
| Install | `go install` |
| Test  | `go test`  |

## Building and testing with Bazel (deprecated)

To update the files, run:

```
'bazel run //:gazelle
```

To build the project, execute:

```
bazel build ...
```

To pass the tests,

```
bazel test ...
```

## How to use it

Defining a new error automatically extracts the StackTrace

```
return derrors.NewEntityError(descriptor, errors.NetworkDoesNotExists, err)
```

Will print the following message when calling `String()`.

```
[Operation] network does not exists
```

And a detailed one when calling `DebugReport()`.

```
[Operation] network does not exists
Parameters:
P0: []interface {}{"n1b59e008-a9f2-4a25-866a-ace0cabc38b2asdf"}

StackTrace:
<stack trace from the caller>
```
