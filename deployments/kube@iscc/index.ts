import * as k8s from "@pulumi/kubernetes";
import * as vault from "@pulumi/vault";
import * as pulumi from "@pulumi/pulumi"
import { S3Bridge } from './s3bridge'

export = async () => {

    const appLabels = { app: "s3bridge" };

    const namespaceName = "s3bridge-app";

    const cephEndpoint = "http://hsuper-ceph.hsu-hh.de:8100"

    const cephCredentials = await vault.generic.getSecret({
        path: "kv/kube@iscc/s3bridge-ceph-credentials"
    })

    const backends = [
        {
            name: "compdes",
            bucket: "teaching-compdes-video",
            externalHostname: "assets.kramer.science",
            path: "/compdes/video",
            allowedOrigins: [ "https://compdes.hsu-hh.info" ]
        }
    ]

    const ns = new k8s.core.v1.Namespace(namespaceName, {
        metadata: { name: namespaceName }
    })

    const bridges = backends.map( backend => {

        return new S3Bridge(`s3bridge-${backend.name}`,{
            namespace: namespaceName,
            appLabels: {...appLabels, module: backend.name },
            backend: {
                endpoint: cephEndpoint,
                bucket: backend.bucket,
                s3_access_key: cephCredentials.data["s3_access_key"],
                s3_secret_key: cephCredentials.data["s3_secret_key"],
            },
            externalURL: {
                host: backend.externalHostname,
                path: backend.path,
            },
            allowedOrigins: backend.allowedOrigins
        })

    })

}


