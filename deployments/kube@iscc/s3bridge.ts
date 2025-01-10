import * as pulumi from "@pulumi/pulumi";
import * as k8s from "@pulumi/kubernetes";

type S3BridgeArgs = {
  appLabels?: any
  namespace: string
  externalURL?: {
    host: string,
    path?: string,
  },
  allowedOrigins?: string[]
  backend: {
    endpoint: string,
    bucket: string,
    s3_access_key: string,
    s3_secret_key: string,
  }
}

type S3BridgeOutputs = {
  ip: pulumi.Output<string>,
  deployment: pulumi.Output<string>,
  service?: pulumi.Output<string>,
  ingress?: pulumi.Output<string>,
}

export class S3Bridge extends pulumi.ComponentResource {

  constructor(name: string, args: S3BridgeArgs, opts?: pulumi.ComponentResourceOptions) {
    super("kramergroup:s3Bridge", name, args, opts);

    const appLabels = args.appLabels

    const deployment = new k8s.apps.v1.Deployment(name, {
      metadata: {
          namespace: args.namespace,
          name: name
      },
      spec: {
          selector: { matchLabels: appLabels },
          replicas: 1,
          template: {
              metadata: { labels: appLabels },
              spec: { containers: [{ 
                  name: "s3bridge", 
                  image: "kramergroup/s3bridge",
                  ports: [
                      { name: "http", containerPort: 80 }
                  ],
                  env: [
                      { name: "ENDPOINT", value: args.backend.endpoint },
                      { name: "BUCKET", value: args.backend.bucket },
                      { name: "AWS_ACCESS_KEY_ID", value: args.backend.s3_access_key},
                      { name: "AWS_SECRET_ACCESS_KEY", value: args.backend.s3_secret_key},
                  ]
              }] }
          }
      }
    }, { parent: this });

    const service = new k8s.core.v1.Service(name, {
      metadata: { 
        labels: appLabels,
        name: name,
        namespace: args.namespace,
      },
      spec: {
        type: "ClusterIP",
        ports: [{ port: 80, targetPort: 80, protocol: "TCP", name: "http" }],
        selector: appLabels
      }
    }, { parent: this });

    let outputs : S3BridgeOutputs = {
      ip: service.spec.clusterIP,
      deployment: deployment.metadata.name,
    }

    if (args.externalURL ) {

      const middlewares = []

      if ( args.externalURL.path) {

        middlewares.push( new k8s.apiextensions.CustomResource(`${name}-strip-prefix`, {
          apiVersion: "traefik.io/v1alpha1",
          kind: "Middleware",
          metadata: {
            labels: appLabels,
            namespace: args.namespace
          },
          spec: {
            stripPrefix: {
              prefixes: [ args.externalURL.path ],
              forceSlash: false
            }
          }
        }, { parent: this }))

        middlewares.push(new k8s.apiextensions.CustomResource(`${name}-cors`, {
          apiVersion: "traefik.io/v1alpha1",
          kind: "Middleware",
          metadata: {
            labels: appLabels,
            namespace: args.namespace
          },
          spec: {
            headers: {
              accessControlAllowMethods: [ "GET" ],
              accessControlAllowOriginList: args.allowedOrigins ? args.allowedOrigins : [ "*" ],
            }
          }
        }, { parent: this }))
      }


      const ingress = new k8s.apiextensions.CustomResource(name, {
        apiVersion: "traefik.io/v1alpha1",
        kind: "IngressRoute",
        metadata: { 
            labels: appLabels,
            namespace: args.namespace
        },
        spec: {
            entryPoints: [ "websecure" ],
            routes: [{
                kind: "Rule",
                match: `Host(\`${args.externalURL.host}\`)` + (args.externalURL.path ? ` && PathPrefix(\`${args.externalURL.path}\`)` : ''),
                services: [{
                    kind: "Service", name: service.metadata.name, port: 80
                }],
                middlewares: middlewares.map( m => { return { name: m.metadata.name } } ),
            }],
            tls: { certResolver: "prod" },
        }
      }, { parent: this });

      outputs = {...outputs, ingress: ingress.metadata.name, service: service.metadata.name }
    }

    this.registerOutputs(outputs);

  }

}