import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { decode, getKubeConfig } from '$lib/helpers.js';
import fs from "fs";
import path from "path";

export async function POST({ request }) {
    let id_token = request.headers.get('Authorization');
    const formData = await request.formData();

    let token = null;
    try {
        token = decode(id_token);
    } catch (err) {
        return new Response(JSON.stringify({ message: 'Invalid token' }), { status: 401, statusText: 'Error: Invalid token' });
    }

    let response = new Response(JSON.stringify({ message: 'Internal server error' }), { status: 500 });

    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.CoreV1Api);
        try {
            // if this requests fails, we know the token is invalid or has been tampered with
            await k8sApi.listNamespacedPod(token.user_id);

            const folderPath = "/students/" + token.user_id + "/.kubelab/";
            const filePath = path.join(folderPath, "kubelab_key");
            const data = await formData.get('file').text();

            // create folder if not container was never created
            fs.mkdirSync(folderPath, { recursive: true });
            // write file
            fs.writeFileSync(filePath, data);

            response = new Response(JSON.stringify({ status: "ok" }), { status: 200, statusText: 'Success' });
        } catch (err) {
            console.log(err);
            response = new Response(JSON.stringify(err), { status: "Failed to create File!", statusText: err });
        }

        return response;
    }
}