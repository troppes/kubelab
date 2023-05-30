import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { decode, getKubeConfig } from '$lib/helpers.js';

export async function PUT({ request, params }) {
    let id_token = request.headers.get('Authorization');
    let deployName = params.slug;
    let user_id = '';
    try {
        user_id = decode(id_token).user_id;
    } catch (err) {
        return new Response(JSON.stringify({ message: 'Invalid token' }), { status: 401, statusText: 'Error: Invalid token' });
    }
    let response = new Response(JSON.stringify({ message: 'Internal server error' }), { status: 500 });

    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.AppsV1Api);
        let allDeploys = await k8sApi.listNamespacedDeployment(user_id);

        // shut all down except for the one we want to use
        for (const deploy of allDeploys.body.items) {
            if (deploy.metadata.name === deployName) {
                // Switch off or on depending on state
                deploy.spec.replicas = deploy.spec.replicas === 0 ? 1 : 0;
            } else {
                deploy.spec.replicas = 0;
            }
            try {
                const res = await k8sApi.replaceNamespacedDeployment(deploy.metadata.name, user_id, deploy);
                response = new Response(JSON.stringify(null), { status: 200, statusText: 'Success' });
            } catch (err) {
                response = new Response(JSON.stringify(err.body), { status: err.statusCode, statusText: err.body.message });
            }
        }

        return response;
    }
}