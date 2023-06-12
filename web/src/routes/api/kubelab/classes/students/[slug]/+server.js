import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { decode, getKubeConfig } from '$lib/helpers.js';

export async function GET({ request, params }) {
    let id_token = request.headers.get('Authorization');
    let className = params.slug;
    let user_id = '';
    try {
        user_id = decode(id_token).user_id;
    } catch (err) {
        return json({ message: 'Invalid token' }, { status: 401, statusText: 'Invalid token' });
    }
    let response = json({ message: 'Internal server error' }, { status: 500, statusText: 'Internal Server Error' });

    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.AppsV1Api);
        await k8sApi.listDeploymentForAllNamespaces(undefined, undefined, undefined, `class=` + className)
            .then((res) => {
                response = json(res.body, { status: 200, statusText: 'Success' });
            }).catch((err) => {
                response = json({ message: err.body.message }, { status: err.statusCode, statusText: err.body.message });
            });

        return response;
    }
}