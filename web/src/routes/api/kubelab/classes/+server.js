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

        await k8sApi.listClusterCustomObject('kubelab.kubelab.local', 'v1', 'classrooms')
            .then((res) => {
                const classrooms = res.body.items.filter((classroom) =>
                    JSON.parse(classroom.metadata.annotations['kubectl.kubernetes.io/last-applied-configuration']).spec.teacher.spec.id === user_id
                ).map((classroom) => {
                    classroom.metadata.annotations = JSON.parse(classroom.metadata.annotations['kubectl.kubernetes.io/last-applied-configuration']);
                    return classroom;
                });
                response = json(classrooms, { status: 200, statusText: 'Success' });
            })
            .catch((err) => {
                response = json({ message: err.body.message }, { status: err.statusCode, statusText: err.body.message });
            });
    }
    return response;
}