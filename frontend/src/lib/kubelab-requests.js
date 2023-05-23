import { deleteReq, get, post, put } from "$lib/requests.js";

export async function getDeployments(token) {
    return get(token, '/api/kubelab/');
}

export async function scaleDeployment(token, name) {
    return post(token, { name: name }, '/api/kubelab/scale/');
}