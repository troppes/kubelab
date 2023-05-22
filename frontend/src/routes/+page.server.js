import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private'

export const load = async (event) => {

    let session = await event.locals.getSession();
    if (session && session.user) {
        let kc = getKubeConfig(session.user.id_token);
        let k8sApi = kc.makeApiClient(k8s.CoreV1Api);
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

}

// getApiClient returns a k8s api client
const getKubeConfig = (idToken) => {
    const cluster = {
        name: 'kubelab-cluster',
        server: env.KUBERNETES_SERVER_URL,
        caFile: env.KUBERNETES_CA_Path
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