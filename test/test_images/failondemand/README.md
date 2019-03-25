# Fail-on-demand test image

The image contains a simple Go webserver, `failondemand.go`, that will by
default, listen on port `8080` and expose a service at `/`.

When called, the server simulate a crashing image. It is useful for testing the
liveliness probes.

## Trying out

To run the image as a Service outside of the test suite:

`ko apply -f service.yaml`

## Building

For details about building and adding new images, see the
[section about test images](/test/README.md#test-images).