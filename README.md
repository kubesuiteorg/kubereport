![KubeReport Logo](assets/kubereport.png)

# KubeReport 

KubeReport is a versatile CLI tool for generating both general and detailed reports on Kubernetes clusters. It enables users to easily send these reports via the command line and securely store them in the cloud. With KubeReport, managing cluster insights has never been more efficient and accessible.
   
## Demo

You can view the demo in the [test](test/README.md).

## Your Support Matters

KubeReport is a community-driven open-source project, maintained with dedication and effort. We are committed to keeping it free for everyone!

If KubeReport enhances your Kubernetes experience, please consider supporting us! Your [sponsorship](https://buymeacoffee.com/sachinran) will help us continue development and deliver valuable updates. Every contribution counts, and we greatly appreciate your support!

## Documentation

For detailed instructions on installation, usage, customization, and more, please refer to our [KubeReport documentation site](https://www.kubesuite.org/doc).

## Installation

**KubeReport is available for Linux, macOS, and Windows platforms. You can find the installation packages on the [release page](https://github.com/kubesuiteorg/kubereport/releases). After downloading the tarball, follow the steps below for your platform to install KubeReport.**

### Linux & macOS

1. **Extract the tarball**:

   ```bash
   tar -xvzf kubereport-<version>.tar.gz
   ```
   Replace <version> with the actual version number.

2. **Move the binary to a directory in your PATH (e.g., /usr/local/bin)**:

   ```bash
    sudo mv ./kubereport /usr/local/bin/
    ```

3. **Make the binary executable**:

   ```bash
    sudo chmod +x /usr/local/bin/kubereport
    ```

4. **Verify the installation by checking the version**:

   ```bash
    kubereport --version
    ```

### Windows

1. **Extract the tarball using a tool like 7-Zip or any other tar-compatible extraction tool.**

2. **Move the binary to a folder that is part of your PATH.**

3. **Verify the installation by checking the version**:

   ```bash
    kubereport --version
    ```

## The Command Line

To generate a report with KubeReport

```bash
# To generate report
kubereport

# To check the version
kubereport --version

# List all available CLI options
kubereport --help
```

### Command Options

| Flag              | Shorthand | Default Value | Description                                                                           |
|-------------------|-----------|---------------|---------------------------------------------------------------------------------------|
| `--version`       | `-v`      | `false`       | Displays the current version of KubeReport          .                                 |
| `--report`        | `-d`      | `general`     | Type of report to generate ( general [default], detailed ). |
| `--kubeconfig`    | `-k`      | `""`          | File path to the kubeconfig file used for accessing the Kubernetes cluster. Defaults to `$KUBECONFIG` or `~/.kube/config`. |
| `--schedule`      | `-t`      | `""`          | Cron expression to schedule the automatic generation and sending of reports (e.g., '* * * * *' for every minute). |
| `--recipient`     | `-r`      | `""`          | The email address where the generated report will be sent.                            |
| `--sender`        | `-s`      | `""`          | The email address used to send the report.                                            |
| `--password`      | `-p`      | `""`          | The SMTP password for the sender's email account.                                     |
| `--subject`       | `-j`      | `""`          | The subject line of the email containing the report.                                  |
| `--body`          | `-b`      | `""`          | The body content of the email that accompanies the report.                            |
| `--smtp-server`   | `-m`      | `""`          | Address of the SMTP server used to send the email (e.g., `smtp.gmail.com`).           |
| `--smtp-port`     | `-o`      | `587`         | Port number of the SMTP server (default is `587`).                                    |
| `--use-tls`       | `-u`      | `true`        | Indicates whether to use TLS (Transport Layer Security) for the SMTP connection (default is `true`). |

## To Deploy to Kubernetes Cluster

For the Helm chart required for KubeReport deployment, please refer to this [KubeReport Helm Chart Repository](https://github.com/kubesuiteorg/kubereport-helm-chart) for detailed installation instructions and configuration options.

## Contribution Guidelines

* **File an Issue or Suggestions**
* **Comment Your Code**
* **Attach Tested Output**

## Community & Support

Want to discuss KubeReport features with other users or show your support for this tool?

- **Invite**: Get your [KubeReport Slack Invite](https://join.slack.com/t/kubesuite/shared_invite/zt-2rh7j3whw-We_16ybaeK5tNjRXGenX_Q).
- **Slack Channel**: Join the conversation on [KubeReport Slack](https://kubesuite.slack.com/archives/C07PPLEUR7B).

You can also connect with us on [LinkedIn](https://www.linkedin.com/company/kubesuite/) to stay updated and engage with the community.
