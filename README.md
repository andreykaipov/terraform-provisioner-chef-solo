# terraform-provisioner-chef-solo

This is a Terraform provisioner plugin for Chef Solo. It lets us
provision instances with Chef Solo in our Terraform scripts:

```hcl
resource "aws_instance" "web" {
  # ...

  provisioner "chef-solo" {
    cookbook_paths  = ["cookbooks"]
    run_list        = ["book::recipe"]
    json            = <<-EOF
      {
        "a": "b",
        "c": "d"
      }
    EOF
  }
}
```

It's inspired by the similar [Packer provisioner for
Chef Solo](https://www.packer.io/docs/provisioners/chef-solo.html), so if you're
familiar with that one, then you'll be familiar with this one!

## Usage

To use this provisioner, first download the zipped binary for your system from
the [releases](https://github.com/andreykaipov/terraform-provisioner-chef-solo/releases)
page and unzip it. Alternatively, you can build it yourself (see below).

Terraform searches for plugins within the same directory as itself, so
you'll have to move this binary into that directory. If you're using an older
version of Terraform, you'll also have to rename the plugin to not include the
version information at the end of the name.

Documentation for the plugin is [here](DOCUMENTATION.md).

## Building

However you put Go projects into your GOPATH (whether it's with `go get` or
just manually cloning and moving it), do that for this project.

Since this is a plugin for Terraform, you need to have the Terraform source in
your GOPATH too. If you don't, run a `go get -v ./...` from the root of this repo.

Here's an example:
```
$ go get github.com/andreykaipov/terraform-provisioner-chef-solo
$ cd $GOPATH/src/github.com/andreykaipov/terraform-provisioner-chef-solo
$ make build
```

You should now have the plugin in the `bin/` directory.
