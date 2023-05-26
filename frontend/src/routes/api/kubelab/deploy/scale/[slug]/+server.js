import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { decode, getKubeConfig } from '$lib/helpers.js';

export async function GET({ request, params }) {
    let id_token = request.headers.get('Authorization');
    let deployName = params.slug;
    let user_id = '';
    try {
        user_id = decode(id_token).preferred_username;
    } catch (err) {
        return new Response(JSON.stringify({ message: 'Invalid token' }), { status: 401, statusText: 'Error: Invalid token' });
    }
    let response = new Response(JSON.stringify({ message: 'Internal server error' }), { status: 500 });

    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.AppsV1Api);
        let deployDetail = await k8sApi.readNamespacedDeployment(deployName, user_id);

        // Switch off or on depending on state
        deployDetail.body.spec.replicas = deployDetail.body.spec.replicas === 0 ? 1 : 0;

        try {
            const res = await k8sApi.replaceNamespacedDeployment(deployName, user_id, deployDetail.body);
            response = new Response(JSON.stringify(null), { status: 200, statusText: 'Success' });

        } catch (err) {
            response = new Response(JSON.stringify(err.body), { status: err.statusCode, statusText: err.body.message });
        }

        return response;
    }
}