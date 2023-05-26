import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { decode, getKubeConfig } from '$lib/helpers.js';


export async function GET({ request }) {
    let id_token = request.headers.get('Authorization');
    let user_id = '';
    try {
        user_id = decode(id_token).preferred_username;
    } catch (err) {
        return new Response(JSON.stringify({ message: 'Invalid token' }), { status: 401, statusText: 'Invalid token' });
    }

    let response = new Response(JSON.stringify({ message: 'Internal server error' }), { status: 500 });
    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.AppsV1Api);
        await k8sApi
            .listNamespacedDeployment(user_id)
            .then((res) => {
                response = new Response(JSON.stringify(res.body), { status: 200, statusText: 'Success' });
            })
            .catch((err) => {
                response = new Response(JSON.stringify(err.body), { status: 401, statusText: 'Invalid token' });
            });
    }
    return response;
}