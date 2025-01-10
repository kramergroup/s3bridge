# S3Bridge deployment on kube@iscc

This is the deployment configuration of s3bridges on the kubernetes cluster at ISCC. It
uses [pulumi](https://www.pulumi.com) as the engine to control state in kubernetes.

## Prerequirements

- `kubecfg` must be working on the computer. Pulumi gets its kubernetes credentials from there.
- a valid Vault token must be available in the environment.

## Obtaining and setting a Vault token

Pulumi pulls secrets from the groups VAULT instance. You need to have access to it on the groups VPN and point pulumi to 
the right address:

```bash
esc env set s3bridge/dev pulumiConfig.vault:address <Vault URL>
```

Because pulumi stores access token (albeit encrypted), it is best practice to generate a short-lived and restrictive token. To create one
`POST` the following using a sufficiently permissive token:

```bash
curl -X POST -H "X-Vault-Token: <my secret token>" -d '{"policies": [ "read kv store"], "ttl": "24h" }' https://10.147.17.10:8200/v1/auth/token/create | jq .auth.client_token
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