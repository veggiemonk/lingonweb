# Lingon Web

Lingon Web is a web-based interface for [Lingon](https://github.com/volvo-cars/lingon). 

[Lingon](https://github.com/volvo-cars/lingon) is a library and command line tool to write HCL (<a href="https://www.terraform.io/" rel="nofollow">Terraform</a>)
and <a href="https://kubernetes.io" rel="nofollow">kubernetes</a> manifest (YAML) in Go.


This web app is an example of how to use the library to convert kubernetes manifests to Go code.

See <a href="https://github.com/volvo-cars/lingon/blob/main/docs/rationale.md">Rationale</a> for why we built this.

Lingon is not a platform, it is a library meant to be consumed in a Go application that platform engineers write to manage their platforms.
It is a tool to build and automate the creation and the management of platforms regardless of the target infrastructure and services.

Who is this for? 

Lingon is aimed at people who need to automate the lifecycle of their cloud infrastructure
and have suffered the pain of configuration languages and complexity of gluing tools together with more tools.


All the <a href="https://github.com/volvo-cars/lingon/blob/main/docs">Examples</a> are in the <a href="https://github.com/volvo-cars/lingon/blob/main/docs">documentation</a>.


A big example is <a href="https://github.com/volvo-cars/lingon/blob/main/docs/platypus">Platypus</a> which shows how
the <a href="https://github.com/volvo-cars/lingon/blob/main/docs/kubernetes">kubernetes</a>
and <a href="https://github.com/volvo-cars/lingon/blob/main/docs/terraform">terraform</a> libraries can be used together.


The output format is called <a href="https://pkg.go.dev/github.com/rogpeppe/go-internal@v1.10.0/txtar">txtar, short for text archive</a>.
It is also used by the Go playground. See the example: https://go.dev/play/p/3ThdpZyPj-b

```go 
-- other.go --
package main

func hello() string {
	return "Hello, 世界"
}
-- main.go --
// You can edit this code!
// Click here and start typing.
package main

import "fmt"

func main() {
	fmt.Println(hello())
}

```