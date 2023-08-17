import * as k8s from '@kubernetes/client-node';
import { env } from '$env/dynamic/private';
import { json } from '@sveltejs/kit';
import { decode, getKubeConfig } from '$lib/helpers.js';
import fs from "fs";
import path from "path";

export async function POST({ request, fetch, params }) {
    let id_token = request.headers.get('Authorization');
    const formData = await request.formData();
    const className = params.slug;
    let token = null;
    try {
        token = decode(id_token);
    } catch (err) {
        return json({ message: 'Invalid token' }, { status: 401, statusText: 'Invalid token' });
    }

    let response = json({ message: 'Internal server error' }, { status: 500, statusText: 'Internal Server Error' });

    if (id_token) {
        let kc = getKubeConfig(id_token, env.KUBERNETES_SERVER_URL, env.KUBERNETES_CA_Path);
        let k8sApi = kc.makeApiClient(k8s.CoreV1Api);
        try {
            // if this requests fails, we know the token is invalid or has been tampered with
            await k8sApi.listNamespacedPod(className);

            const folderPath = "/classes/" + className;

            for (const key of formData.keys()) {
                const filePath = path.join(folderPath, key);
                const stream = await formData.get(key).stream();
                const buffers = [];
                for await (const data of stream) {
                    buffers.push(data);
                }
                fs.writeFileSync(filePath, Buffer.concat(buffers), 'binary');
            }
            response = json({}, { status: 200, statusText: 'Success' });
        } catch (err) {
            console.log(err)
            response = json({ message: err }, { status: 500, statusText: "Failed to create file!" });
        }
        return response;
    }
}