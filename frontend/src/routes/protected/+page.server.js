import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private'

export const load = async (event) => {

    let session = await event.locals.getSession();

    const cluster = {
        name: 'kubelab-cluster',
        server: env.KUBERNETES_SERVER_URL,
        caFile: env.KUBERNETES_CA_Path
    };

    const user = {
        name: 'user',
        token: session.user.id_token,
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

    const k8sApi = kc.makeApiClient(k8s.CoreV1Api);
    let answer = '';
    await k8sApi
        .listNamespacedPod('default')
        .then((res) => {
            answer = res.body;
        })
        .catch((err) => {
            answer = err.body;
        });
    return {
        apiAnswer: JSON.stringify(answer)
    }
}