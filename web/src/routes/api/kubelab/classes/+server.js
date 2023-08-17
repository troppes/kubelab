import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { decode, getKubeConfig } from '$lib/helpers.js';


export async function GET({ request }) {
    let id_token = request.headers.get('Authorization');
    let user_id = '';
    try {
        user_id = decode(id_token).user_id;
    } catch (err) {
        return json({ message: 'Invalid token' }, { status: 401, statusText: 'Invalid token' });
    }

    let response = json({ message: 'Internal server error' }, { status: 500, statusText: 'Internal Server Error' });
    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);
        await k8sApi.listClusterCustomObject('kubelab.kubelab.local', 'v1', 'classrooms', null, null, null, null, `teacher=${user_id}`)
            .then((res) => {
                response = json(res.body.items, { status: 200, statusText: 'Success' });
            })
            .catch((err) => {
                response = json({ message: err.body.message }, { status: err.statusCode, statusText: err.body.message });
            });
    }
    return response;
}