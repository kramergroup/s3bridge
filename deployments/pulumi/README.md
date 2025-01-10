# S3Bridge deployment using pulumi

This is the deployment configuration of s3bridges on a kubernetes cluster using Pulumi. [Pulumi](https://www.pulumi.com)  is 
a powerfull IaC engine.

## Prerequirements

- `kubecfg` must be working on the computer. Pulumi gets its kubernetes credentials from there.
- a valid Vault token must be available in the environment.

## Using esc environments

Pulumi is configured to access config and secrets from ESC. You need to initialse this the first time

```bash
esc env init s3bridge/dev
```

## Obtaining and setting a Vault token

Pulumi pulls secrets from a Vault instance. You need to point pulumi to
the right address:

```bash
esc env set s3bridge/dev pulumiConfig.vault:address <Vault URL>
```

Because pulumi stores access token (albeit encrypted), it is best practice to generate a short-lived and restrictive token. To create one `POST` the following using a sufficiently permissive token:

```bash
curl -X POST -H "X-Vault-Token: <my secret token>" -d '{"policies": [ "read kv store"], "ttl": "24h" }' <Vault URL>/v1/auth/token/create | jq .auth.client_token
```

The returned token can then be stored in the environment with

```bash
esc env set s3bridge/dev pulumiConfig.vault:token --secret <new token>
```

## Update deployment

Use

```bash
pulumi up
```

to push an update of the desired state to kubernetes.