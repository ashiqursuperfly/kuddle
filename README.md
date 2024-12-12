# kuddle

[![Go Report Card](https://goreportcard.com/badge/github.com/ashiqursuperfly/kuddle)](https://goreportcard.com/report/github.com/ashiqursuperfly/kuddle)

Simple CLI that extends the functionality of kubectl logs to display logs from multiple pods matching a filter.

## Installation
Go binaries are automatically built with each release by [GoReleaser](https://github.com/goreleaser/goreleaser). These can be accessed on the GitHub [releases page](https://github.com/ashiqursuperfly/kuddle/releases) for this project.

### Direct download
You can directly download the executable binary from the github [releases](https://github.com/ashiqursuperfly/kuddle/releases)

```bash
os=linux
arc=x86_64
version=v0.0.1
wget "https://github.com/ashiqursuperfly/kuddle/releases/download/${version}/kuddle_${version}_${os}_${arc}.tar.gz"
tar -xvzf kuddle_${version}_${os}_${arc}.tar.gz
mv kuddle /usr/local/bin/kubectl-kuddle
```

### Krew
This project is not yet accepted in the Krew Index. Once accepted, this can be installed with [Krew](https://github.com/GoogleContainerTools/krew):
```
kubectl krew install kuddle
```
Meanwhile, you can download the krew manifest and install it directly from the manifest:
```yaml
apiVersion: krew.googlecontainertools.github.com/v1alpha2
kind: Plugin
metadata:
  name: kuddle
spec:
  version: v0.0.1
  homepage: https://github.com/ashiqursuperfly/kuddle
  shortDescription: Extends the functionality of kubectl logs to display logs from multiple pods.
  description: |
    Simple CLI that extends the functionality of kubectl logs to display logs from multiple pods matching a filter.
  platforms:
  - selector:
      matchLabels:
        os: darwin
        arch: amd64
    bin: kuddle
    files:
    - from: "*"
      to: "."
    uri: https://github.com/ashiqursuperfly/kuddle/releases/download/v0.0.1/kuddle_v0.0.1_darwin_x86_64.tar.gz
    sha256: 6ab0ad4264b25265c2c19b1ac4fa9b795b79330a0c63e3d5bcebf089772ecb0f
  - selector:
      matchLabels:
        os: darwin
        arch: arm64
    bin: kuddle
    files:
    - from: "*"
      to: "."
    uri: https://github.com/ashiqursuperfly/kuddle/releases/download/v0.0.1/kuddle_v0.0.1_darwin_arm64.tar.gz
    sha256: 4b9fe9ff3356f3234d2aa1a01e01f94061c067d0b2aca2e6838cf805caa0e952
  - selector:
      matchLabels:
        os: linux
        arch: amd64
    bin: kuddle
    files:
    - from: "*"
      to: "."
    uri: https://github.com/ashiqursuperfly/kuddle/releases/download/v0.0.1/kuddle_v0.0.1_linux_x86_64.tar.gz
    sha256: 2edf949f2b5929db8e14c96953250eeac2beaf54db36e00b41c612711713c622
  - selector:
      matchLabels:
        os: linux
        arch: arm64
    bin: kuddle
    files:
    - from: "*"
      to: "."
    uri: https://github.com/ashiqursuperfly/kuddle/releases/download/v0.0.1/kuddle_v0.0.1_linux_arm64.tar.gz
    sha256: 4c5539512294c0b997f8cd6cc013becb46462fec7fe496903348c04c8c54e6c8
  - selector:
      matchLabels:
        os: windows
        arch: amd64
    bin: kuddle.exe
    files:
    - from: "*"
      to: "."
    uri: https://github.com/ashiqursuperfly/kuddle/releases/download/v0.0.1/kuddle_v0.0.1_windows_x86_64.zip
    sha256: 70b0466ffa2a54f44f705e253aee084c0421a20c2b65c7b8c9f44f01a9eaeae4
```
Save the above manifest into `kuddle.yaml`. Then run,
```
kubectl krew install --manifest=kuddle.yaml
```
## Usage
```
❯ kubectl kuddle --help

Usage:
kuddle [options] --filter <regex> [additional kubectl logs flags]
0.0.1
Options:
  --filter <regex>              Regex to filter pod names (mandatory). Must be matchable using go regexp: regex.MatchString
  -n, --namespace <namespace>   Namespace to query pods from (default: "default")
  --extraArgs                   flags to be passed into kubectl logs command
  --help                        Show this usage information%
```

- Each log line follows the following format: `<pod-name> ]- <log-line>`
- Log lines coming from each unique pod will have it's own unique color for better visibility.

![kubectl-kuddle-example](images/image.png)

## Contributing
Always open to new features. Feel free to raise a PR from your fork!

## License
Apache License 2.0