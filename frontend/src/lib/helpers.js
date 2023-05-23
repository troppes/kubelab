
import * as k8s from '@kubernetes/client-node';

export const decode = function (token) {
    return JSON.parse(Buffer.from(token.split('.')[1], 'base64').toString())
}

export const getKubeConfig = (idToken, server, caFile) => {
    const cluster = {
        name: 'kubelab-cluster',
        server: server,
        caFile: caFile
    };

    const user = {
        name: 'user',
        token: idToken,
    };

    const context = {
        name: 'context',
        user: user.name,
        cluster: cluster.name
    };

    const kc = new k8s.KubeConfig();

    kc.loadFromOptions({
        clusters: [cluster],
        users: [user],
        contexts: [context],
        currentContext: context.name
    });

    return kc;
}
