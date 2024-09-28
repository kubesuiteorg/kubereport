# Setting Up Minikube for KubeReport

This demo will guide you through the process of setting up Minikube and configuring it to generate reports using the `KubeReport` CLI tool.

## Prerequisites

Before starting, ensure you have the following installed:

- **Docker**: Ensure Docker is installed and running on your machine.
- **Minikube**: Install Minikube and verify that it is correctly set up.

## Cluster Setup

1. **Start Minikube**:
    ```bash
    minikube start
    ```

2. **Deploy the Metrics Server**:

    Apply the Metrics Server deployment using the following command: 
    ```bash
    kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
    ```

3. **Configure the Metrics Server**:

    To ensure the Metrics Server works properly with Minikube, you'll need to modify the deployment to allow insecure TLS connections.

    Edit the Metrics Server deployment:
    ```bash
    kubectl edit deployment metrics-server -n kube-system
    ```

    In the editor, add the following under `args`:
    ```yaml
    args:
    - /metrics-server
    - --kubelet-insecure-tls
    ```

4. **Verify the Metrics Server**:

    After making the changes, ensure that the Metrics Server is up and running:
    ```bash
    kubectl get deployment metrics-server -n kube-system
    ```

    You should see the `metrics-server` deployment listed with the desired number of pods running.

## Generating Reports with KubeReport

1. **Generate a General Report**:
    ```bash
    kubereport
    ```
    OR
    ```bash
    kubereport --report=general
    ```
    Sample Output: [General Report](test/output/General_Report.png)

2. **Generate a Detailed Report**:
    ```bash
    kubereport --report=detailed
    ```
    Sample Output: [Detailed Report](test/output/Detailed_Report.png)

3. **Generate Report for Target Cluster**:
    ```bash
    kubereport --kubeconfig ~/.kube/config
    ```

4. **Schedule Reports**:
    ```bash
    kubereport --schedule "* * * * *"
    ```

5. **Send Reports via Email**:
    Use the following command to send reports through email:
    ```bash
    kubereport --recipient recipient@example.com --sender sender@example.com --password xxxxxxxx --subject "Mail Subject" --body "Mail Body" --smtp-server "smtp.gmail.com" --smtp-port "587" --use-tls true
    ```

    - **Gmail App Password**: Generate an app password from your Google account [here](https://myaccount.google.com/apppasswords).
    - **Outlook App Password**: Generate an app password from your Outlook account [here](https://account.live.com/proofs/AppPassword).

Now you're all set to use `KubeReport` to generate and manage your Kubernetes cluster reports!