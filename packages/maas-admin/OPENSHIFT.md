# Deploying MaaS Admin on OpenShift

This guide explains how to deploy the application on OpenShift using the `Dockerfile.openshift` container file.

## Prerequisites

- Access to an OpenShift cluster and the `oc` CLI
- The source code repository is accessible (public or with proper credentials)

## Steps

### 1. Create the OpenShift App

You can create a new app from your local directory or from a Git repository. Here is an example using the local directory:

```sh
oc new-app .
```

Or, using a Git repository:

```sh
oc new-app https://github.com/<your-org>/maas-admin.git
```

### 2. Patch the BuildConfig to Use `Dockerfile.openshift`

By default, OpenShift uses `Dockerfile` as the build file. To use `Dockerfile.openshift`, patch the BuildConfig after creation:

```sh
oc patch buildconfig maas-admin --type=merge -p '{"spec":{"strategy":{"dockerStrategy":{"dockerfilePath":"Dockerfile.openshift"}}}}'
```

Replace `maas-admin` with the name of your app (e.g., `maas-admin`).

### 3. Start a New Build

After patching, trigger a new build to use the updated Dockerfile:

```sh
oc start-build maas-admin
```

### 4. Monitor the Build

You can follow the build logs with:

```sh
oc logs -f buildconfig/maas-admin
```

### 5. Expose the Service (Optional)

To make your app accessible externally:

```sh
oc create route edge --service=maas-admin
```

## Notes

- Ensure any required environment variables (such as `MAAS_URL`) are set in your deployment configuration.
- You can view and edit environment variables with:
  ```sh
  oc set env deployment/maas-admin MAAS_URL=http://maas-service:8080
  ```
- For development/testing, you can enable mock clients:
  ```sh
  oc set env deployment/maas-admin MOCK_MAAS_CLIENT=true
  ```

---

For more details, see the main `README.md`.
