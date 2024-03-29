import { deleteReq, get, post, put, postFile } from "$lib/requests.js";

export async function getDeployments(token) {
    return get(token, '/api/kubelab/deploy');
}

export async function scaleDeployment(token, data, name) {
    return put(token, data, '/api/kubelab/deploy/scale/' + name);
}

export async function getConnectionString(token, data, name) {
    return post(token, data, '/api/kubelab/deploy/connection/' + name);
}

export async function postSSHToken(token, data) {
    return postFile(token, data, '/api/kubelab/ssh/upload/');
}

export async function getClasses(token) {
    return get(token, '/api/kubelab/classes/');
}

export async function getStudentsForClass(token, name) {
    return get(token, '/api/kubelab/classes/students/' + name);
}

export async function uploadFiles(token, data, className) {
    return postFile(token, data, '/api/kubelab/classes/files/upload/' + className);
}

export async function getFiles(token, name) {
    return get(token, '/api/kubelab/classes/files/get/' + name);
}

export async function deleteFiles(token, data, name) {
    return deleteReq(token, data, '/api/kubelab/classes/files/delete/' + name);
}
