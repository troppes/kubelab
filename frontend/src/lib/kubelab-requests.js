import { deleteReq, get, post, put, postFile } from "$lib/requests.js";

export async function getDeployments(token) {
    return get(token, '/api/kubelab/deploy');
}

export async function scaleDeployment(token, name) {
    return put(token, {}, '/api/kubelab/deploy/scale/' + name);
}

export async function getConnectionString(token, name) {
    return get(token, '/api/kubelab/deploy/connection/' + name);
}

export async function postSSHToken(token, data) {
    return postFile(token, data, '/api/kubelab/ssh/upload/');
}