# bima

Over the years, container images have become the industry standard for software delivery. The widespread adoption of containerization has led to the development of numerous tools and container registries that facilitate various software development and delivery operations. Although the primary purpose of container images is to spawn container instances, there are certain software fields that could benefit from the container image paradigm even if the delivered software is not intended to be run as a container instance.

For example, developers can package their unikernel binary as a container image that requires a special runtime to spawn a unikernel VM. Similarly, an IoT device firmware and/or configuration options could be packaged as a container image. The OCI model provides a filesystem for delivering artifacts and a config.json with custom annotations to deliver configuration.

Inspired by this idea, we are thrilled to introduce `bima` ("βήμα" in Greek means step), the first step in exploring this novel approach to software delivery for non-container deployments. Currently, `bima` creates unikernel "container" images that can be utilized with our custom unikernel runtime `urunc`. Our team is tirelessly working to introduce support for IoT devices in the near future.
