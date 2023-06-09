import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { decode, getKubeConfig } from '$lib/helpers.js';

export async function GET({ request, params }) {
    let id_token = request.headers.get('Authorization');
    let svcName = params.slug;
    let token = null;
    try {
        token = decode(id_token);
    } catch (err) {
        return json({ message: 'Invalid token' }, { status: 401, statusText: 'Invalid token' });
    }

    let response = json({ message: 'Internal server error' }, { status: 500, statusText: 'Internal Server Error' });
    if (token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.CoreV1Api);
        try {
            const svcDetail = await k8sApi.readNamespacedService(svcName, token.user_id);

            const connectionString =
                'ssh -p' + svcDetail.body.spec.ports[0].nodePort + ' ' +
                token.preferred_username + '@' + env.LOADBALANCER_IP;

            response = json(connectionString, { status: 200, statusText: 'Success' });
        } catch (err) {
            response = json({ message: err.body.message }, { status: err.statusCode, statusText: err.body.message });
        }

        return response;
    }
}