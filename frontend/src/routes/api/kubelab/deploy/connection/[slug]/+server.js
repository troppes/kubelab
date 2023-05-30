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
        return new Response(JSON.stringify({ message: 'Invalid token' }), { status: 401, statusText: 'Error: Invalid token' });
    }

    let response = new Response(JSON.stringify({ message: 'Internal server error' }), { status: 500 });

    if (token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.CoreV1Api);
        try {
            const svcDetail = await k8sApi.readNamespacedService(svcName, token.user_id);

            const connectionString =
                'ssh -o "UserKnownHostsFile=/dev/null" -o "StrictHostKeyChecking=no" ' +
                '-p' + svcDetail.body.spec.ports[0].nodePort + ' ' +
                token.preferred_username + '@' + env.LOADBALANCER_IP;


            response = new Response(JSON.stringify(connectionString), { status: 200, statusText: 'Success' });
        } catch (err) {
            response = new Response(JSON.stringify(err.body), { status: err.statusCode, statusText: err.body.message });
        }

        return response;
    }
}