# Lingon Web <!-- omit in toc -->

* [Intro](#intro)
* [Screenshot](#screenshot)
* [Who is this for?](#who-is-this-for)
* [Works with CRDs](#works-with-crds)
* [Output format is _txtar_](#output-format-is-txtar)

## Intro

Lingon Web is a web-based interface for [Lingon](https://github.com/volvo-cars/lingon).
[Lingon](https://github.com/volvo-cars/lingon) is a library and command line tool to write HCL (<a href="https://www.terraform.io/" rel="nofollow">Terraform</a>)
and <a href="https://kubernetes.io" rel="nofollow">kubernetes</a> manifest (YAML) in Go.

_This web app is an example of how to use the library to convert kubernetes manifests to Go code._

See <a href="https://github.com/volvo-cars/lingon/blob/main/docs/rationale.md">Rationale</a> for why we built this.

> Lingon is not a platform, it is a library meant to be consumed in a Go application that platform engineers write to manage their platforms.
> It is a tool to build and automate the creation and the management of platforms regardless of the target infrastructure and services.

## Screenshot

<p align="center">
<img src="https://user-images.githubusercontent.com/5487021/231055446-49ea2307-e16a-47ce-95e9-8ad11c24df73.png" alt="lingon webapp screenshot">
</p>


## Who is this for?

Lingon is aimed at people who need to automate the lifecycle of their cloud infrastructure
and have suffered the pain of configuration languages and complexity of gluing tools together with more tools.
We prefer to write Go code and use the same language for everything. 
It's not a popular opinion but it works for us.

All the <a href="https://github.com/volvo-cars/lingon/blob/main/docs">Examples</a> are in the <a href="https://github.com/volvo-cars/lingon/blob/main/docs">documentation</a>.

A big example is <a href="https://github.com/volvo-cars/lingon/blob/main/docs/platypus2">Platypus</a> which shows how
the <a href="https://github.com/volvo-cars/lingon/blob/main/docs/kubernetes">kubernetes</a>
and <a href="https://github.com/volvo-cars/lingon/blob/main/docs/terraform">terraform</a> libraries can be used together.

## Works with CRDs

Lingon can convert CRDs (Custom Resource Definitions) as well. 
Although it is not possible to convert all CRDs, as we would need to register them all.
See [serializer.go](./knowntypes/serializer.go) for an example of how to register the custom resource types.

> **Open an issue or a PR if you want to add more CRDs.** 🥲

## Output format is _txtar_

A _txtar_ archive is zero or more comment lines and then a sequence of file entries. 
Each file entry begins with a file marker line of the form "-- FILENAME --" and 
is followed by zero or more file content lines making up the file data. 
The comment or file content ends at the next file marker line. 
The file marker line must begin with the three-byte sequence "-- " and end with the three-byte sequence " --", 
but the enclosed file name can be surrounding by additional white space, all of which is stripped.
If the txtar file is missing a trailing newline on the final line, 
parsers should consider a final newline to be present anyway.
There are no possible syntax errors in a _txtar_ archive.

We are using [github.com/rogpeppe/go-internal](github.com/rogpeppe/go-internal) version.
<a href="https://pkg.go.dev/github.com/rogpeppe/go-internal@v1.10.0/txtar">see the doc for the API</a>.

It is also used by the [Go playground](https://go.dev/play). See the example: https://go.dev/play/p/3ThdpZyPj-b

Example:
```go 
-- other.go --
package main

func hello() string {
	return "Hello, 世界"
}

-- main.go --
package main

import "fmt"

func main() {
	fmt.Println(hello())
}

```

